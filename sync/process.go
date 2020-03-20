package sync

import (
	"encoding/json"
	"fmt"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"os"
	"path"
	"strings"

	"github.com/baetyl/baetyl-core/common"
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
	appInfo := map[string]string{}
	for name, ver := range info {
		appInfo[name] = ver.(string)
	}
	
	bs, err := s.generateRequest(common.Application, appInfo)
	if err != nil {
		return err
	}
	res, err := s.syncResource(bs)
	if err != nil {
		return fmt.Errorf("failed to sync resource: %s", err.Error())
	}
	cInfo := map[string]string{}
	sInfo := map[string]string{}
	apps := map[string]*v1.Application{}
	for _, r := range res {
		if app := r.GetApplication(); app != nil {
			apps[app.Name] = app
			for _, v := range app.Volumes {
				if v.Config != nil {
					cInfo[v.Config.Name] = v.Config.Version
				} else if v.Secret != nil {
					sInfo[v.Secret.Name] = v.Secret.Version
				}
			}
		} else {
			return fmt.Errorf("failed to sync application (%s) (%s)", r.Name, r.Version)
		}
	}

	reqs, err := s.generateRequest(common.Configuration, cInfo)
	if err != nil {
		return err
	}
	res, err = s.syncResource(reqs)
	if err != nil {
		return fmt.Errorf("failed to sync resource: %s", err.Error())
	}
	configs := map[string]*v1.Configuration{}
	for _, r := range res {
		if cfg := r.GetConfiguration(); cfg != nil {
			configs[cfg.Name] = cfg
		} else {
			return fmt.Errorf("failed to sync configuration (%s) (%s)", r.Name, r.Version)
		}
	}

	reqs, err = s.generateRequest(common.Secret, sInfo)
	if err != nil {
		return err
	}
	res, err = s.syncResource(reqs)
	if err != nil {
		return fmt.Errorf("failed to sync resource: %s", err.Error())
	}
	secrets := map[string]*v1.Secret{}
	for _, r := range res {
		if secret := r.GetSecret(); secret != nil {
			secrets[secret.Name] = secret
		} else {
			return fmt.Errorf("failed to sync secret (%s) (%s)", r.Name, r.Version)
		}
	}

	for _, app := range apps {
		err = s.processVolumes(app.Volumes, configs, secrets)
		if err != nil {
			return err
		}
		// app.volume may change when processing Volumes
		err := s.storeApplication(app)
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

func (s *Sync) processVolumes(volumes []v1.Volume, configs map[string]*v1.Configuration, secrets map[string]*v1.Secret) error {
	for _, volume := range volumes {
		if cfg := volume.VolumeSource.Config; cfg != nil && configs[cfg.Name] != nil {
			err := s.processConfiguration(&volume, configs[cfg.Name])
			if err != nil {
				//a.ctx.Log().Errorf("process module config (%s) failed: %s", name, err.Error())
				return err
			}
		} else if secret := volume.VolumeSource.Secret; secret != nil && secrets[secret.Name] != nil {
			err := s.storeSecret(secrets[secret.Name])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Sync) processConfiguration(volume *v1.Volume, cfg *v1.Configuration) error {
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
			// change app.volume from config to host path of downloaded file path
			volume.Config = nil
			volume.HostPath = &v1.HostPathVolumeSource{
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

func (s *Sync) storeApplication(app *v1.Application) error {
	key := utils.MakeKey(common.Application, app.Name, app.Version)
	if key == "" {
		return fmt.Errorf("app does not have name or version")
	}
	return s.store.Upsert(key, app)
}

func (s *Sync) storeSecret(secret *v1.Secret) error {
	key := utils.MakeKey(common.Secret, secret.Name, secret.Version)
	if key == "" {
		return fmt.Errorf("secret does not have name or version")
	}
	return s.store.Upsert(key, secret)
}

func (s *Sync) generateRequest(resType common.Resource, res map[string]string) (bs []*BaseResource, err error) {
	var num int
	for name, version := range res {
		num = 0
		switch resType {
		case common.Configuration:
			num, err = s.store.Count(&v1.Configuration{}, nil)
		case common.Secret:
			num, err = s.store.Count(&v1.Secret{}, nil)
		}
		if err != nil {
			return nil, err
		}
		if num > 0 {
			delete(res, name)
		}
		b := &BaseResource{
			Type:    resType,
			Name:    name,
			Version: version,
		}
		bs = append(bs, b)
	}
	return
}
