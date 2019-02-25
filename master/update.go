package master

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/baidu/openedge/utils"
	"gopkg.in/yaml.v2"
)

var appDir = path.Join("var", "db", "openedge")
var appConfigFile = path.Join(appDir, "application.yml")
var appBackupFile = path.Join(appDir, "application.yml.old")

// Update updates system
func (m *Master) Update(cfg *openedge.AppConfig) error {
	if cfg == nil {
		err := fmt.Errorf("failed to update system: application config is null")
		m.log.Errorf(err.Error())
		m.context.Set("error", err.Error())
		return err
	}
	err := m.update(cfg)
	if err != nil {
		err := fmt.Errorf("failed to update system: %s", err.Error())
		m.log.Errorf(err.Error())
		m.context.Set("error", err.Error())
		return err
	}
	m.context.Remove("error")
	return nil
}

func (m *Master) update(cfg *openedge.AppConfig) error {
	// backup old config
	err := m.backup()
	if err != nil {
		return err
	}
	defer m.clean()

	// prepare new config
	err = m.prepare(cfg)
	if err != nil {
		return err
	}
	// stop all old services
	m.stopAllServices()
	// start all new services
	err = m.startAllServices()
	if err != nil {
		m.log.Infof("failed to start all new services, to rollback")
		err1 := m.rollback()
		if err1 != nil {
			return fmt.Errorf("%s; failed to rollback: %s", err.Error(), err1.Error())
		}
		// stop all new services
		m.stopAllServices()
		// start all old services
		err1 = m.startAllServices()
		if err1 != nil {
			return fmt.Errorf("%s; failed to rollback: %s", err.Error(), err1.Error())
		}
		return err
	}
	return nil
}

func (m *Master) backup() error {
	if !utils.FileExists(appConfigFile) {
		return nil
	}
	return os.Rename(appConfigFile, appBackupFile)
}

func (m *Master) rollback() error {
	if !utils.FileExists(appBackupFile) {
		return os.RemoveAll(appConfigFile)
	}
	return os.Rename(appBackupFile, appConfigFile)
}

func (m *Master) prepare(cfg *openedge.AppConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(appConfigFile, data, 0755)
}

func (m *Master) load() error {
	if !utils.FileExists(appConfigFile) {
		m.appcfg = openedge.AppConfig{}
		return nil
	}

	var cfg openedge.AppConfig
	err := utils.LoadYAML(appConfigFile, &cfg)
	if err != nil {
		return err
	}
	m.appcfg = cfg
	return nil
}

func (m *Master) clean() {
	err := os.RemoveAll(appBackupFile)
	if err != nil {
		logger.WithError(err).Errorf("failed to remove backup file")
	}
}
