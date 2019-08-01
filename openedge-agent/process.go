package main

import (
	"fmt"
	"path"

	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
)

func (a *agent) processing() error {
	ol, err := newOTALog(a.cfg.OTA, a, nil, a.ctx.Log().WithField("agent", "otalog"))
	if err != nil {
		a.ctx.Log().WithError(err).Warnf("failed to create ota log from log file")
		return err
	}
	if ol != nil {
		ol.wait()
	}
	for {
		select {
		case e := <-a.events:
			a.cleaner.reset()
			a.processEvent(e)
		case <-a.tomb.Dying():
			return nil
		}
	}
}

func (a *agent) processEvent(e *Event) {
	eo := e.Content.(*EventOTA)
	a.ctx.Log().Infof("process ota: type=%s, trace=%s", eo.Type, eo.Trace)
	ol, err := newOTALog(a.cfg.OTA, a, eo, a.ctx.Log().WithField("agent", "otalog"))
	if err != nil {
		a.ctx.Log().WithError(err).Warnf("failed to create ota log")
		return
	}
	defer ol.wait()

	err = a.processOTA(eo)
	if err != nil {
		a.ctx.Log().WithError(err).Warnf("failed to process ota event")
		ol.write(openedge.OTAFailure, "failed to process ota event", err)
	}
}

func (a *agent) processOTA(eo *EventOTA) error {
	hostDir, containerDir, err := a.download(eo.Volume)
	if err != nil {
		return fmt.Errorf("failed to download volume: %s", err.Error())
	}
	var hostTarget string
	if eo.Type == openedge.OTAAPP {
		hostTarget = path.Join(hostDir, openedge.AppConfFileName)
		containerAppFile := path.Join(containerDir, openedge.AppConfFileName)
		var cfg openedge.AppConfig
		err := utils.LoadYAML(containerAppFile, &cfg)
		if err != nil {
			return err
		}
		// check service list, cannot be empty
		if len(cfg.Services) == 0 {
			return fmt.Errorf("app config invalid: service list is empty")
		}
		err = a.downloadAppVolumes(cfg.Volumes)
		if err != nil {
			return fmt.Errorf("failed to download app volumes: %s", err.Error())
		}
		a.cleaner.set(cfg.Version, cfg.Volumes)
	} else if eo.Type == openedge.OTAMST {
		hostTarget = path.Join(hostDir, openedge.DefaultBinFile)
	}
	err = a.ctx.UpdateSystem(eo.Trace, eo.Type, hostTarget)
	if err != nil {
		return fmt.Errorf("failed to update system: %s", err.Error())
	}
	return nil
}
