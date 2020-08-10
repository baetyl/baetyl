package engine

import (
	"crypto/md5"
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	gosync "sync"
	"time"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/baetyl/baetyl/ami"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/node"
	"github.com/baetyl/baetyl/security"
	"github.com/baetyl/baetyl/sync"
	routing "github.com/qiangxue/fasthttp-routing"
	bh "github.com/timshannon/bolthold"
)

const (
	SystemCertVolumePrefix = "baetyl-cert-volume-"
	SystemCertSecretPrefix = "baetyl-cert-secret-"
	SystemCertPath         = "/var/lib/baetyl/system/certs"
)

type Engine struct {
	cfg   config.Config
	syn   sync.Sync
	ami   ami.AMI
	nod   *node.Node
	sto   *bh.Store
	log   *log.Logger
	ns    string
	sysns string
	sec   security.Security
	tomb  utils.Tomb
}

func NewEngine(cfg config.Config, sto *bh.Store, nod *node.Node, syn sync.Sync) (*Engine, error) {
	kube, err := ami.NewAMI(cfg.Engine.AmiConfig)
	if err != nil {
		return nil, errors.Trace(err)
	}
	sec, err := security.NewPKI(cfg.Security, sto)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &Engine{
		ami:   kube,
		sto:   sto,
		syn:   syn,
		nod:   nod,
		cfg:   cfg,
		sec:   sec,
		ns:    "baetyl-edge",
		sysns: "baetyl-edge-system",
		log:   log.With(log.Any("engine", cfg.Engine.Kind)),
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
	isSys := string(ctx.QueryArgs().Peek("system"))
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

	t := time.NewTicker(e.cfg.Engine.Report.Interval)
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

func (e *Engine) Collect(ns string, isSys bool, desire specv1.Desire) specv1.Report {
	nodeInfo, err := e.ami.CollectNodeInfo()
	if err != nil {
		e.log.Warn("failed to collect node info", log.Error(err))
	}
	nodeStats, err := e.ami.CollectNodeStats()
	if err != nil {
		e.log.Warn("failed to collect node stats", log.Error(err))
	}
	appStats, err := e.ami.CollectAppStats(ns)
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
	return r
}

func (e *Engine) reportAndApply(isSys, delete bool, desire specv1.Desire) error {
	var ns string
	if isSys {
		ns = e.sysns
	} else {
		ns = e.ns
	}
	r := e.Collect(ns, isSys, desire)
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
	stats := map[string]specv1.AppStats{}
	for _, s := range r.AppStats(isSys) {
		stats[s.Name] = s
	}
	appData, err := e.syn.SyncApps(dapps)
	if err != nil {
		return errors.Trace(err)
	}
	// will remove invalid app info in update
	checkService(dapps, appData, stats, update)
	checkPort(dapps, appData, stats, update)
	if err = e.reportAppStatsIfNeed(isSys, r, stats); err != nil {
		return errors.Trace(err)
	}
	if delete {
		for n := range del {
			if err := e.ami.DeleteApplication(ns, n); err != nil {
				e.log.Error("failed to delete applications", log.Any("system", isSys), log.Error(err))
				return errors.Trace(err)
			}
		}
	}
	e.applyApps(ns, update, stats)
	if err = e.reportAppStatsIfNeed(isSys, r, stats); err != nil {
		return errors.Trace(err)
	}
	e.log.Info("to apply applications", log.Any("system", isSys), log.Any("apps", dapps))
	return nil
}

func (e *Engine) reportAppStatsIfNeed(isSys bool, r specv1.Report, stats map[string]specv1.AppStats) error {
	if len(stats) == 0 {
		return nil
	}
	appStats := make([]specv1.AppStats, 0)
	for _, s := range stats {
		appStats = append(appStats, s)
	}
	r.SetAppStats(isSys, appStats)
	_, err := e.nod.Report(r)
	if err != nil {
		return err
	}
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

func (e *Engine) applyApps(ns string, infos map[string]specv1.AppInfo, stats map[string]specv1.AppStats) {
	var wg gosync.WaitGroup
	for _, info := range infos {
		wg.Add(1)
		go func(wg *gosync.WaitGroup, info specv1.AppInfo) {
			if err := e.applyApp(ns, info); err != nil {
				e.log.Error("failed to apply application", log.Any("info", info), log.Error(err))
				stat := stats[info.Name]
				stat.Cause += err.Error()
				stats[info.Name] = stat
			}
			wg.Done()
		}(&wg, info)
	}
	wg.Wait()
}

func (e *Engine) applyApp(ns string, info specv1.AppInfo) error {
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
	if err := e.reviseApp(app, cfgs); err != nil {
		e.log.Error("failed to revise applications", log.Any("app", app), log.Error(err))
		return errors.Trace(err)
	}
	// inject internal cert
	if e.sec != nil {
		if err := e.injectCert(app, secs); err != nil {
			return errors.Trace(err)
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
				Name:  context.EnvKeyAppName,
				Value: app.Name,
			},
			{
				Name:  context.EnvKeyServiceName,
				Value: svc.Name,
			},
			{
				Name:  context.EnvKeyAppVersion,
				Value: app.Version,
			},
			{
				Name:  context.EnvKeyNodeName,
				Value: os.Getenv(context.EnvKeyNodeName),
			},
			{
				Name:  context.EnvKeyCertPath,
				Value: SystemCertPath,
			},
		}
		svc.Env = append(svc.Env, env...)
		services = append(services, svc)
	}
	app.Services = services
	return app, nil
}

func (e *Engine) reviseApp(app *specv1.Application, cfgs map[string]specv1.Configuration) error {
	if app == nil {
		return nil
	}
	for i := range app.Volumes {
		if hostPath := app.Volumes[i].HostPath; hostPath != nil {
			if strings.HasPrefix(hostPath.Path, "/") {
				continue
			}
			fullPath := path.Join(appDataHostPath, path.Join("/", hostPath.Path))
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				return err
			}
			app.Volumes[i].HostPath = &specv1.HostPathVolumeSource{Path: fullPath}
		} else if config := app.Volumes[i].Config; config != nil {
			cfg, ok := cfgs[config.Name]
			if !ok {
				continue
			}
			for k := range cfg.Data {
				if !strings.HasPrefix(k, configKeyObject) {
					continue
				}
				if app.Volumes[i].HostPath == nil {
					app.Volumes[i].Config = nil
					app.Volumes[i].HostPath = &specv1.HostPathVolumeSource{
						Path: path.Join(e.cfg.Sync.DownloadPath, cfg.Name, cfg.Version),
					}
				}
			}
		}
	}
	return nil
}

func (e *Engine) injectCert(app *specv1.Application, secs map[string]specv1.Secret) error {
	ca, err := e.sec.GetCA()
	if err != nil {
		return errors.Trace(err)
	}

	var services []specv1.Service
	for _, svc := range app.Services {
		// generate cert
		commonName := fmt.Sprintf("%s.%s", app.Name, svc.Name)
		max := len(svc.Name)
		if max > 10 {
			max = 10
		}
		suffix := fmt.Sprintf("%x-%s", md5.Sum([]byte(commonName)), svc.Name[0:max])
		cert, err := e.sec.IssueCertificate(commonName, security.AltNames{
			IPs: []net.IP{
				net.IPv4(0, 0, 0, 0),
				net.IPv4(127, 0, 0, 1),
			},
			URIs: []*url.URL{
				{
					Scheme: "https",
					Host:   "localhost",
				},
				{
					Scheme: "https",
					Host:   svc.Name,
				},
				{
					Scheme: "https",
					Host:   fmt.Sprintf("%s.%s", svc.Name, e.sysns),
				},
			},
		})
		if err != nil {
			return errors.Trace(err)
		}
		secretName := SystemCertSecretPrefix + suffix
		if _, ok := secs[secretName]; ok {
			e.log.Warn("the secret will be overwritten for internal communication",
				log.Any("name", secretName))
		}

		secret := specv1.Secret{
			Name:      secretName,
			Namespace: app.Namespace,
			Labels: map[string]string{
				"baetyl-app-name": app.Name,
				"security-type":   "certificate",
			},
			Data: map[string][]byte{
				"crt.pem": cert.Crt,
				"key.pem": cert.Key,
				"ca.pem":  ca,
			},
			System: app.Namespace == e.sysns,
		}
		secs[secretName] = secret

		// generate volume mount
		volName := SystemCertVolumePrefix + suffix
		volMount := specv1.VolumeMount{
			Name:      volName,
			MountPath: SystemCertPath,
			ReadOnly:  true,
		}
		svc.VolumeMounts = append(svc.VolumeMounts, volMount)

		// generate volume
		vol := specv1.Volume{
			Name:         volName,
			VolumeSource: specv1.VolumeSource{Secret: &specv1.ObjectReference{Name: secret.Name}},
		}
		app.Volumes = append(app.Volumes, vol)

		services = append(services, svc)
	}
	app.Services = services
	return nil
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
