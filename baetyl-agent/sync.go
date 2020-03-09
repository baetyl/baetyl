package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/baetyl/baetyl/baetyl-agent/common"
	"github.com/baetyl/baetyl/baetyl-agent/config"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	"github.com/docker/distribution/uuid"
	"gopkg.in/yaml.v2"
)

type application struct {
	Name       string `yaml:"name" json:"name"`
	AppVersion string `yaml:"app_version" json:"app_version"`
}

type node struct {
	Name      string
	Namespace string
}

type batch struct {
	Name      string
	Namespace string
}

func (a *agent) processDelta(e *Event) {
	le := e.Content.(*EventLink)
	a.ctx.Log().Infof("process ota: type=%s, trace=%s", le.Type, le.Trace)
	ol := newOTALog(a.cfg.OTA, a, &EventOTA{Type: le.Type, Trace: le.Trace}, a.ctx.Log().WithField("agent", "otalog"))
	defer ol.wait()
	err := a.processResource(le)
	if err != nil {
		a.ctx.Log().WithError(err).Warnf("failed to process ota event")
		ol.write(baetyl.OTAFailure, "failed to process ota event", err)
	}
}

func (a *agent) processResource(le *EventLink) error {
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
	for _, v := range application.Metadata {
		if v.Meta.Version != "" {
			cMap[v.Name] = v.Meta.Version
		}
	}
	reqs, err := generateRequest(common.Config, cMap)
	if err != nil {
		return err
	}
	res, err = a.syncResource(reqs)
	if err != nil {
		return fmt.Errorf("failed to sync resource: %s", err.Error())
	}
	configs := map[string]config.ModuleConfig{}
	for _, r := range res {
		cfg := r.GetConfig()
		if cfg == nil {
			return fmt.Errorf("failed to get config resource")
		}
		configs[cfg.Name] = *cfg
	}

	volumeMetas, hostDir := a.processApplication(*application)

	// avoid duplicated resource synchronization
	var volumes []baetyl.VolumeInfo
	for _, volume := range volumeMetas {
		volumes = append(volumes, volume)
		if _, ok := configs[volume.Name]; !ok || volume.Meta.Version == "" {
			delete(volumeMetas, volume.Name)
		}
	}
	a.cleaner.set(application.AppConfig.AppVersion, volumes)
	err = a.processVolumes(volumeMetas, configs)
	if err != nil {
		return err
	}
	err = a.ctx.UpdateSystem(le.Trace, le.Type, hostDir)
	if err != nil {
		return fmt.Errorf("failed to update system: %s", err.Error())
	}
	return nil
}

func (a *agent) syncResource(res []*config.BaseResource) ([]*config.Resource, error) {
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

func (a *agent) processVolumes(volumes map[string]baetyl.VolumeInfo, configs map[string]config.ModuleConfig) error {
	for _, volume := range volumes {
		if meta := volume.Meta; meta.Version != "" {
			err := a.processModuleConfig(volume, configs[volume.Name])
			if err != nil {
				a.ctx.Log().Errorf("process module config (%s) failed: %s", volume.Name, err.Error())
				return err
			}
		}
	}
	return nil
}

func (a *agent) processModuleConfig(volume baetyl.VolumeInfo, cfg config.ModuleConfig) error {
	rootDir := path.Join(baetyl.DefaultDBDir, "volumes")
	rp, err := filepath.Rel(baetyl.DefaultDBDir, volume.Path)
	if err != nil {
		return fmt.Errorf("illegal path of config (%s): %s", cfg.Name, err.Error())
	}
	containerDir := path.Join(rootDir, rp)
	err = os.MkdirAll(containerDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to prepare volume directory (%s): %s", containerDir, err.Error())
	}
	save := true
	for k, v := range cfg.Data {
		if strings.HasPrefix(k, common.PrefixConfigObject) {
			save = false
			obj := new(config.StorageObject)
			err := json.Unmarshal([]byte(v), &obj)
			if err != nil {
				a.ctx.Log().Warnf("process storage object of volume (%s) failed: %s", volume.Name, err.Error())
				save = true
			}
			volume.Meta.URL = obj.URL
			volume.Meta.MD5 = obj.Md5
			_, _, err = a.downloadVolume(volume, strings.TrimPrefix(k, common.PrefixConfigObject), obj.Compression == common.ZipCompression)
			if err != nil {
				return fmt.Errorf("failed to download volume (%s) with error: %s", volume.Name, err)
			}
		}
		if save {
			volumeFile := path.Join(containerDir, k)
			err = ioutil.WriteFile(volumeFile, []byte(v), 0755)
			if err != nil {
				os.RemoveAll(containerDir)
				return fmt.Errorf("failed to create file (%s): %s", volumeFile, err.Error())
			}
		}
	}
	return nil
}

func (a *agent) processApplication(application config.Application) (map[string]baetyl.VolumeInfo, string) {
	hostDir := path.Join(baetyl.DefaultDBDir, application.AppConfig.Name, application.AppConfig.AppVersion)
	containerDir := path.Join(baetyl.DefaultDBDir, "volumes", application.AppConfig.Name, application.AppConfig.AppVersion)

	err := os.MkdirAll(containerDir, 0755)
	if err != nil {
		a.ctx.Log().WithError(err).Warnf("failed to prepare app directory (%s): %s", containerDir, err.Error())
		return nil, ""
	}

	data, err := yaml.Marshal(application.AppConfig)
	if err != nil {
		os.RemoveAll(containerDir)
		a.ctx.Log().WithError(err).Warnf("failed to marshal app config: %s", err.Error())
		return nil, ""
	}
	err = ioutil.WriteFile(path.Join(containerDir, baetyl.AppConfFileName), data, 0644)
	if err != nil {
		os.RemoveAll(containerDir)
		a.ctx.Log().WithError(err).Warnf("failed to write application config into file (%s): %s", baetyl.AppConfFileName, err.Error())
		return nil, ""
	}
	return application.Metadata, hostDir
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
	case common.Config:
		cRes := res.(map[string]string)
		filterConfigs(cRes)
		for n, v := range cRes {
			b := &config.BaseResource{
				Type:    common.Config,
				Name:    n,
				Version: v,
			}
			bs = append(bs, b)
		}
	}
	return bs, nil
}

func (a *agent) sendRequest(method, path string, body []byte) ([]byte, error) {
	header := map[string]string{
		"x-bce-request-id": uuid.Generate().String(),
		"Content-Type":     "application/json",
	}
	if a.node != nil {
		// for report
		header[common.HeaderKeyNodeNamespace] = a.node.Namespace
		header[common.HeaderKeyNodeName] = a.node.Name
	} else if a.batch != nil {
		// for active
		header[common.HeaderKeyBatchNamespace] = a.batch.Namespace
		header[common.HeaderKeyBatchName] = a.batch.Name
	}
	a.ctx.Log().Debugf("request, method=%s, path=%s , body=%s , header=%+v", method, path, string(body), header)
	return a.http.SendPath(method, path, body, header)
}

func (a *agent) getCurrentApp(inspect *baetyl.Inspect) (map[string]string, error) {
	// TODO 理论上应该总是从inspect获取部署的名称和版本，但是主程序现在没有部署名称信息，临时从application.yml中获取
	var info application
	err := utils.LoadYAML(path.Join(baetyl.DefaultDBDir, "volumes", baetyl.AppConfFileName), &info)
	if err != nil {
		return nil, err
	}
	if inspect.Software.ConfVersion == "" {
		return nil, fmt.Errorf("app version is empty")
	}
	return map[string]string{
		info.Name: inspect.Software.ConfVersion,
	}, nil
}

func filterConfigs(configs map[string]string) {
	for name, version := range configs {
		configPath := path.Join(baetyl.DefaultDBDir, "volumes", name, version)
		if utils.DirExists(configPath) {
			delete(configs, name)
		}
	}
}
