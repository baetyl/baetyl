package engine

import (
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"os"
	"path/filepath"
)

func (e *Engine) recycle() error {
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
		if isObjectMetaConfig(cfg) {
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
