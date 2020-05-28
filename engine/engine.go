package engine

import (
	"fmt"
	"github.com/baetyl/baetyl-core/sync"
	"os"
	"strconv"
	gosync "sync"
	"time"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
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
	syn   *sync.Sync
	Ami   AMI
	nod   *node.Node
	cfg   config.EngineConfig
	tomb  utils.Tomb
	sto   *bh.Store
	log   *log.Logger
	ns    string
	sysns string
}

func NewEngine(cfg config.EngineConfig, sto *bh.Store, nod *node.Node, syn *sync.Sync) (*Engine, error) {
	kube, err := GenAMI(cfg, sto)
	if err != nil {
		return nil, err
	}
	return &Engine{
		Ami:   kube,
		sto:   sto,
		syn:   syn,
		nod:   nod,
		cfg:   cfg,
		ns:    "baetyl-edge",
		sysns: "baetyl-edge-system",
		log:   log.With(log.Any("engine", cfg.Kind)),
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
	if err := e.reportAndApply(e.sysns, specv1.SYSTEM, delete); err != nil {
		return err
	}
	if err := e.reportAndApply(e.ns, specv1.USER, delete); err != nil {
		return err
	}
	return nil
}

func (e Engine) Collect(ns string, scope specv1.Scope) (specv1.Report, error) {
	var stats string
	if scope == specv1.SYSTEM {
		stats = "sysappstats"
	} else {
		stats = "appstats"
	}
	nodeInfo, err := e.Ami.CollectNodeInfo()
	if err != nil {
		e.log.Warn("failed to collect node info", log.Error(err))
	}
	nodeStats, err := e.Ami.CollectNodeStats()
	if err != nil {
		e.log.Warn("failed to collect node stats", log.Error(err))
	}
	appStats, err := e.Ami.CollectAppStatus(ns)
	if err != nil {
		e.log.Warn("failed to collect app stats", log.Error(err))
	}
	no, err := e.nod.Get()
	if err != nil {
		return nil, err
	}

	apps := make([]specv1.AppInfo, 0)
	for _, info := range appStats {
		app := specv1.AppInfo{
			Name:    info.Name,
			Version: info.Version,
		}
		apps = append(apps, app)
	}
	rapps := alignApps(apps, no.Desire.AppInfos(scope))
	// to report app status into local shadow, and return shadow delta
	r := specv1.Report{
		"time":      time.Now(),
		"node":      nodeInfo,
		"nodestats": nodeStats,
		stats:       appStats,
	}
	r.SetAppInfos(scope, rapps)
	return r, nil
}

func (e Engine) reportAndApply(ns string, scope specv1.Scope, delete bool) error {
	r, err := e.Collect(ns, scope)
	if err != nil {
		return err
	}
	rapps := r.AppInfos(scope)
	delta, err := e.nod.Report(r)
	if err != nil {
		return err
	}
	// if apps are updated, to apply new apps
	if delta != nil {
		dapps := delta.AppInfos(scope)
		if dapps != nil {
			if delete {
				if err := e.deleteApps(ns, dapps, rapps); err != nil {
					e.log.Error("failed to delete sys apps", log.Error(err))
				}
			}
			e.applyApps(ns, dapps, rapps)
			e.log.Info("to apply sys apps", log.Any("apps", dapps))
		}
	}
	return nil
}

func (e Engine) applyApps(ns string, desires []specv1.AppInfo, reports []specv1.AppInfo) {
	rs := make(map[string]specv1.AppInfo)
	for _, r := range reports {
		rs[r.Name] = r
	}
	var wg gosync.WaitGroup
	for _, appInfo := range desires {
		if _, ok := rs[appInfo.Name]; !ok {
			wg.Add(1)
			go func(wg *gosync.WaitGroup, info specv1.AppInfo) {
				if err := e.applyApp(ns, info); err != nil {
					e.log.Error("failed to apply application", log.Any("app info", appInfo), log.Error(err))
				}
				wg.Done()
			}(&wg, appInfo)
		}
	}
	wg.Wait()
}

func (e Engine) deleteApps(ns string, desire, report []specv1.AppInfo) error {
	del := make(map[string]specv1.AppInfo)
	for _, app := range report {
		del[app.Name] = app
	}
	for _, app := range desire {
		if d, ok := del[app.Name]; ok && d.Version == app.Version {
			delete(del, app.Name)
		}
	}
	for _, app := range del {
		key := makeKey(specv1.KindApplication, app.Name, app.Version)
		if key == "" {
			return fmt.Errorf("failed to get app name: (%s) version: (%s)", app.Name, app.Version)
		}
		var app specv1.Application
		if err := e.sto.Get(key, &app); err != nil {
			return err
		}
		if err := e.Ami.DeleteApplication(ns, app); err != nil {
			return err
		}
	}
	return nil
}

func (e Engine) applyApp(ns string, info specv1.AppInfo) error {
	if err := e.syn.SyncResource(info); err != nil {
		e.log.Error("failed to sync resource", log.Any("info", info), log.Error(err))
		return err
	}
	app, err := e.injectEnv(info)
	if err != nil {
		e.log.Error("failed to inject env to sys apps", log.Error(err))
		return err
	}
	cfgs := make(map[string]specv1.Configuration)
	secs := make(map[string]specv1.Secret)
	for _, v := range app.Volumes {
		if cfg := v.VolumeSource.Config; cfg != nil {
			key := makeKey(specv1.KindConfiguration, cfg.Name, cfg.Version)
			if key == "" {
				return fmt.Errorf("failed to get configuration name: (%s) version: (%s)", cfg.Name, cfg.Version)
			}
			var config specv1.Configuration
			if err := e.sto.Get(key, &config); err != nil {
				return err
			}
			cfgs[config.Name] = config
		} else if sec := v.VolumeSource.Secret; sec != nil {
			key := makeKey(specv1.KindSecret, sec.Name, sec.Version)
			if key == "" {
				return fmt.Errorf("failed to get secret name: (%s) version: (%s)", sec.Name, sec.Version)
			}
			var secret specv1.Secret
			if err := e.sto.Get(key, &secret); err != nil {
				return err
			}
			secs[secret.Name] = secret
		}
	}
	if err := e.Ami.ApplyConfigurations(ns, cfgs); err != nil {
		return err
	}
	if err := e.Ami.ApplySecrets(ns, secs); err != nil {
		return err
	}
	var imagePullSecs []string
	for n, sec := range secs {
		if isRegistrySecret(sec) {
			imagePullSecs = append(imagePullSecs, n)
		}
	}
	if err := e.Ami.ApplyApplication(ns, *app, imagePullSecs); err != nil {
		return err
	}
	return nil
}

func (e *Engine) injectEnv(info specv1.AppInfo) (*specv1.Application, error) {
	key := makeKey(specv1.KindApplication, info.Name, info.Version)
	app := new(specv1.Application)
	err := e.sto.Get(key, app)
	if err != nil {
		e.log.Error("failed to get resource from store", log.Any("key", key), log.Error(err))
		return nil, err
	}
	var services []specv1.Service
	for _, svc := range app.Services {
		env := []specv1.Environment{
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
	return app, nil
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

func makeKey(kind specv1.Kind, name, ver string) string {
	if name == "" || ver == "" {
		return ""
	}
	return string(kind) + "-" + name + "-" + ver
}

// ensuring apps have same order in report and desire list
func alignApps(reApps, deApps []specv1.AppInfo) []specv1.AppInfo {
	if len(reApps) == 0 || len(deApps) == 0 {
		return reApps
	}
	as := map[string]specv1.AppInfo{}
	for _, a := range reApps {
		as[a.Name] = a
	}
	var res []specv1.AppInfo
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

func isRegistrySecret(secret specv1.Secret) bool {
	registry, ok := secret.Labels[specv1.SecretLabel]
	return ok && registry == specv1.SecretRegistry
}
