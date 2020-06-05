package engine

import (
	"os"
	"strconv"
	gosync "sync"
	"time"

	"github.com/baetyl/baetyl-go/errors"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/baetyl/baetyl/ami"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/node"
	"github.com/baetyl/baetyl/sync"
	routing "github.com/qiangxue/fasthttp-routing"
	bh "github.com/timshannon/bolthold"
)

const (
	EnvKeyAppName     = "BAETYL_APP_NAME"
	EnvKeyNodeName    = "BAETYL_NODE_NAME"
	EnvKeyServiceName = "BAETYL_SERVICE_NAME"
)

type Engine struct {
	syn   sync.Sync
	ami   ami.AMI
	nod   *node.Node
	cfg   config.EngineConfig
	tomb  utils.Tomb
	sto   *bh.Store
	log   *log.Logger
	ns    string
	sysns string
}

func NewEngine(cfg config.EngineConfig, sto *bh.Store, nod *node.Node, syn sync.Sync) (*Engine, error) {
	kube, err := ami.NewAMI(cfg.AmiConfig)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &Engine{
		ami:   kube,
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
	return errors.Trace(e.reportAndDesireAsync(false))
}

func (e *Engine) GetServiceLog(ctx *routing.Context) error {
	service := ctx.Param("service")
	isSys   := string(ctx.QueryArgs().Peek("system"))
	tailLines := string(ctx.QueryArgs().Peek("tailLines"))
	sinceSeconds := string(ctx.QueryArgs().Peek("sinceSeconds"))

	tail, since, err := e.validParam(tailLines, sinceSeconds)
	if err != nil {
		http.RespondMsg(ctx, 400, "RequestParamInvalid", err.Error())
		return nil
	}
	ns := e.ns
	if isSys == "true" {
		ns = e.sysns
	}
	reader, err := e.ami.FetchLog(ns, service, tail, since)
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
	node, err := e.nod.Get()
	if err != nil {
		return errors.Trace(err)
	}
	if err := e.reportAndApply(true, delete, node.Desire); err != nil {
		return errors.Trace(err)
	}
	if err := e.reportAndApply(false, delete, node.Desire); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (e Engine) Collect(ns string, isSys bool, desire specv1.Desire) (specv1.Report, error) {
	nodeInfo, err := e.ami.CollectNodeInfo()
	if err != nil {
		e.log.Warn("failed to collect node info", log.Error(err))
	}
	nodeStats, err := e.ami.CollectNodeStats()
	if err != nil {
		e.log.Warn("failed to collect node stats", log.Error(err))
	}
	appStats, err := e.ami.CollectAppStatus(ns)
	if err != nil {
		e.log.Warn("failed to collect app stats", log.Error(err))
	}
	apps := make([]specv1.AppInfo, 0)
	for _, info := range appStats {
		app := specv1.AppInfo{
			Name:    info.Name,
			Version: info.Version,
		}
		apps = append(apps, app)
	}
	if desire != nil {
		apps = alignApps(apps, desire.AppInfos(isSys))
	}
	r := specv1.Report{
		"time":      time.Now(),
		"node":      nodeInfo,
		"nodestats": nodeStats,
	}
	r.SetAppInfos(isSys, apps)
	r.SetAppStats(isSys, appStats)
	return r, nil
}

func (e Engine) reportAndApply(isSys, delete bool, desire specv1.Desire) error {
	var ns string
	if isSys {
		ns = e.sysns
	} else {
		ns = e.ns
	}
	r, err := e.Collect(ns, isSys, desire)
	if err != nil {
		return errors.Trace(err)
	}
	rapps := r.AppInfos(isSys)
	delta, err := e.nod.Report(r)
	if err != nil {
		return errors.Trace(err)
	}
	// if apps are updated, to apply new apps
	if delta == nil {
		return nil
	}
	dapps := delta.AppInfos(isSys)
	if dapps == nil {
		return nil
	}
	del, update := getDeleteAndUpdate(dapps, rapps)
	if delete {
		for n := range del {
			if err := e.ami.DeleteApplication(ns, n); err != nil {
				e.log.Error("failed to delete applications", log.Any("system", isSys), log.Error(err))
				return errors.Trace(err)
			}
		}
	}
	e.applyApps(ns, update)
	e.log.Info("to apply applications", log.Any("system", isSys), log.Any("apps", dapps))
	return nil
}

func getDeleteAndUpdate(desires, reports []specv1.AppInfo) (map[string]specv1.AppInfo, map[string]specv1.AppInfo) {
	del := make(map[string]specv1.AppInfo)
	update := make(map[string]specv1.AppInfo)
	for _, d := range desires {
		update[d.Name] = d
	}
	for _, r := range reports {
		del[r.Name] = r
		if app, ok := update[r.Name]; ok && app.Version == r.Version {
			delete(update, app.Name)
		}
	}
	for _, app := range desires {
		if _, ok := del[app.Name]; ok {
			delete(del, app.Name)
		}
	}
	return del, update
}

func (e Engine) applyApps(ns string, infos map[string]specv1.AppInfo) {
	var wg gosync.WaitGroup
	for _, info := range infos {
		wg.Add(1)
		go func(wg *gosync.WaitGroup, info specv1.AppInfo) {
			if err := e.applyApp(ns, info); err != nil {
				e.log.Error("failed to apply application", log.Any("info", info), log.Error(err))
			}
			wg.Done()
		}(&wg, info)
	}
	wg.Wait()
}

func (e Engine) applyApp(ns string, info specv1.AppInfo) error {
	if err := e.syn.SyncResource(info); err != nil {
		e.log.Error("failed to sync resource", log.Any("info", info), log.Error(err))
		return errors.Trace(err)
	}
	app, err := e.injectEnv(info)
	if err != nil {
		e.log.Error("failed to inject env to applications", log.Any("info", info), log.Error(err))
		return errors.Trace(err)
	}
	cfgs := make(map[string]specv1.Configuration)
	secs := make(map[string]specv1.Secret)
	for _, v := range app.Volumes {
		if cfg := v.VolumeSource.Config; cfg != nil {
			key := makeKey(specv1.KindConfiguration, cfg.Name, cfg.Version)
			if key == "" {
				return errors.Errorf("failed to get configuration name: (%s) version: (%s)", cfg.Name, cfg.Version)
			}
			var config specv1.Configuration
			if err := e.sto.Get(key, &config); err != nil {
				return errors.Trace(err)
			}
			cfgs[config.Name] = config
		} else if sec := v.VolumeSource.Secret; sec != nil {
			key := makeKey(specv1.KindSecret, sec.Name, sec.Version)
			if key == "" {
				return errors.Errorf("failed to get secret name: (%s) version: (%s)", sec.Name, sec.Version)
			}
			var secret specv1.Secret
			if err := e.sto.Get(key, &secret); err != nil {
				return errors.Trace(err)
			}
			secs[secret.Name] = secret
		}
	}
	if err := e.ami.ApplyConfigurations(ns, cfgs); err != nil {
		return errors.Trace(err)
	}
	if err := e.ami.ApplySecrets(ns, secs); err != nil {
		return errors.Trace(err)
	}
	var imagePullSecs []string
	for n, sec := range secs {
		if isRegistrySecret(sec) {
			imagePullSecs = append(imagePullSecs, n)
		}
	}
	if err := e.ami.ApplyApplication(ns, *app, imagePullSecs); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (e *Engine) injectEnv(info specv1.AppInfo) (*specv1.Application, error) {
	key := makeKey(specv1.KindApplication, info.Name, info.Version)
	app := new(specv1.Application)
	err := e.sto.Get(key, app)
	if err != nil {
		e.log.Error("failed to get resource from store", log.Any("key", key), log.Error(err))
		return nil, errors.Trace(err)
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
			err = errors.Errorf("The request parameter is invalid.(%s)", "tailLines is invalid")
			return
		}
	}
	if sinceSeconds != "" {
		if isinceSeconds, err = strconv.ParseInt(sinceSeconds, 10, 64); err != nil {
			return
		}
		if isinceSeconds < 0 {
			err = errors.Errorf("The request parameter is invalid.(%s)", "sinceSeconds is invalid")
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
