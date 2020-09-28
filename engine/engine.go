package engine

import (
	"crypto/md5"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	gosync "sync"
	"time"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/mitchellh/mapstructure"
	routing "github.com/qiangxue/fasthttp-routing"
	bh "github.com/timshannon/bolthold"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/helper"
	"github.com/baetyl/baetyl/v2/node"
	"github.com/baetyl/baetyl/v2/security"
	"github.com/baetyl/baetyl/v2/sync"
)

const (
	SystemCertVolumePrefix = "baetyl-cert-volume-"
	SystemCertSecretPrefix = "baetyl-cert-secret-"
	SystemCertPath         = "/var/lib/baetyl/system/certs"
	DefaultHostPathLib     = "/var/lib/baetyl"
)

//go:generate mockgen -destination=../mock/engine.go -package=mock -source=engine.go Engine

type Engine interface {
	Start()
	ReportAndDesire() error
	GetServiceLog(ctx *routing.Context) error
	Collect(ns string, isSys bool, desire specv1.Desire) specv1.Report
	Close()
}

// pipes: remote debugging of the routing table between the router and the channel. key={ns}_{name}_{container}
type engineImpl struct {
	mode           string
	hostHostPath   string
	objectHostPath string
	cfg            config.Config
	syn            sync.Sync
	ami            ami.AMI
	nod            *node.Node
	sto            *bh.Store
	log            *log.Logger
	sec            security.Security
	hp             helper.Helper
	chains         gosync.Map
	tomb           utils.Tomb
}

func NewEngine(cfg config.Config, sto *bh.Store, nod *node.Node, syn sync.Sync, helper helper.Helper) (Engine, error) {
	mode := context.RunMode()
	log.L().Info("app running mode", log.Any("mode", mode))

	var hostPathLib string
	if val := os.Getenv(context.KeyBaetylHostPathLib); val == "" {
		err := os.Setenv(context.KeyBaetylHostPathLib, DefaultHostPathLib)
		if err != nil {
			return nil, errors.Trace(err)
		}
		hostPathLib = DefaultHostPathLib
	} else {
		hostPathLib = val
	}
	am, err := ami.NewAMI(mode, cfg.AMI)
	if err != nil {
		return nil, errors.Trace(err)
	}
	sec, err := security.NewPKI(cfg.Security, sto)
	if err != nil {
		return nil, errors.Trace(err)
	}
	eng := &engineImpl{
		mode:           mode,
		hostHostPath:   filepath.Join(hostPathLib, "host"),
		objectHostPath: filepath.Join(hostPathLib, "object"),
		ami:            am,
		sto:            sto,
		syn:            syn,
		nod:            nod,
		cfg:            cfg,
		sec:            sec,
		hp:             helper,
		chains:         gosync.Map{},
		log:            log.With(),
	}
	return eng, nil
}

func (e *engineImpl) Start() {
	e.tomb.Go(e.reporting)
	e.hp.Subscribe(helper.TopicDownside, &handlerDownside{engineImpl: e})
}

func (e *engineImpl) ReportAndDesire() error {
	return errors.Trace(e.reportAndDesireAsync(false))
}

func (e *engineImpl) GetServiceLog(ctx *routing.Context) error {
	service := ctx.Param("service")
	isSys := string(ctx.QueryArgs().Peek("system"))
	tailLines := string(ctx.QueryArgs().Peek("tailLines"))
	sinceSeconds := string(ctx.QueryArgs().Peek("sinceSeconds"))

	tail, since, err := e.validParam(tailLines, sinceSeconds)
	if err != nil {
		http.RespondMsg(ctx, 400, "RequestParamInvalid", err.Error())
		return nil
	}
	ns := context.BaetylEdgeNamespace
	if isSys == "true" {
		ns = context.BaetylEdgeSystemNamespace
	}
	reader, err := e.ami.FetchLog(ns, service, tail, since)
	if err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	http.RespondStream(ctx, 200, reader, -1)
	return nil
}

func (e *engineImpl) reporting() error {
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

func (e *engineImpl) reportAndDesireAsync(delete bool) error {
	node, err := e.nod.Get()
	if err != nil {
		return errors.Trace(err)
	}
	if err := e.recycleIfNeed(node); err != nil {
		e.log.Error("failed to recycle", log.Error(err))
	}
	if err := e.reportAndApply(true, delete, node.Desire); err != nil {
		return errors.Trace(err)
	}
	if err := e.reportAndApply(false, delete, node.Desire); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (e *engineImpl) recycleIfNeed(node *specv1.Node) error {
	report := node.Report
	val, ok := report["nodestats"]
	if !ok {
		return errors.Errorf("node stats not exist in report data")
	}
	var nodeStats specv1.NodeStats
	if err := mapstructure.Decode(val, &nodeStats); err != nil {
		return errors.Trace(err)
	}
	if nodeStats.DiskPressure {
		return e.recycle()
	}
	return nil
}

func (e *engineImpl) Collect(ns string, isSys bool, desire specv1.Desire) specv1.Report {
	nodeInfo, err := e.ami.CollectNodeInfo()
	if err != nil {
		e.log.Warn("failed to collect node info", log.Error(err))
	}
	nodeStats, err := e.ami.CollectNodeStats()
	if err != nil {
		e.log.Warn("failed to collect node stats", log.Error(err))
	}
	appStats, err := e.ami.StatsApps(ns)
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

func (e *engineImpl) reportAndApply(isSys, delete bool, desire specv1.Desire) error {
	var ns string
	if isSys {
		ns = context.BaetylEdgeSystemNamespace
	} else {
		ns = context.BaetylEdgeNamespace
	}
	r := e.Collect(ns, isSys, desire)
	e.log.Debug("collect stats of node and apps", log.Any("report", r))

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
			if err := e.ami.DeleteApp(ns, n); err != nil {
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

func (e *engineImpl) reportAppStatsIfNeed(isSys bool, r specv1.Report, stats map[string]specv1.AppStats) error {
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

func (e *engineImpl) applyApps(ns string, infos map[string]specv1.AppInfo, stats map[string]specv1.AppStats) {
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

func (e *engineImpl) applyApp(ns string, info specv1.AppInfo) error {
	if err := e.syn.SyncResource(info); err != nil {
		e.log.Error("failed to sync resource", log.Any("info", info), log.Error(err))
		return errors.Trace(err)
	}
	key := makeKey(specv1.KindApplication, info.Name, info.Version)
	app := new(specv1.Application)
	err := e.sto.Get(key, app)
	if err != nil {
		return errors.Errorf("failed to get app name: (%s) version: (%s) with error: %s", app.Name, app.Version, err.Error())
	}
	cfgs := make(map[string]specv1.Configuration)
	secs := make(map[string]specv1.Secret)
	for _, v := range app.Volumes {
		if cfg := v.VolumeSource.Config; cfg != nil {
			key := makeKey(specv1.KindConfiguration, cfg.Name, cfg.Version)
			if key == "" {
				return errors.Errorf("failed to get config name: (%s) version: (%s)", cfg.Name, cfg.Version)
			}
			var config specv1.Configuration
			if err := e.sto.Get(key, &config); err != nil {
				return errors.Errorf("failed to get config name: (%s) version: (%s) with error: %s", cfg.Name, cfg.Version, err.Error())
			}
			cfgs[config.Name] = config
		} else if sec := v.VolumeSource.Secret; sec != nil {
			key := makeKey(specv1.KindSecret, sec.Name, sec.Version)
			if key == "" {
				return errors.Errorf("failed to get secret name: (%s) version: (%s)", sec.Name, sec.Version)
			}
			var secret specv1.Secret
			if err := e.sto.Get(key, &secret); err != nil {
				return errors.Errorf("failed to get secret name: (%s) version: (%s) with error: %s", sec.Name, sec.Version, err.Error())
			}
			secs[secret.Name] = secret
		}
	}
	if err := sync.PrepareApp(e.hostHostPath, e.objectHostPath, app, cfgs); err != nil {
		e.log.Error("failed to revise applications", log.Any("app", app), log.Error(err))
		return errors.Trace(err)
	}
	// inject system cert
	if e.sec != nil {
		if err := e.injectCert(app, secs); err != nil {
			return errors.Trace(err)
		}
	}
	// apply app
	return errors.Trace(e.ami.ApplyApp(ns, *app, cfgs, secs))
}

func (e *engineImpl) injectCert(app *specv1.Application, secs map[string]specv1.Secret) error {
	ca, err := e.sec.GetCA()
	if err != nil {
		return errors.Trace(err)
	}

	var services []specv1.Service
	for _, svc := range app.Services {
		ns := context.BaetylEdgeNamespace
		if app.System {
			ns = context.BaetylEdgeSystemNamespace
		}
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
			DNSNames: []string{
				fmt.Sprintf("%s.%s", svc.Name, ns),
				fmt.Sprintf("%s", svc.Name),
				"localhost",
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
			System: app.Namespace == context.BaetylEdgeSystemNamespace,
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

func (e *engineImpl) validParam(tailLines, sinceSeconds string) (itailLines, isinceSeconds int64, err error) {
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

func (e *engineImpl) Close() {
	e.tomb.Kill(nil)
	e.tomb.Wait()
	if e.hp != nil {
		e.hp.Unsubscribe(helper.TopicDownside)
	}
}
