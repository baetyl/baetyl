package engine

import (
	"os"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

const (
	GCInterval = time.Second * 60
)

func (e *Engine) gc() error {
	t := time.NewTicker(GCInterval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			recycle, err := e.ami.CheckRecycle()
			if err != nil {
				e.log.Error("failed to get recycle status", log.Error(err))
				continue
			}
			if recycle {
				if err := e.recycle(); err != nil {
					e.log.Error("failed to recycle", log.Error(err))
					continue
				}
			}
		case <-e.tomb.Dying():
			return nil
		}
	}
}

func (e *Engine) recycle() error {
	nod, err := e.nod.Get()
	if err != nil {
		return errors.Trace(err)
	}
	sysApps := nod.Report.AppInfos(true)
	apps := nod.Report.AppInfos(false)
	apps = append(apps, sysApps...)
	usedCfg := map[string]struct{}{}
	for _, app := range apps {
		key := makeKey(specv1.KindApplication, app.Name, app.Version)
		app := new(specv1.Application)
		err := e.sto.Get(key, app)
		if err != nil {
			return errors.Trace(err)
		}
		for _, v := range app.Volumes {
			if cfg := v.Config; cfg != nil {
				key = makeKey(specv1.KindConfiguration, cfg.Name, cfg.Version)
				if key == "" {
					return errors.Errorf("illegal configuration key [%s]", key)
				}
				usedCfg[key] = struct{}{}
			}
		}
	}
	var refresh bool
	// avoid remove configuration which is used currently
	if e.cache.Len() > CacheSize/2 {
		refresh = true
	}
	for key := range usedCfg {
		cfg := new(specv1.Configuration)
		err := e.sto.Get(key, cfg)
		if err != nil {
			return errors.Trace(err)
		}
		if !isObjectMetaConfig(cfg) {
			delete(usedCfg, key)
			continue
		}
		if refresh {
			_, ok := e.cache.Get(key)
			if !ok {
				return errors.Errorf("failed to get value from lru cache")
			}
		}
	}
	for {
		key, val, ok := e.cache.RemoveOldest()
		if !ok {
			return errors.Errorf("failed to get value from lru cache")
		}
		if _, ok := usedCfg[key.(string)]; ok {
			e.cache.Add(key, val)
		} else {
			if err := os.RemoveAll(val.(string)); err != nil {
				return errors.Errorf("failed to remove dir %v", val)
			}
			break
		}
	}
	return nil
}
