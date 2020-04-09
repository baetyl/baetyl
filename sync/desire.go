package sync

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/spec/crd"
	"github.com/baetyl/baetyl-go/spec/v1"
)

// extended features of config resourece
const (
	configKeyObject = "_object_"
	configValueZip  = "zip"
)

func (s *Sync) syncResources(ais []v1.AppInfo) error {
	if len(ais) == 0 {
		return nil
	}

	appInfo := map[string]string{}
	for _, a := range ais {
		appInfo[a.Name] = a.Version
	}
	crds, err := s.syncCRDs(s.genCRDInfos(crd.KindApplication, appInfo))
	if err != nil {
		s.log.Error("failed to sync application resource", log.Error(err))
		return err
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
		s.log.Error("failed to sync configuration resource", log.Error(err))
		return err
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
		s.log.Error("failed to sync secret resource", log.Error(err))
		return err
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
			s.log.Error("failed to process volumes", log.Error(err))
			return err
		}
		// app.volume may change when processing Volumes
		err := s.storeApplication(app)
		if err != nil {
			s.log.Error("failed to store application", log.Error(err))
			return err
		}
	}
	return nil
}

func (s *Sync) syncCRDs(crds []v1.CRDInfo) ([]v1.CRDData, error) {
	if len(crds) == 0 {
		return nil, nil
	}
	req := v1.CRDRequest{CRDInfos: crds}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	data, err = s.http.PostJSON(s.cfg.Cloud.Desire.URL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to send resource request: %s", err.Error())
	}
	var res v1.CRDResponse
	err = json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res.CRDDatas, nil
}

func (s *Sync) processVolumes(volumes []crd.Volume, configs map[string]*crd.Configuration, secrets map[string]*crd.Secret) error {
	for i := range volumes {
		if cfg := volumes[i].VolumeSource.Config; cfg != nil && configs[cfg.Name] != nil {
			err := s.processConfiguration(&volumes[i], configs[cfg.Name])
			if err != nil {
				return err
			}
		} else if secret := volumes[i].VolumeSource.Secret; secret != nil && secrets[secret.Name] != nil {
			err := s.storeSecret(secrets[secret.Name])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Sync) processConfiguration(volume *crd.Volume, cfg *crd.Configuration) error {
	var base, dir string
	for k, v := range cfg.Data {
		if strings.HasPrefix(k, configKeyObject) {
			if base == "" {
				base = filepath.Join(s.cfg.Edge.DownloadPath, cfg.Name)
				dir = filepath.Join(base, cfg.Version)
				err := os.MkdirAll(dir, 0755)
				if err != nil {
					return err
				}
			}
			obj := new(v1.CRDConfigObject)
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
			if volume.HostPath == nil {
				volume.Config = nil
				volume.HostPath = &crd.HostPathVolumeSource{
					Path: dir,
				}
				err = cleanDir(base, cfg.Version)
				if err != nil {
					return err
				}
			}
		}
	}
	key := makeKey(crd.KindConfiguration, cfg.Name, cfg.Version)
	if key == "" {
		return fmt.Errorf("configuration does not have name or version")
	}
	return s.store.Upsert(key, cfg)
}

func cleanDir(dir, retain string) error {
	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		if f.Name() != retain {
			os.RemoveAll(filepath.Join(dir, f.Name()))
		}
	}
	return nil
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

func (s *Sync) genCRDInfos(kind crd.Kind, infos map[string]string) []v1.CRDInfo {
	var crds []v1.CRDInfo
	for name, version := range infos {
		crds = append(crds, v1.CRDInfo{
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
	return string(kind) + "-" + name + "-" + ver
}
