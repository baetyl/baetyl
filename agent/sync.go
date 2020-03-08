package agent

import (
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/models"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"path"
	"time"
)

type EventType string

type Event struct {
	Time    time.Time   `json:"time"`
	Type    EventType   `json:"event"`
	Content interface{} `json:"content"`
}

func (a *Agent) Process() error {
	for {
		select {
		case e := <-a.events:
			a.processDelta(e)
		case <-a.tomb.Dying():
			return nil
		}
	}
}

func (a *Agent) processDelta(e *Event) {
	le := e.Content.(*EventLink)
	a.log.Info("process ota", log.Any("type", le.Type), log.Any("trace", le.Trace))
	err := a.processResource(le)
	if err != nil {
		a.log.Warn("failed to process ota event", log.Error(err))
	}
}

func (a *Agent) processResource(le *EventLink) error {
	apps, ok := le.Info["apps"]
	if !ok {
		return fmt.Errorf("no application info in delta info")
	}
	bs, err := generateRequest(common.Application, apps)
	if err != nil {
		return err
	}
	res, err := a.syncResource(bs)
	if err != nil {
		return fmt.Errorf("failed to sync resource: %s", err.Error())
	}
	application := res[0].GetApplication()
	if application == nil {
		return fmt.Errorf("failed to get application resource")
	}

	cMap := map[string]string{}
	for _, v := range application.Volumes {
		if v.ConfigMap != nil {
			cMap[v.ConfigMap.Name] = v.ConfigMap.Version
		}
	}
	reqs, err := generateRequest(common.Configuration, cMap)
	if err != nil {
		return err
	}
	res, err = a.syncResource(reqs)
	if err != nil {
		return fmt.Errorf("failed to sync resource: %s", err.Error())
	}
	configs := map[string]models.Configuration{}
	for _, r := range res {
		cfg := r.GetConfiguration()
		if cfg == nil {
			return fmt.Errorf("failed to get config resource")
		}
		configs[cfg.Name] = *cfg
	}

	err = a.processVolumes(application.Volumes, configs)
	if err != nil {
		return err
	}

	err = a.processApplication(*application)
	if err != nil {
		return err
	}
	return nil
}

func (a *Agent) syncResource(res []*config.BaseResource) ([]*config.Resource, error) {
	req := config.DesireRequest{Resources: res}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	resData, err := a.sendRequest("POST", a.cfg.Remote.Desire.URL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to send resource request: %s", err.Error())
	}
	var response config.DesireResponse
	err = json.Unmarshal(resData, &response)
	if err != nil {
		return nil, err
	}
	return response.Resources, nil
}

func (a *Agent) processVolumes(volumes []models.Volume, configs map[string]models.Configuration) error {
	for _, volume := range volumes {
		if configMap := volume.VolumeSource.ConfigMap; configMap != nil {
			err := a.processConfiguration(volume, configs[configMap.Name])
			if err != nil {
				//a.ctx.Log().Errorf("process module config (%s) failed: %s", name, err.Error())
				return err
			}
		} else if secret := volume.VolumeSource.Secret; secret != nil {
			// TODO handle secret
		} else if pvc := volume.VolumeSource.PersistentVolumeClaim; pvc != nil {
			// TODO handle pvc
		} else if hostPath := volume.VolumeSource.HostPath; hostPath != nil {
			// TODO handle hostPath
		}
	}
	return nil
}

func (a *Agent) processConfiguration(volume models.Volume, cfg models.Configuration) error {
	return nil
}

func (a *Agent) processApplication(app models.Application) error {
	// TODO transform app to deployment and apply
	return nil
}

func generateRequest(resType common.Resource, res interface{}) ([]*config.BaseResource, error) {
	var bs []*config.BaseResource
	switch resType {
	case common.Application:
		for n, v := range res.(map[string]interface{}) {
			b := &config.BaseResource{
				Type:    common.Application,
				Name:    n,
				Version: v.(string),
			}
			if b.Name == "" || b.Version == "" {
				return nil, fmt.Errorf("can not request application with empty name or version")
			}
			bs = append(bs, b)
		}
	case common.Configuration:
		cRes := res.(map[string]string)
		filterConfigs(cRes)
		for n, v := range cRes {
			b := &config.BaseResource{
				Type:    common.Configuration,
				Name:    n,
				Version: v,
			}
			bs = append(bs, b)
		}
	}
	return bs, nil
}

func filterConfigs(configs map[string]string) {
	for name, version := range configs {
		configPath := path.Join(baetyl.DefaultDBDir, "volumes", name, version)
		if utils.DirExists(configPath) {
			delete(configs, name)
		}
	}
}
