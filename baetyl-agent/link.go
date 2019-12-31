package main

import (
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl/baetyl-agent/common"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

func (a *agent) processLinkEvent(e *Event) {
	le := e.Content.(*EventLink)
	a.ctx.Log().Infof("process ota: type=%s, trace=%s", le.Type, le.Trace)
	ol := newOTALog(a.cfg.OTA, a, &EventOTA{Type: le.Type, Trace: le.Trace}, a.ctx.Log().WithField("agent", "otalog"))
	defer ol.wait()
	err := a.processLinkOTA(le)
	if err != nil {
		a.ctx.Log().WithError(err).Warnf("failed to process ota event")
		ol.write(baetyl.OTAFailure, "failed to process link ota event", err)
	}
}

func (a *agent) processLinkOTA(le *EventLink) error {
	d, ok := le.Info["deployment"]
	if !ok {
		return fmt.Errorf("no deployment info in delta info")
	}
	var deploy deployment
	for k, v := range d.(map[string]interface{}) {
		deploy.Name = k
		deploy.AppVersion = v.(string)
	}

	req := a.newRequest("deployment", deploy.Name)
	resData, err := a.sendData(*req)
	if err != nil {
		return fmt.Errorf("failed to send request by link: %s", err.Error())
	}
	var res BackwardInfo
	err = json.Unmarshal(resData, &res)
	var dep Deployment
	err = mapstructure.Decode(res.Response["deployment"], &dep)
	if err != nil {
		return fmt.Errorf("error to transform from map to deployment: %s", err.Error())
	}
	metadata, hostDir := a.processApplication(dep.Snapshot.Apps)

	// avoid duplicated resource synchronization
	var volumes []baetyl.VolumeInfo
	for name, volume := range metadata {
		volumes = append(volumes, volume)
		metaVersion := volume.Meta.Version
		if metaVersion == "" {
			delete(metadata, name)
		}
		if a.checkVolumeExists(volume) {
			delete(metadata, name)
		}
	}
	a.cleaner.set(deploy.AppVersion, volumes)
	err = a.processVolumes(metadata)
	if err != nil {
		return err
	}
	err = a.ctx.UpdateSystem(le.Trace, le.Type, hostDir)
	if err != nil {
		return fmt.Errorf("failed to update system: %s", err.Error())
	}
	return nil
}

func (a *agent) processVolumes(volumes map[string]baetyl.VolumeInfo) error {
	rootDir := path.Join(baetyl.DefaultDBDir, "volumes")
	for name, volume := range volumes {
		if meta := volume.Meta; meta.Version != "" {
			if volume.Meta.URL != "" {
				err := a.processURL(rootDir, name, volume)
				if err != nil {
					a.ctx.Log().Errorf("download volume (%s) failed: %s", name, err.Error())
				}
			} else {
				req := a.newRequest("config", name)
				resData, err := a.sendData(*req)
				if err != nil {
					return fmt.Errorf("failed to send request by link: %s", err.Error())
				}
				var res BackwardInfo
				err = json.Unmarshal(resData, &res)
				var config ModuleConfig
				err = mapstructure.Decode(res.Response["config"], &config)
				if err != nil {
					return fmt.Errorf("error to transform from map to config: %s", err.Error())
				}
				err = a.processModuleConfig(rootDir, volume.Path, &config)
				if err != nil {
					a.ctx.Log().Errorf("download module config (%s) failed: %s", name, err.Error())
				}
			}
		}
	}
	return nil
}

func (a *agent) processModuleConfig(rootDir string, volumePath string, config *ModuleConfig) error {
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

func (a *agent) processApplication(appInfo map[string]string) (map[string]baetyl.VolumeInfo, string) {
	var appName string
	for k, _ := range appInfo {
		appName = k
	}
	req := a.newRequest("application", appName)
	resData, err := a.sendData(*req)
	if err != nil {
		a.ctx.Log().WithError(err).Warnf("failed to get response by link")
		return nil, ""
	}
	var res BackwardInfo
	err = json.Unmarshal(resData, &res)
	deployConfig, err := getDeployConfig(&res)
	if err != nil {
		a.ctx.Log().WithError(err).Warnf("failed to get deploy config")
		return nil, ""
	}
	appDir := deployConfig.AppConfig.Name
	hostDir := path.Join(baetyl.DefaultDBDir, appDir, deployConfig.AppConfig.AppVersion)
	containerDir := path.Join(baetyl.DefaultDBDir, "volumes", appDir, deployConfig.AppConfig.AppVersion)

	err = os.MkdirAll(containerDir, 0755)
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
	err = ioutil.WriteFile(path.Join(containerDir, baetyl.AppConfFileName), data, 0755)
	if err != nil {
		os.RemoveAll(containerDir)
		a.ctx.Log().WithError(err).Warnf("failed to write applicationt config into file (%s): %s", baetyl.AppConfFileName, err.Error())
		return nil, ""
	}
	return deployConfig.Metadata, hostDir
}

func getDeployConfig(info *BackwardInfo) (*DeployConfig, error) {
	appData, err := json.Marshal(info.Response["application"])
	if err != nil {
		return nil, err
	}
	var deployConfig DeployConfig
	err = json.Unmarshal(appData, &deployConfig)
	if err != nil {
		return nil, err
	}
	return &deployConfig, nil
}

func (a *agent) newRequest(resourceType, resourceName string) *ForwardInfo {
	return &ForwardInfo{
		Namespace: a.node.Namespace,
		Name:      a.node.Name,
		Request: map[string]string{
			common.ResourceType: resourceType,
			common.ResourceName: resourceName,
		},
	}
}

func (a *agent) sendData(request interface{}) ([]byte, error) {
	content, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	msg := &link.Message{
		Content: content,
	}
	resMsg, err := a.link.Call(msg)
	if err != nil {
		return nil, err
	}
	return resMsg.Content, nil
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
