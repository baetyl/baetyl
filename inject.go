package baetyl

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/protocol/mqtt"
	schema "github.com/baetyl/baetyl/schema/v3"

	"gopkg.in/yaml.v2"
)

const injectHostDir = "inject"
const injectVolumeDir = "volumes"

const (
	injectConfigDirOnPosix   = "/etc/baetyl"
	injectConfigDirOnWindows = "C:\\ProgramData\\baetyl\\config"
	injectConfigPath         = "service.yml"

	injectAPIOnPosix    = "/var/run/baetyl.sock"
	injectAPIOnWindows  = "C:\\ProgramData\\baetyl\\runtime\\baetyl.sock"
	injectLAPIOnPosix   = "/var/run/baetyl-legacy.sock"
	injectLAPIOnWindows = "C:\\ProgramData\\baetyl\\runtime\\baetyl-legacy.sock"

	injectLogDirOnPosix   = "/var/log/baetyl"
	injectLogDirOnWindows = "C:\\ProgramData\\baetyl\\log"
)

type apiConfig struct {
	Address          string `yaml:"address" json:"address"`
	TimeoutInSeconds int    `yaml:"timeout_s" json:"timeout_s"`
}

type injectConfig struct {
	Name            string          `yaml:"name" json:"name"`
	Logger          logger.LogInfo  `yaml:"logger" json:"logger"`
	Certificate     string          `yaml:"certificate" json:"certificate"`
	APIServer       apiConfig       `yaml:"apiserver" json:"apiserver"`
	LegacyAPIServer apiConfig       `yaml:"legacy_apiserver" json:"legacy_apiserver"`
	Hub             mqtt.ClientInfo `yaml:"hub" json:"hub"`
}

type injectEnv struct {
	hostRoot        string
	hostVolumes     string
	injectConfigDir string
	injectAPI       string
	injectLAPI      string
	injectLogDir    string
}

func (rt *runtime) injectAppConfig(ctx context.Context, appcfg *schema.ComposeAppConfig) error {
	iroot := filepath.Join(rt.cfg.DataPath, injectHostDir)
	var err error
	os.RemoveAll(iroot) // clean inject root
	err = os.MkdirAll(iroot, 0755)
	if err != nil {
		return err
	}
	ie := injectEnv{
		hostRoot:        iroot,
		hostVolumes:     filepath.Join(rt.cfg.DataPath, injectVolumeDir, appcfg.AppVersion),
		injectConfigDir: injectConfigDirOnPosix,
		injectAPI:       injectAPIOnPosix,
		injectLAPI:      injectLAPIOnPosix,
		injectLogDir:    injectLogDirOnPosix,
	}
	if rt.e.OSType() == "windows" {
		ie.injectConfigDir = injectConfigDirOnWindows
		ie.injectAPI = injectAPIOnWindows
		ie.injectLAPI = injectLAPIOnWindows
		ie.injectLogDir = injectLogDirOnWindows
	}
	for name := range appcfg.Services {
		svc := appcfg.Services[name]
		err = rt.injectService(ctx, name, &svc, &ie)
		if err != nil {
			return err
		}
		appcfg.Services[name] = svc
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}

func (rt *runtime) injectService(ctx context.Context, name string, svc *schema.ComposeService, ie *injectEnv) error {
	for _, v := range svc.Volumes {
		v.Source = filepath.Join(ie.hostVolumes, v.Source)
	}
	var err error
	iConfPath := filepath.Join(ie.hostRoot, fmt.Sprintf("%s-conf", name))
	err = os.MkdirAll(iConfPath, 0755)
	if err != nil {
		return err
	}
	err = rt.genServiceCert(filepath.Join(iConfPath, "service.cert"), name)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cfg := injectConfig{
		Name: name,
		Logger: logger.LogInfo{
			Path:  filepath.Join(ie.injectLogDir, "service.log"),
			Level: "info",
		},
		Certificate: "service.cert",
		APIServer: apiConfig{
			Address:          ie.injectAPI,
			TimeoutInSeconds: 60,
		},
		LegacyAPIServer: apiConfig{
			Address:          ie.injectLAPI,
			TimeoutInSeconds: 60,
		},
		Hub: mqtt.ClientInfo{}, // FIXME
	}
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(
		filepath.Join(iConfPath, injectConfigPath),
		[]byte(data),
		0644,
	)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	svc.Volumes = append(svc.Volumes, &schema.ServiceVolume{
		Type:     "bind",
		Source:   iConfPath,
		Target:   ie.injectConfigDir,
		ReadOnly: true,
	})
	svc.Volumes = append(svc.Volumes, &schema.ServiceVolume{
		Type:     "bind",
		Source:   rt.cfg.APIConfig.Address,
		Target:   ie.injectAPI,
		ReadOnly: true,
	})
	svc.Volumes = append(svc.Volumes, &schema.ServiceVolume{
		Type:     "bind",
		Source:   rt.cfg.LegacyAPIConfig.Address,
		Target:   ie.injectLAPI,
		ReadOnly: true,
	})
	iLogPath := filepath.Join(ie.hostRoot, fmt.Sprintf("%s-log", name))
	err = os.MkdirAll(iLogPath, 0755)
	if err != nil {
		return err
	}
	svc.Volumes = append(svc.Volumes, &schema.ServiceVolume{
		Type:     "bind",
		Source:   iLogPath,
		Target:   ie.injectLogDir,
		ReadOnly: true,
	})
	rt.log.Infof("inject service (%s) done", name)
	return nil
}
