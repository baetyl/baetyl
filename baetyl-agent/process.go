package main

import (
	"fmt"
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
			a.processEvent(e)
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
	// transform volume format to compose volume format
	v := baetyl.ComposeVolume{
		DriverOpts: map[string]string{
			"device": eo.Volume.Path,
		},
		Meta: eo.Volume.Meta,
	}

	hostDir, containerDir, err := a.downloadVolume(eo.Volume.Name, v)
	if err != nil {
		return fmt.Errorf("failed to download volume: %s", err.Error())
	}
	var hostTarget string
	if eo.Type == baetyl.OTAAPP {
		hostTarget = path.Join(hostDir, baetyl.AppConfFileName)
		containerAppFile := path.Join(containerDir, baetyl.AppConfFileName)

		var cfg baetyl.ComposeAppConfig
		err = utils.LoadYAML(containerAppFile, &cfg)
		if err != nil {
			var c baetyl.AppConfig
			err = utils.LoadYAML(containerAppFile, &c)
			if err != nil {
				return err
			}
			cfg = baetyl.ToComposeAppConfig(c)
		}

		err := utils.LoadYAML(containerAppFile, &cfg)
		if err != nil {
			return err
		}
		// check service list, cannot be empty
		if len(cfg.Services) == 0 {
			return fmt.Errorf("app config invalid: service list is empty")
		}
		err = a.downloadVolumes(cfg.Volumes)
		if err != nil {
			return fmt.Errorf("failed to download app volumes: %s", err.Error())
		}
		a.cleaner.set(cfg.AppVersion, cfg.Volumes)
	} else if eo.Type == baetyl.OTAMST {
		hostTarget = path.Join(hostDir, baetyl.DefaultBinFile)
	}
	err = a.ctx.UpdateSystem(eo.Trace, eo.Type, hostTarget)
	if err != nil {
		return fmt.Errorf("failed to update system: %s", err.Error())
	}
	return nil
}
