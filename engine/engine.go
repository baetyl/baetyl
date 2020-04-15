package engine

import (
	"errors"
	"github.com/baetyl/baetyl-go/spec/crd"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"math/rand"
	"os"
	"time"

	"github.com/baetyl/baetyl-core/ami"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
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
	e := &Engine{
		Ami: kube,
		sto: sto,
		nod: nod,
		cfg: cfg,
		ns:  "baetyl-edge",
		log: log.With(log.Any("engine", cfg.Kind)),
	}
	return e, nil
}

func (e *Engine) Start() {
	e.tomb.Go(e.reporting)
}

func (e *Engine) ReportAndDesire() error {
	return e.reportAndDesireAsync()
}

func (e *Engine) reporting() error {
	e.log.Info("engine starts to report")
	defer e.log.Info("engine has stopped reporting")

	t := time.NewTicker(e.cfg.Report.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			err := e.reportAndDesireAsync()
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

func (e *Engine) reportAndDesireAsync() error {
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
			return err
		}
		err = e.Ami.Apply(e.ns, apps, "!" + ami.LabelSystemApp)
		if err != nil {
			return err
		}
		e.log.Info("to apply apps", log.Any("apps", apps))
	}
	sysApps := delta.SysAppInfos()
	if sysApps != nil {
		err = e.injectEnv(sysApps)
		if err != nil {
			return err
		}
		err = e.Ami.Apply(e.ns, sysApps, ami.LabelSystemApp)
		if err != nil {
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
			return err
		}
		var services []crd.Service
		for _, svc := range app.Services {
			env := []crd.Environment {
				{
					Name:EnvKeyAppName,
					Value: app.Name,
				},
				{
					Name: EnvKeyServiceName,
					Value: svc.Name,
				},
				{
					Name: EnvKeyNodeName,
					Value: os.Getenv(EnvKeyNodeName),
				},
			}
			svc.Env = append(svc.Env, env...)
			services = append(services, svc)
		}
		app.Services = services
		err = e.sto.Upsert(key, app)
		if err != nil {
			return err
		}
	}
	return nil
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
