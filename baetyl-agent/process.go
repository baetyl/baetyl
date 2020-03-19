package main

import (
	"fmt"
	"github.com/baetyl/baetyl/baetyl-agent/config"
	"path"

	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
)

func (a *agent) processing() error {
	ol := newOTALog(a.cfg.OTA, a, nil, a.ctx.Log().WithField("agent", "otalog"))
	if ol != nil {
		ol.wait()
	}
	for {
		select {
		case e := <-a.events:
			a.cleaner.reset()
			if a.mqtt == nil {
				a.processDelta(e)
			} else {
				a.processEvent(e)
			}
		case <-a.tomb.Dying():
			return nil
		}
	}
}

func (a *agent) processEvent(e *Event) {
	eo := e.Content.(*EventOTA)
	a.ctx.Log().Infof("process ota: type=%s, trace=%s", eo.Type, eo.Trace)
	ol := newOTALog(a.cfg.OTA, a, eo, a.ctx.Log().WithField("agent", "otalog"))
	defer ol.wait()

	err := a.processOTA(eo)
	if err != nil {
		a.ctx.Log().WithError(err).Warnf("failed to process ota event")
		ol.write(baetyl.OTAFailure, "failed to process ota event", err)
	}
}

func (a *agent) processOTA(eo *EventOTA) error {
	hostDir, containerDir, err := a.downloadVolume(eo.Volume, "", true)
	if err != nil {
		return fmt.Errorf("failed to download volume: %s", err.Error())
	}
	var hostTarget string
	if eo.Type == baetyl.OTAAPP {
		hostTarget = path.Join(hostDir, baetyl.AppConfFileName)
		containerAppFile := path.Join(containerDir, baetyl.AppConfFileName)
		containerMetadataFile := path.Join(containerDir, baetyl.MetadataFileName)
		var meta config.Metadata
		file := containerMetadataFile
		if !utils.FileExists(containerMetadataFile) {
			file = containerAppFile
		}
		if err = utils.LoadYAML(file, &meta); err != nil {
			return err
		}
		cfg, err := baetyl.LoadComposeAppConfigCompatible(containerAppFile)
		if err != nil {
			return err
		}
		// check service list, cannot be empty
		if len(cfg.Services) == 0 {
			return fmt.Errorf("app config invalid: service list is empty")
		}
		err = a.downloadVolumes(meta.Volumes)
		if err != nil {
			return fmt.Errorf("failed to download app volumes: %s", err.Error())
		}
		a.cleaner.set(meta.Version, meta.Volumes)
	} else if eo.Type == baetyl.OTAMST {
		hostTarget = path.Join(hostDir, baetyl.DefaultBinFile)
	}
	err = a.ctx.UpdateSystem(eo.Trace, eo.Type, hostTarget)
	if err != nil {
		return fmt.Errorf("failed to update system: %s", err.Error())
	}
	return nil
}
