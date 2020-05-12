package engine

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/baetyl/baetyl-core/ami"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/spec/crd"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl-go/utils"
	routing "github.com/qiangxue/fasthttp-routing"
	bh "github.com/timshannon/bolthold"
)

const (
	EnvKeyAppName     = "BAETYL_APP_NAME"
	EnvKeyNodeName    = "BAETYL_NODE_NAME"
	EnvKeyServiceName = "BAETYL_SERVICE_NAME"
)

type Engine struct {
	Ami  ami.AMI
	nod  *node.Node
	cfg  config.EngineConfig
	tomb utils.Tomb
	sto  *bh.Store
	log  *log.Logger
	ns   string
}

func NewEngine(cfg config.EngineConfig, sto *bh.Store, nod *node.Node) (*Engine, error) {
	kube, err := ami.GenAMI(cfg, sto)
	if err != nil {
		return nil, err
	}
	return &Engine{
		Ami: kube,
		sto: sto,
		nod: nod,
		cfg: cfg,
		ns:  "baetyl-edge",
		log: log.With(log.Any("engine", cfg.Kind)),
	}, nil
}

func (e *Engine) Start() {
	e.tomb.Go(e.reporting)
}

func (e *Engine) ReportAndDesire() error {
	return e.reportAndDesireAsync(false)
}

func (e *Engine) GetServiceLog(ctx *routing.Context) error {
	service := ctx.Param("service")
	tailLines := string(ctx.QueryArgs().Peek("tailLines"))
	sinceSeconds := string(ctx.QueryArgs().Peek("sinceSeconds"))

	tail, since, err := e.validParam(tailLines, sinceSeconds)
	if err != nil {
		http.RespondMsg(ctx, 400, "RequestParamInvalid", err.Error())
		return nil
	}

	reader, err := e.Ami.FetchLog(e.ns, service, tail, since)
	if err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	http.RespondStream(ctx, 200, reader, -1)
	return nil
}

func (e *Engine) reporting() error {
	e.log.Info("engine starts to report")
	defer e.log.Info("engine has stopped reporting")

	t := time.NewTicker(e.cfg.Report.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			err := e.reportAndDesireAsync(true)
			if err != nil {
				e.log.Error("failed to report local shadow", log.Error(err))
			} else {
				e.log.Debug("engine reports local shadow")
			}
		case <-e.tomb.Dying():
			return nil
		}
	}
}

func (e *Engine) reportAndDesireAsync(delete bool) error {
	// to collect app status
	info, err := e.Ami.Collect(e.ns)
	if err != nil {
		return err
	}
	if len(info) == 0 {
		return errors.New("no status collected")
	}
	no, err := e.nod.Get()
	if err != nil {
		return err
	}
	if info["apps"] != nil {
		info["apps"] = alignApps(info.AppInfos(), no.Desire.AppInfos())
	}
	if info["sysapps"] != nil {
		info["sysapps"] = alignApps(info.SysAppInfos(), no.Desire.SysAppInfos())
	}

	// to report app status into local shadow, and return shadow delta
	delta, err := e.nod.Report(info)
	if err != nil {
		return err
	}
	// if apps are updated, to apply new apps
	if delta == nil {
		return nil
	}
	apps := delta.AppInfos()
	if apps != nil {
		err = e.injectEnv(apps)
		if err != nil {
			e.log.Error("failed to inject env to apps", log.Error(err))
			return err
		}
		err = e.Ami.Apply(e.ns, apps, "!"+ami.LabelSystemApp, delete)
		if err != nil {
			e.log.Error("failed to apply apps", log.Error(err))
			return err
		}
		e.log.Info("to apply apps", log.Any("apps", apps))
	}
	sysApps := delta.SysAppInfos()
	if sysApps != nil {
		err = e.injectEnv(sysApps)
		if err != nil {
			e.log.Error("failed to inject env to sys apps", log.Error(err))
			return err
		}
		err = e.Ami.Apply(e.ns, sysApps, ami.LabelSystemApp, delete)
		if err != nil {
			e.log.Error("failed to apply sys apps", log.Error(err))
			return err
		}
		e.log.Info("to apply sys apps", log.Any("apps", sysApps))
	}
	return nil
}

func (e *Engine) injectEnv(appInfos []v1.AppInfo) error {
	for _, info := range appInfos {
		key := makeKey(crd.KindApplication, info.Name, info.Version)
		var app crd.Application
		err := e.sto.Get(key, &app)
		if err != nil {
			e.log.Error("failed to get key from store", log.Any("key", key), log.Error(err))
			return err
		}
		var services []crd.Service
		for _, svc := range app.Services {
			env := []crd.Environment{
				{
					Name:  EnvKeyAppName,
					Value: app.Name,
				},
				{
					Name:  EnvKeyServiceName,
					Value: svc.Name,
				},
				{
					Name:  EnvKeyNodeName,
					Value: os.Getenv(EnvKeyNodeName),
				},
			}
			svc.Env = append(svc.Env, env...)
			services = append(services, svc)
		}
		app.Services = services
		err = e.sto.Upsert(key, app)
		if err != nil {
			e.log.Error("failed to get key from store", log.Any("key", key), log.Error(err))
			return err
		}
	}
	return nil
}

func (e *Engine) validParam(tailLines, sinceSeconds string) (itailLines, isinceSeconds int64, err error) {
	if tailLines != "" {
		if itailLines, err = strconv.ParseInt(tailLines, 10, 64); err != nil {
			return
		}
		if itailLines < 0 {
			err = fmt.Errorf("The request parameter is invalid.(%s)", "tailLines is invalid")
			return
		}
	}
	if sinceSeconds != "" {
		if isinceSeconds, err = strconv.ParseInt(sinceSeconds, 10, 64); err != nil {
			return
		}
		if isinceSeconds < 0 {
			err = fmt.Errorf("The request parameter is invalid.(%s)", "sinceSeconds is invalid")
			return
		}
	}
	return
}

func (e *Engine) Close() {
	e.tomb.Kill(nil)
	e.tomb.Wait()
}

func makeKey(kind crd.Kind, name, ver string) string {
	if name == "" || ver == "" {
		return ""
	}
	return string(kind) + "-" + name + "-" + ver
}

func alignApps(reApps, deApps []v1.AppInfo) []v1.AppInfo {
	if len(reApps) == 0 || len(deApps) == 0 {
		return reApps
	}
	as := map[string]v1.AppInfo{}
	for _, a := range reApps {
		as[a.Name] = a
	}
	var res []v1.AppInfo
	for _, a := range deApps {
		if r, ok := as[a.Name]; ok {
			res = append(res, r)
			delete(as, a.Name)
		}
	}
	for _, a := range as {
		res = append(res, a)
	}
	return res
}
