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
	"gopkg.in/yaml.v2"
)

type deployment struct {
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
	err := a.processDeployment(le)
	if err != nil {
		a.ctx.Log().WithError(err).Warnf("failed to process ota event")
		ol.write(baetyl.OTAFailure, "failed to process ota event", err)
	}
}

func (a *agent) processDeployment(le *EventLink) error {
	d, ok := le.Info[string(common.Deployment)]
	if !ok {
		return fmt.Errorf("no deployment info in delta info")
	}
	var reqs []config.ResourceRequest
	for k, v := range d.(map[string]interface{}) {
		req := config.ResourceRequest{
			Type:    string(common.Deployment),
			Name:    k,
			Version: v.(string),
		}
		if req.Name == "" || req.Version == "" {
			return fmt.Errorf("can not request deployment with empty name or version")
		}
		reqs = append(reqs, req)
	}
	res, err := a.syncResource(reqs)
	if err != nil {
		return fmt.Errorf("failed to sync resource: %s", err.Error())
	}
	var deploy config.Deployment
	for i, req := range reqs {
		switch req.Type {
		case string(common.Deployment):
			err := json.Unmarshal(res[i], &deploy)
			if err != nil {
				return err
			}
		}
	}

	reqs = generateRequest(deploy.Snapshot)
	res, err = a.syncResource(reqs)
	if err != nil {
		return fmt.Errorf("failed to sync resource: %s", err.Error())
	}
	var app config.DeployConfig
	var cfg config.ModuleConfig
	configs := map[string]config.ModuleConfig{}
	for i, req := range reqs {
		switch req.Type {
		case string(common.Application):
			err = json.Unmarshal(res[i], &app)
			if err != nil {
				return err
			}
		case string(common.Config):
			err = json.Unmarshal(res[i], &cfg)
			if err != nil {
				return err
			}
			configs[cfg.Name] = cfg
		}
	}
	volumeMetas, hostDir := a.processApplication(app)

	// avoid duplicated resource synchronization
	var volumes []baetyl.VolumeInfo
	for name, volume := range volumeMetas {
		volumes = append(volumes, volume)
		if _, ok := configs[name]; !ok || volume.Meta.Version == "" {
			delete(volumeMetas, name)
		}
	}
	a.cleaner.set(deploy.Version, volumes)
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

func (a *agent) syncResource(reqs []config.ResourceRequest) ([][]byte, error) {
	data, err := json.Marshal(reqs)
	if err != nil {
		return nil, err
	}
	resData, err := a.sendRequest("POST", a.cfg.Remote.Desire.URL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to send resource request: %s", err.Error())
	}
	var res [][]byte
	err = json.Unmarshal(resData, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (a *agent) processVolumes(volumes map[string]baetyl.VolumeInfo, configs map[string]config.ModuleConfig) error {
	for name, volume := range volumes {
		if meta := volume.Meta; meta.Version != "" {
			err := a.processModuleConfig(volume, configs[name])
			if err != nil {
				a.ctx.Log().Errorf("process module config (%s) failed: %s", name, err.Error())
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
		if strings.HasPrefix(k, common.StorageObjectPrefix) {
			save = false
			obj := new(config.StorageObject)
			err := json.Unmarshal([]byte(v), &obj)
			if err != nil {
				a.ctx.Log().Warnf("process storage object of volume (%s) failed: %s", volume.Name, err.Error())
				save = true
			}
			volume.Meta.URL = obj.URL
			volume.Meta.MD5 = obj.Md5
			_, _, err = a.downloadVolume(volume, strings.TrimPrefix(k, common.StorageObjectPrefix), obj.Compression == common.ZipCompression)
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

func (a *agent) processApplication(deployConfig config.DeployConfig) (map[string]baetyl.VolumeInfo, string) {
	hostDir := path.Join(baetyl.DefaultDBDir, deployConfig.AppConfig.Name, deployConfig.AppConfig.AppVersion)
	containerDir := path.Join(baetyl.DefaultDBDir, "volumes", deployConfig.AppConfig.Name, deployConfig.AppConfig.AppVersion)

	err := os.MkdirAll(containerDir, 0755)
	if err != nil {
		a.ctx.Log().WithError(err).Warnf("failed to prepare app directory (%s): %s", containerDir, err.Error())
		return nil, ""
	}

	data, err := yaml.Marshal(deployConfig.AppConfig)
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
	return deployConfig.Metadata, hostDir
}

func generateRequest(snapshot config.Snapshot) []config.ResourceRequest {
	var reqs []config.ResourceRequest
	for n, v := range snapshot.Apps {
		req := config.ResourceRequest{
			Type:    string(common.Application),
			Name:    n,
			Version: v,
		}
		reqs = append(reqs, req)
	}
	filterConfigs(snapshot.Configs)
	for n, v := range snapshot.Configs {
		req := config.ResourceRequest{
			Type:    string(common.Config),
			Name:    n,
			Version: v,
		}
		reqs = append(reqs, req)
	}
	return reqs
}

func (a *agent) sendRequest(method, path string, body []byte) ([]byte, error) {
	header := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
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

func (a *agent) getCurrentDeploy(inspect *baetyl.Inspect) (map[string]string, error) {
	// TODO 理论上应该总是从inspect获取部署的名称和版本，但是主程序现在没有部署名称信息，临时从application.yml中获取
	var info deployment
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
