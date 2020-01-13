package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/baetyl/baetyl/baetyl-agent/common"
	"github.com/baetyl/baetyl/baetyl-agent/config"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	"gopkg.in/yaml.v2"
)

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

	data, err := json.Marshal(reqs)
	if err != nil {
		return fmt.Errorf("failed to marshal resource request: %s", err.Error())
	}
	resData, err := a.sendRequest("POST", a.cfg.Remote.Desire.URL, data)
	if err != nil {
		return fmt.Errorf("failed to send resource request: %s", err.Error())
	}
	var deployRes config.ResourceResponse
	err = json.Unmarshal(resData, &deployRes)
	if err != nil {
		return fmt.Errorf("failed to unmarshal deployment resource: %s", err.Error())
	}
	data, err = generateRequest(deployRes.Deployment.Snapshot)
	if err != nil {
		return fmt.Errorf("failed to generate resource request of application and configs: %s", err.Error())
	}
	resData, err = a.sendRequest("POST", a.cfg.Remote.Desire.URL, data)
	if err != nil {
		return fmt.Errorf("failed to send resource request: %s", err.Error())
	}
	var res config.ResourceResponse
	err = json.Unmarshal(resData, &res)
	if err != nil {
		return fmt.Errorf("failed to unmarshal application and config resource: %s", err.Error())
	}
	volumeMetas, hostDir := a.processApplication(res.Application)

	// avoid duplicated resource synchronization
	var volumes []baetyl.VolumeInfo
	for name, volume := range volumeMetas {
		volumes = append(volumes, volume)
		metaVersion := volume.Meta.Version
		if metaVersion == "" {
			delete(volumeMetas, name)
		}
		if a.checkVolumeExists(volume) {
			delete(volumeMetas, name)
		}
	}
	a.cleaner.set(deployRes.Deployment.Version, volumes)
	err = a.processVolumes(volumeMetas, res.Configs)
	if err != nil {
		return err
	}
	err = a.ctx.UpdateSystem(le.Trace, le.Type, hostDir)
	if err != nil {
		return fmt.Errorf("failed to update system: %s", err.Error())
	}
	return nil
}

func (a *agent) processVolumes(volumes map[string]baetyl.VolumeInfo, configs map[string]config.ModuleConfig) error {
	rootDir := path.Join(baetyl.DefaultDBDir, "volumes")
	for name, volume := range volumes {
		if meta := volume.Meta; meta.Version != "" {
			if volume.Meta.URL != "" {
				err := a.processURL(rootDir, name, volume)
				if err != nil {
					a.ctx.Log().Errorf("download volume (%s) failed: %s", name, err.Error())
				}
			} else {
				err := a.processModuleConfig(rootDir, volume.Path, configs[name])
				if err != nil {
					a.ctx.Log().Errorf("download module config (%s) failed: %s", name, err.Error())
				}
			}
		}
	}
	return nil
}

func (a *agent) processModuleConfig(rootDir string, volumePath string, config config.ModuleConfig) error {
	rp, err := filepath.Rel(baetyl.DefaultDBDir, volumePath)
	if err != nil {
		return fmt.Errorf("illegal path of config (%s): %s", config.Name, err.Error())
	}
	containerDir := path.Join(rootDir, rp)
	err = os.MkdirAll(containerDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to prepare volume directory (%s): %s", containerDir, err.Error())
	}
	for file, content := range config.Data {
		volumeFile := path.Join(containerDir, file)
		err = ioutil.WriteFile(volumeFile, []byte(content), 0755)
		if err != nil {
			os.RemoveAll(containerDir)
			return fmt.Errorf("failed to create file (%s): %s", volumeFile, err.Error())
		}
	}
	return nil
}

func (a *agent) processURL(rootDir string, name string, volume baetyl.VolumeInfo) error {
	meta := volume.Meta
	rp, err := filepath.Rel(baetyl.DefaultDBDir, volume.Path)
	if err != nil {
		return fmt.Errorf("illegal path of volume %s", volume.Name)
	}
	containerDir := path.Join(rootDir, rp)
	err = os.MkdirAll(containerDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to prepare volume directory (%s): %s", volume.Path, err.Error())
	}
	volumeFile := path.Join(containerDir, name)
	if utils.FileExists(volumeFile) {
		md5, err := utils.CalculateFileMD5(volumeFile)
		if err == nil && md5 == meta.MD5 {
			a.ctx.Log().Debugf("volume (%s) exists", name)
			return nil
		}
	}
	res, err := a.http.SendUrl("GET", meta.URL, nil, nil)
	if err != nil || res == nil {
		// retry
		time.Sleep(time.Second)
		res, err = a.http.SendUrl("GET", meta.URL, nil, nil)
		if err != nil || res == nil {
			return fmt.Errorf("failed to download volume (%s): %v", name, err)
		}
	}
	err = utils.WriteFile(volumeFile, res)
	if err != nil {
		os.RemoveAll(containerDir)
		return fmt.Errorf("failed to prepare volume (%s): %s", name, err.Error())
	}
	md5, err := utils.CalculateFileMD5(volumeFile)
	if err != nil {
		os.RemoveAll(containerDir)
		return fmt.Errorf("failed to calculate MD5 of volume (%s): %s", name, err.Error())
	}
	if md5 != meta.MD5 {
		os.RemoveAll(containerDir)
		return fmt.Errorf("MD5 of volume (%s) invalid", name)
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

func generateRequest(snapshot config.Snapshot) ([]byte, error) {
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
	res, err := json.Marshal(reqs)
	if err != nil {
		return nil, err
	}
	return res, nil
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

func (a *agent) checkVolumeExists(volume baetyl.VolumeInfo) bool {
	rp, err := filepath.Rel(baetyl.DefaultDBDir, volume.Path)
	if err != nil {
		a.ctx.Log().Warnf("illegal path of volume: %s", volume.Name)
		return false
	}
	volumePath := path.Join(baetyl.DefaultDBDir, "volumes", rp)
	if !utils.DirExists(volumePath) {
		return false
	}
	if volume.Meta.MD5 != "" {
		volumeFile := path.Join(volumePath, volume.Name)
		md5, err := utils.CalculateFileMD5(volumeFile)
		if err != nil {
			a.ctx.Log().Warnf("failed to calculate md5 of volume file %s", volumeFile)
			return false
		}
		return md5 == volume.Meta.MD5
	}
	return true
}

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
