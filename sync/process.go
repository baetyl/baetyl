package sync

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/models"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/mqtt"
)

func NewEvent(v []byte) (*Event, error) {
	e := Event{}
	err := json.Unmarshal(v, &e.Content)
	if err != nil {
		return nil, fmt.Errorf("event content invalid: %s", err.Error())
	}
	return &e, nil
}

func (s *sync) OnPublish(p *mqtt.Publish) error {
	if p.Message.QOS == 1 {
		puback := mqtt.NewPuback()
		puback.ID = p.ID
		err := s.mqtt.Send(puback)
		if err != nil {
			return err
		}
	}
	e, err := NewEvent(p.Message.Payload)
	if err != nil {
		return err
	}
	select {
	case oe := <-s.events:
		s.log.Warn("discard old event", log.Any("event", *oe))
		s.events <- e
	case s.events <- e:
	case <-s.tomb.Dying():
	}
	return nil
}

func (s *sync) OnPuback(*mqtt.Puback) error {
	return nil
}

func (s *sync) OnError(err error) {
	if err != nil {
		s.log.Error("get mqtt error", log.Error(err))
	}
}

func (s *sync) processing() error {
	for {
		select {
		case e := <-s.events:
			s.processDelta(e)
		case <-s.tomb.Dying():
			return nil
		}
	}
}

func (s *sync) processDelta(e *Event) {
	s.log.Info("process ota", log.Any("type", e.Type), log.Any("trace", e.Trace))
	err := s.ProcessResource(e.Content)
	if err != nil {
		s.log.Warn("failed to process ota event", log.Error(err))
	}
}

func (s *sync) ProcessResource(content interface{}) error {
	info := content.(map[string]interface{})
	aMap, ok := info["apps"]
	if !ok {
		return fmt.Errorf("no application info in delta info")
	}
	bs, err := generateRequest(common.Application, aMap)
	if err != nil {
		return err
	}
	res, err := s.syncResource(bs)
	if err != nil {
		return fmt.Errorf("failed to sync resource: %s", err.Error())
	}
	cMap := map[string]string{}
	apps := map[string]*models.Application{}
	for _, r := range res {
		app := r.GetApplication()
		if app != nil {
			apps[app.Name] = app
			for _, v := range app.Volumes {
				if v.Configuration != nil {
					cMap[v.Configuration.Name] = v.Configuration.Version
				}
			}
		}
	}

	reqs, err := generateRequest(common.Configuration, cMap)
	if err != nil {
		return err
	}
	res, err = s.syncResource(reqs)
	if err != nil {
		return fmt.Errorf("failed to sync resource: %s", err.Error())
	}
	configs := map[string]*models.Configuration{}
	for _, r := range res {
		cfg := r.GetConfiguration()
		if cfg == nil {
			return fmt.Errorf("failed to get config resource")
		}
		configs[cfg.Name] = cfg
	}

	for _, app := range apps {
		err := s.ProcessApplication(app)
		if err != nil {
			return err
		}
		err = s.ProcessVolumes(app.Volumes, configs)
		if err != nil {
			return err
		}
	}

	// TODO start k8s engine
	return nil
}

func (s *sync) syncResource(res []*config.BaseResource) ([]*config.Resource, error) {
	req := config.DesireRequest{Resources: res}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	resp, err := s.http.Post(s.cfg.Cloud.Desire.URL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to send resource request: %s", err.Error())
	}
	data, err = http.HandleResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to send resource request: %s", err.Error())
	}
	var response config.DesireResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}
	return response.Resources, nil
}

func (s *sync) ProcessVolumes(volumes []models.Volume, configs map[string]*models.Configuration) error {
	for _, volume := range volumes {
		if cfg := volume.VolumeSource.Configuration; cfg != nil {
			err := s.ProcessConfiguration(volume, configs[cfg.Name])
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

func (s *sync) ProcessConfiguration(volume models.Volume, cfg *models.Configuration) error {
	key := makeKey(common.Configuration, cfg.Name, cfg.Version)
	if key == nil {
		return fmt.Errorf("configuration does not have name or version")
	}
	return s.store.Insert(key, cfg)
}

func (s *sync) ProcessApplication(app *models.Application) error {
	key := makeKey(common.Application, app.Name, app.Version)
	if key == nil {
		return fmt.Errorf("app does not have name or version")
	}
	return s.store.Insert(key, app)
}

func makeKey(resType common.Resource, name, ver string) []byte {
	if name == "" || ver == "" {
		return nil
	}
	return []byte(string(resType) + "/" + name + "/" + ver)
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
		// filterConfigs(cRes)
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

// func filterConfigs(configs map[string]string) {
// 	for name, version := range configs {
// 		configPath := path.Join(baetyl.DefaultDBDir, "volumes", name, version)
// 		if utils.DirExists(configPath) {
// 			delete(configs, name)
// 		}
// 	}
// }
