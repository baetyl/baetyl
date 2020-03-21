package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/baetyl/baetyl-core/event"
	"github.com/baetyl/baetyl-go/faas"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/spec"
	"github.com/baetyl/baetyl-go/spec/api"
	"github.com/baetyl/baetyl-go/spec/crd"
)

// extended features of config resourece
const (
	configKeyObject = "_object_"
	configValueZip  = "zip"
)

// Desire process desire delta, to sync crds
func (s *Sync) Desire(msg faas.Message) error {
	var delta spec.Desire
	err := json.Unmarshal(msg.Payload, &delta)
	if err != nil {
		return err
	}
	ais := delta.AppInfos()
	if len(ais) == 0 {
		return fmt.Errorf("apps does not exist")
	}
	appInfo := map[string]string{}
	for _, a := range ais {
		appInfo[a.Name] = a.Version
	}
	crds, err := s.syncCRDs(s.genCRDInfos(crd.KindApplication, appInfo))
	if err != nil {
		return fmt.Errorf("failed to sync resource: %s", err.Error())
	}
	cInfo := map[string]string{}
	sInfo := map[string]string{}
	apps := map[string]*crd.Application{}
	for _, r := range crds {
		if app := r.App(); app != nil {
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

	crds, err = s.syncCRDs(s.genCRDInfos(crd.KindConfiguration, cInfo))
	if err != nil {
		return fmt.Errorf("failed to sync resource: %s", err.Error())
	}
	configs := map[string]*crd.Configuration{}
	for _, r := range crds {
		if cfg := r.Config(); cfg != nil {
			configs[cfg.Name] = cfg
		} else {
			return fmt.Errorf("failed to sync configuration (%s) (%s)", r.Name, r.Version)
		}
	}

	crds, err = s.syncCRDs(s.genCRDInfos(crd.KindSecret, sInfo))
	if err != nil {
		return fmt.Errorf("failed to sync resource: %s", err.Error())
	}
	secrets := map[string]*crd.Secret{}
	for _, r := range crds {
		if secret := r.Secret(); secret != nil {
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
	message := &faas.Message{Metadata: map[string]string{"topic": event.EngineAppEvent}, Payload: msg.Payload}
	err = s.cent.Trigger(message)
	if err != nil {
		return err
	}
	return nil
}

func (s *Sync) syncCRDs(crds []api.CRDInfo) ([]api.CRDData, error) {
	if len(crds) == 0 {
		return nil, nil
	}
	req := api.CRDRequest{CRDInfos: crds}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	data, err = s.http.PostJSON(s.cfg.Cloud.Desire.URL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to send resource request: %s", err.Error())
	}
	var response api.CRDResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}
	return response.CRDDatas, nil
}

func (s *Sync) processVolumes(volumes []crd.Volume, configs map[string]*crd.Configuration, secrets map[string]*crd.Secret) error {
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

func (s *Sync) processConfiguration(volume *crd.Volume, cfg *crd.Configuration) error {
	var dir string
	for k, v := range cfg.Data {
		if strings.HasPrefix(k, configKeyObject) {
			if dir == "" {
				dir = path.Join(s.cfg.Edge.DownloadPath, cfg.Name, cfg.Version)
			}
			obj := new(api.CRDConfigObject)
			err := json.Unmarshal([]byte(v), &obj)
			if err != nil {
				s.log.Warn("process storage object of volume failed: %s", log.Any("name", volume.Name), log.Error(err))
				return err
			}
			filename := path.Join(dir, strings.TrimPrefix(k, configKeyObject))
			err = s.downloadObject(obj, dir, filename, obj.Compression == configValueZip)
			if err != nil {
				os.RemoveAll(dir)
				return fmt.Errorf("failed to download volume (%s) with error: %s", volume.Name, err)
			}
			// change app.volume from config to host path of downloaded file path
			volume.Config = nil
			volume.HostPath = &crd.HostPathVolumeSource{
				Path: dir,
			}
		}
	}

	key := makeKey(crd.KindConfiguration, cfg.Name, cfg.Version)
	if key == "" {
		return fmt.Errorf("configuration does not have name or version")
	}
	return s.store.Upsert(key, cfg)
}

func (s *Sync) storeApplication(app *crd.Application) error {
	key := makeKey(crd.KindApplication, app.Name, app.Version)
	if key == "" {
		return fmt.Errorf("app does not have name or version")
	}
	return s.store.Upsert(key, app)
}

func (s *Sync) storeSecret(secret *crd.Secret) error {
	key := makeKey(crd.KindSecret, secret.Name, secret.Version)
	if key == "" {
		return fmt.Errorf("secret does not have name or version")
	}
	return s.store.Upsert(key, secret)
}

func (s *Sync) genCRDInfos(kind crd.Kind, infos map[string]string) []api.CRDInfo {
	var crds []api.CRDInfo
	for name, version := range infos {
		crds = append(crds, api.CRDInfo{
			Kind:    kind,
			Name:    name,
			Version: version,
		})
	}
	return crds
}

func makeKey(kind crd.Kind, name, ver string) string {
	if name == "" || ver == "" {
		return ""
	}
	return string(kind) + "/" + name + "/" + ver
}
