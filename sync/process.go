package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/models"
	"github.com/baetyl/baetyl-core/utils"
	"github.com/baetyl/baetyl-go/faas"
	"github.com/baetyl/baetyl-go/log"
)

func (s *Sync) ProcessDelta(msg faas.Message) error {
	var delta map[string]interface{}
	err := json.Unmarshal(msg.Payload, &delta)
	if err != nil {
		return err
	}
	info, ok := delta[common.DefaultAppsKey].(map[string]interface{})
	if !ok {
		return fmt.Errorf("apps does not exist")
	}
	apps := map[string]string{}
	for name, ver := range info {
		apps[name] = ver.(string)
	}
	bs, err := s.generateRequest(common.Application, apps)
	if err != nil {
		return err
	}
	res, err := s.syncResource(bs)
	if err != nil {
		return fmt.Errorf("failed to sync resource: %s", err.Error())
	}
	cMap := map[string]string{}
	aMap := map[string]*models.Application{}
	for _, r := range res {
		app := r.GetApplication()
		if app != nil {
			aMap[app.Name] = app
			for _, v := range app.Volumes {
				if v.Configuration != nil {
					cMap[v.Configuration.Name] = v.Configuration.Version
				}
			}
		}
	}
	reqs, err := s.generateRequest(common.Configuration, cMap)
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
	for _, app := range aMap {
		err := s.ProcessApplication(app)
		if err != nil {
			return err
		}
		err = s.ProcessVolumes(app.Volumes, configs)
		if err != nil {
			return err
		}
	}
	message := &faas.Message{Metadata: map[string]string{"topic": common.EngineAppEvent}, Payload: msg.Payload}
	err = s.cent.Trigger(message)
	if err != nil {
		return err
	}
	return nil
}

func (s *Sync) syncResource(res []*BaseResource) ([]*Resource, error) {
	req := DesireRequest{Resources: res}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	data, err = s.http.PostJSON(s.cfg.Cloud.Desire.URL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to send resource request: %s", err.Error())
	}
	var response DesireResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}
	return response.Resources, nil
}

func (s *Sync) ProcessVolumes(volumes []models.Volume, configs map[string]*models.Configuration) error {
	for _, volume := range volumes {
		if cfg := volume.VolumeSource.Configuration; cfg != nil && configs[cfg.Name] != nil {
			err := s.ProcessConfiguration(&volume, configs[cfg.Name])
			if err != nil {
				//a.ctx.Log().Errorf("process module config (%s) failed: %s", name, err.Error())
				return err
			}
		} else if secret := volume.VolumeSource.Secret; secret != nil {
			// TODO handle secret
		}
	}
	return nil
}

func (s *Sync) ProcessConfiguration(volume *models.Volume, cfg *models.Configuration) error {
	var dir string
	for k, v := range cfg.Data {
		if strings.HasPrefix(k, common.PrefixConfigObject) {
			if dir == "" {
				dir = path.Join(s.cfg.Edge.DownloadPath, cfg.Name, cfg.Version)
			}
			obj := new(StorageObject)
			err := json.Unmarshal([]byte(v), &obj)
			if err != nil {
				s.log.Warn("process storage object of volume failed: %s", log.Any("name", volume.Name), log.Error(err))
				return err
			}
			filename := path.Join(dir, strings.TrimPrefix(k, common.PrefixConfigObject))
			err = s.downloadFile(obj, dir, filename, obj.Compression == common.ZipCompression)
			if err != nil {
				os.RemoveAll(dir)
				return fmt.Errorf("failed to download volume (%s) with error: %s", volume.Name, err)
			}
			volume.Configuration = nil
			volume.HostPath = &models.HostPathVolumeSource{
				Path: dir,
			}
		}
	}

	key := utils.MakeKey(common.Configuration, cfg.Name, cfg.Version)
	if key == "" {
		return fmt.Errorf("configuration does not have name or version")
	}
	return s.store.Upsert(key, cfg)
}

func (s *Sync) ProcessApplication(app *models.Application) error {
	key := utils.MakeKey(common.Application, app.Name, app.Version)
	if key == "" {
		return fmt.Errorf("app does not have name or version")
	}
	return s.store.Upsert(key, app)
}

func (s *Sync) generateRequest(resType common.Resource, res map[string]string) ([]*BaseResource, error) {
	var bs []*BaseResource
	switch resType {
	case common.Application:
		//s.filterApps(res)
		for n, v := range res {
			b := &BaseResource{
				Type:    common.Application,
				Name:    n,
				Version: v,
			}
			if b.Name == "" || b.Version == "" {
				return nil, fmt.Errorf("can not request application with empty name or version")
			}
			bs = append(bs, b)
		}
	case common.Configuration:
		//s.filterConfigs(res)
		for n, v := range res {
			b := &BaseResource{
				Type:    common.Configuration,
				Name:    n,
				Version: v,
			}
			bs = append(bs, b)
		}
	}
	return bs, nil
}

func (s *Sync) filterApps(apps map[string]string) {
	for name, ver := range apps {
		key := utils.MakeKey(common.Application, name, ver)
		var app models.Application
		err := s.store.Get(key, &app)
		if err != nil {
			s.log.Error("failed to get app", log.Error(err))
			continue
		}
		if app.Version != "" {
			delete(apps, name)
		}
	}
}

func (s *Sync) filterConfigs(configs map[string]string) {
	for name, ver := range configs {
		key := utils.MakeKey(common.Configuration, name, ver)
		var config models.Configuration
		err := s.store.Get(key, &config)
		if err != nil {
			s.log.Error("failed to get config", log.Error(err))
			continue
		}
		if config.Version != "" {
			delete(configs, name)
		}
	}
}
