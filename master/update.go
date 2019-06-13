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
func (m *Master) UpdateSystem(dir string) error {
	err := m.update(dir)
	if err != nil {
		err = fmt.Errorf("failed to update system: %s", err.Error())
		m.log.Errorf(err.Error())
	}
	m.infostats.updateError(err)
	return err
}

func (m *Master) update(dir string) error {
	m.log.Infof("system is updating")
	defer m.log.Infof("system is updated")

	cur, old, err := m.reload(dir)
	if err != nil {
		m.rollback()
		return err
	}

	// prepare services
	keepServices := diffServices(cur, old)
	m.engine.Prepare(cur.Services)

	// stop all removed or updated services
	m.stopServices(keepServices)
	// start all updated or added services
	err = m.startServices(cur)
	if err != nil {
		m.log.Infof("failed to start new services, to rollback")
		rberr := m.rollback()
		if rberr != nil {
			return fmt.Errorf("%s; failed to rollback: %s", err.Error(), rberr.Error())
		}
		// stop all updated or added services
		m.stopServices(keepServices)
		// start all removed or updated services
		rberr = m.startServices(old)
		if rberr != nil {
			return fmt.Errorf("%s; failed to rollback: %s", err.Error(), rberr.Error())
		}
		m.commit(old.Version)
		return err
	}
	m.commit(cur.Version)
	return nil
}

func (m *Master) reload(dir string) (cur, old openedge.AppConfig, err error) {
	if dir != "" {
		// backup
		if utils.FileExists(appConfigFile) {
			// application.yml --> application.yml.old
			err = os.Rename(appConfigFile, appBackupFile)
			if err != nil {
				return
			}
		} else {
			// none --> application.yml.old (empty)
			var f *os.File
			f, err = os.Create(appBackupFile)
			if err != nil {
				return
			}
			f.Close()
		}

		// copy {dir}/application.yml to application.yml
		err = utils.CopyFile(path.Join(dir, openedge.AppConfFileName), appConfigFile)
		if err != nil {
			return
		}
	}
	if utils.FileExists(appConfigFile) {
		err = utils.LoadYAML(appConfigFile, &cur)
		if err != nil {
			return
		}
	}
	if utils.FileExists(appBackupFile) {
		err = utils.LoadYAML(appBackupFile, &old)
		if err != nil {
			return
		}
	}
	return
}

func (m *Master) rollback() error {
	if !utils.FileExists(appBackupFile) {
		return nil
	}
	// application.yml.old --> application.yml
	return os.Rename(appBackupFile, appConfigFile)
}

func (m *Master) commit(ver string) {
	// update config version
	m.infostats.updateVersion(ver)
	// remove application.yml.old
	err := os.RemoveAll(appBackupFile)
	if err != nil {
		logger.WithError(err).Errorf("failed to remove backup file (application.yml.old)")
	}
}
