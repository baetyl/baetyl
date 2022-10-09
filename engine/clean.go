package engine

import (
	"os"
	"path/filepath"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

func (e *engineImpl) recycle() error {
	e.log.Info("start recycling useless object storage space")
	nod, err := e.nod.Get()
	if err != nil {
		return errors.Trace(err)
	}
	rSysApps := nod.Report.AppInfos(true)
	rApps := nod.Report.AppInfos(false)
	rApps = append(rApps, rSysApps...)

	dSysApps := nod.Desire.AppInfos(true)
	dApps := nod.Desire.AppInfos(false)
	rApps = append(rApps, dApps...)
	rApps = append(rApps, dSysApps...)
	usedCfg := map[string]struct{}{}
	for _, info := range rApps {
		app := new(specv1.Application)
		err := e.sto.Get(makeKey(specv1.KindApplication, info.Name, info.Version), app)
		if err != nil {
			return errors.Trace(err)
		}
		for _, v := range app.Volumes {
			if cfg := v.Config; cfg != nil {
				usedCfg[makeKey(specv1.KindConfiguration, cfg.Name, cfg.Version)] = struct{}{}
			}
		}
	}
	del := make(map[string]specv1.Configuration)
	if err := e.sto.ForEach(nil, func(cfg *specv1.Configuration) error {
		if isObjectConfig(cfg) {
			key := makeKey(specv1.KindConfiguration, cfg.Name, cfg.Version)
			if _, ok := usedCfg[key]; !ok {
				del[key] = *cfg
			}
		}
		return nil
	}); err != nil {
		return errors.Trace(err)
	}
	for k, v := range del {
		if err := e.sto.Delete(k, specv1.Configuration{}); err != nil {
			e.log.Error("failed to delete configuration", log.Error(err))
		}
		dir := filepath.Join(e.cfg.Sync.Download.Path, v.Name)
		if err := os.RemoveAll(dir); err != nil {
			e.log.Error("failed to clean dir", log.Any("dir", dir))
		}
	}
	e.log.Info("complete recycling useless object storage space")
	return nil
}

func (e *engineImpl) cleanObjectStorage() (int, error) {
	node, err := e.nod.Get()
	if err != nil {
		return 0, errors.Trace(err)
	}
	objectCfgs := map[string]*specv1.Configuration{}
	err = e.sto.ForEach(nil, func(cfg *specv1.Configuration) error {
		if isObjectConfig(cfg) {
			key := makeKey(specv1.KindConfiguration, cfg.Name, cfg.Version)
			if key != "" {
				objectCfgs[key] = cfg
			}
		}
		return nil
	})
	if err != nil {
		return 0, errors.Trace(err)
	}

	var infos []specv1.AppInfo
	infos = append(infos, node.Report.AppInfos(false)...)
	infos = append(infos, node.Desire.AppInfos(false)...)
	infos = append(infos, node.Report.AppInfos(true)...)
	infos = append(infos, node.Desire.AppInfos(true)...)
	occupied := map[string]string{}
	for _, info := range infos {
		occupied[info.Name] = info.Version
	}
	obsoleteApps := make(map[string]*specv1.Application)
	occupiedApps := make(map[string]*specv1.Application)
	err = e.sto.ForEach(nil, func(app *specv1.Application) error {
		if ver, ok := occupied[app.Name]; ok && ver == app.Version {
			occupiedApps[app.Name] = app
			return nil
		}
		if prev, ok := obsoleteApps[app.Name]; !ok || app.UpdateTime.After(prev.UpdateTime) {
			obsoleteApps[app.Name] = app
		}
		return nil
	})
	finishedJobs := getFinishedJobs(occupiedApps, node)
	usedObjectCfgs := getUsedObjectCfgs(occupiedApps, finishedJobs)
	for name, ver := range usedObjectCfgs {
		key := makeKey(specv1.KindConfiguration, name, ver)
		if _, ok := objectCfgs[key]; ok && key != "" {
			delete(objectCfgs, key)
		}
	}

	dels := getDelObjectCfgs(occupiedApps, obsoleteApps, objectCfgs, finishedJobs)
	var subs []os.DirEntry
	for k, v := range dels {
		if err = e.sto.Delete(k, specv1.Configuration{}); err != nil {
			e.log.Error("failed to delete configuration", log.Error(err))
		}
		dir := filepath.Join(e.cfg.Sync.Download.Path, v.Name)
		subs, err = os.ReadDir(dir)
		if err != nil {
			e.log.Error("failed to read sub dirs", log.Error(err))
			continue
		}
		verDir := filepath.Join(dir, v.Version)
		if err = os.RemoveAll(verDir); err != nil {
			e.log.Error("failed to clean dir", log.Any("dir", verDir))
		}
		if len(subs) == 1 && subs[0].Name() == v.Version {
			if err = os.RemoveAll(dir); err != nil {
				e.log.Error("failed to clean dir", log.Any("dir", dir))
			}
		}
	}
	e.log.Info("complete cleaning useless object storage space")
	return len(dels), nil
}

func getDelObjectCfgs(occupied, obsolete map[string]*specv1.Application, objectCfgs map[string]*specv1.Configuration,
	finishedJobs map[string]struct{}) map[string]*specv1.Configuration {
	dels := map[string]*specv1.Configuration{}
	// get object config of obsolete app
	for _, app := range obsolete {
		refers := map[string]*specv1.ObjectReference{}
		for _, v := range app.Volumes {
			if v.Config != nil {
				refers[v.Name] = v.Config
			}
		}
		for _, svc := range app.Services {
			for _, vm := range svc.VolumeMounts {
				refer, ok := refers[vm.Name]
				if !ok || !vm.AutoClean {
					continue
				}
				key := makeKey(specv1.KindConfiguration, refer.Name, refer.Version)
				if cfg, ok := objectCfgs[key]; ok && key != "" {
					dels[key] = cfg
				}
			}
		}
	}
	// get object config of finished jobs
	for _, app := range occupied {
		refers := map[string]*specv1.ObjectReference{}
		for _, v := range app.Volumes {
			if v.Config != nil {
				refers[v.Name] = v.Config
			}
		}
		if _, ok := finishedJobs[app.Name]; !ok {
			continue
		}
		for _, svc := range app.Services {
			for _, vm := range svc.VolumeMounts {
				refer, ok := refers[vm.Name]
				if !ok || !vm.AutoClean {
					continue
				}
				key := makeKey(specv1.KindConfiguration, refer.Name, refer.Version)
				if cfg, ok := objectCfgs[key]; ok && key != "" {
					dels[key] = cfg
				}
			}
		}
	}

	return dels
}

func getUsedObjectCfgs(apps map[string]*specv1.Application, finishedJobs map[string]struct{}) map[string]string {
	used := map[string]string{}
	for _, app := range apps {
		refers := map[string]*specv1.ObjectReference{}
		for _, v := range app.Volumes {
			if v.Config != nil {
				refers[v.Name] = v.Config
			}
		}
		if app.Workload == specv1.WorkloadJob {
			if _, ok := finishedJobs[app.Name]; ok {
				continue
			}
		}
		for _, svc := range app.Services {
			for _, vm := range svc.VolumeMounts {
				cfg, ok := refers[vm.Name]
				if !ok {
					continue
				}
				used[cfg.Name] = cfg.Version
			}
		}
	}
	return used
}

func getFinishedJobs(apps map[string]*specv1.Application, node *specv1.Node) map[string]struct{} {
	jobSvcs := map[string]struct{}{}
	for _, app := range apps {
		if app.Workload == specv1.WorkloadJob {
			// jobs with same service name are not allowed on edge node,
			// thus it's reasonable to use service name as key
			jobSvcs[app.Name] = struct{}{}
		}
	}
	jobStatus := map[string]specv1.Status{}
	for _, stat := range node.Report.AppStats(false) {
		for _, ins := range stat.InstanceStats {
			_, ok := jobSvcs[ins.AppName]
			if !ok {
				continue
			}
			if status, ok := jobStatus[ins.AppName]; !ok || status == specv1.Succeeded {
				jobStatus[ins.AppName] = ins.Status
			}
		}
	}
	res := map[string]struct{}{}
	for name, status := range jobStatus {
		if status == specv1.Succeeded {
			res[name] = struct{}{}
		}
	}
	return res
}

func (e *engineImpl) cleaning() error {
	e.log.Info("engine starts to clean")
	defer e.log.Info("engine has stopped cleaning")

	t := time.NewTicker(e.cfg.Engine.Clean.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			n, err := e.cleanObjectStorage()
			if err != nil {
				e.log.Error("failed to clean object storage", log.Error(err))
			} else {
				e.log.Debug("engine clean object storage", log.Any("cleaned directory number", n))
			}
		case <-e.tomb.Dying():
			return nil
		}
	}
}
