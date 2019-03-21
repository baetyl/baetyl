package master

import (
	"fmt"
	"os"
	"path"

	"github.com/baidu/openedge/logger"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
)

var appDir = path.Join("var", "db", "openedge")
var appConfigFile = path.Join(appDir, "application.yml")
var appBackupFile = path.Join(appDir, "application.yml.old")

// UpdateSystem updates system
func (m *Master) UpdateSystem(dir string, clean bool) error {
	err := m.update(dir, clean)
	if err != nil {
		err := fmt.Errorf("failed to update system: %s", err.Error())
		m.log.Errorf(err.Error())
		m.context.Set("error", err.Error())
		return err
	}
	m.context.Remove("error")
	return nil
}

func (m *Master) update(dir string, clean bool) error {
	m.log.Infof("system is updating")

	// backup application.yml
	err := m.backup()
	if err != nil {
		return err
	}
	defer m.clean()

	// copy new config into application.yml
	err = m.copy(dir)
	if err != nil {
		m.rollback()
		return err
	}

	// prepare services
	rvs, err := m.prepareServices()
	if err != nil {
		m.rollback()
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
	m.log.Infof("system is updated")
	if clean {
		err = os.RemoveAll(dir)
		if err != nil {
			m.log.Warnf("failed to remove app config dir (%s)", dir)
		}
		for _, v := range rvs {
			err = os.RemoveAll(v.Path)
			if err != nil {
				m.log.Warnf("failed to remove old volume (%s:%s)", v.Name, v.Path)
			}
		}
		m.log.Infof("old volumes are removed")
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

func (m *Master) copy(dir string) error {
	return utils.CopyFile(path.Join(dir, openedge.AppConfFileName), appConfigFile)
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
