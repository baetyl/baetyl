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
var appConfigFile = path.Join(appDir, openedge.AppConfFileName)
var appBackupFile = path.Join(appDir, openedge.AppBackupFileName)

// UpdateSystem updates system
func (m *Master) UpdateSystem(dir string, clean bool) error {
	err := m.update(dir, clean, true)
	if err != nil {
		err = fmt.Errorf("failed to update system: %s", err.Error())
		m.log.Errorf(err.Error())
	}
	m.infostats.updateError(err)
	return err
}

func (m *Master) update(dir string, clean, reload bool) error {
	m.log.Infof("system is updating")

	if reload {
		// backup application.yml
		err := m.backup()
		if err != nil {
			return err
		}

		// copy new config into application.yml
		err = m.copy(dir)
		if err != nil {
			m.rollback()
			return err
		}
	}

	defer m.clean()

	// prepare services
	rvs, updatedServices, removedServices, err := m.prepareServices()
	if err != nil {
		m.rollback()
		return err
	}

	// stop all removed services and updated services if it's reload
	if reload {
		m.stopAllServices(removedServices)
	}
	// start all updated services and new services
	err = m.startAllServices(updatedServices)

	if err != nil {
		m.log.Infof("failed to start all new services, to rollback")
		err1 := m.rollback()
		if err1 != nil {
			return fmt.Errorf("%s; failed to rollback: %s", err.Error(), err1.Error())
		}
		_, _, _, err2 := m.prepareServices()
		if err2 != nil {
			return fmt.Errorf("%s; failed to rollback, cannot reset configuration.", err2.Error())
		}
		// stop all updated services and new services
		m.stopAllServices(updatedServices)
		// start all removed services and updated service
		err1 = m.startAllServices(removedServices)
		if err1 != nil {
			return fmt.Errorf("%s; failed to rollback: %s", err.Error(), err1.Error())
		}
		return err
	}
	if clean {
		if os.RemoveAll(dir) != nil {
			m.log.Warnf("failed to remove app config dir (%s)", dir)
		}
		for _, v := range rvs {
			if os.RemoveAll(v.Path) != nil {
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
