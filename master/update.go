package master

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/baidu/openedge/logger"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
	"github.com/inconshreveable/go-update"
)

var appDir = path.Join("var", "db", "openedge")
var appConfigFile = path.Join(appDir, openedge.AppConfFileName)
var appBackupFile = path.Join(appDir, openedge.AppBackupFileName)

// UpdateSystem updates application or master
func (m *Master) UpdateSystem(trace, tp, target string) (err error) {
	switch tp {
	case openedge.OTAMST:
		err = m.UpdateMST(trace, target, openedge.DefaultBinBackupFile)
	default:
		err = m.UpdateAPP(trace, target)
	}
	if err != nil {
		err = fmt.Errorf("failed to update system: %s", err.Error())
		m.log.Errorf(err.Error())
	}
	m.infostats.setError(err)
	return err
}

// UpdateAPP updates application
func (m *Master) UpdateAPP(trace, target string) error {
	log := m.log
	isOTA := target != "" || utils.FileExists(m.cfg.OTALog.Path)
	if isOTA {
		log = logger.New(m.cfg.OTALog, openedge.OTAKeyTrace, trace, openedge.OTAKeyType, openedge.OTAAPP)
		log.WithField(openedge.OTAKeyStep, openedge.OTAUpdating).Infof("app is updating")
	}

	cur, old, err := m.loadAPPConfig(target)
	if err != nil {
		log.WithField(openedge.OTAKeyStep, openedge.OTARollingBack).WithError(err).Infof("failed to reload config")
		rberr := m.rollBackAPP()
		if rberr != nil {
			log.WithField(openedge.OTAKeyStep, openedge.OTAFailure).WithError(rberr).Infof("failed to roll back")
			return fmt.Errorf("failed to reload config: %s; failed to roll back: %s", err.Error(), rberr.Error())
		}
		log.WithField(openedge.OTAKeyStep, openedge.OTARolledBack).Infof("app is rolled back")
		return fmt.Errorf("failed to reload config: %s", err.Error())
	}

	// prepare services
	keepServices := diffServices(cur, old)
	m.engine.Prepare(cur.Services)

	// stop all removed or updated services
	m.stopServices(keepServices)
	// start all updated or added services
	err = m.startServices(cur)
	if err != nil {
		log.WithField(openedge.OTAKeyStep, openedge.OTARollingBack).WithError(err).Infof("failed to start app")
		rberr := m.rollBackAPP()
		if rberr != nil {
			log.WithField(openedge.OTAKeyStep, openedge.OTAFailure).WithError(rberr).Infof("failed to roll back")
			return fmt.Errorf("failed to start app: %s; failed to roll back: %s", err.Error(), rberr.Error())
		}
		// stop all updated or added services
		m.stopServices(keepServices)
		// start all removed or updated services
		rberr = m.startServices(old)
		if rberr != nil {
			log.WithField(openedge.OTAKeyStep, openedge.OTAFailure).WithError(rberr).Infof("failed to roll back")
			return fmt.Errorf("failed to restart old app: %s; failed to roll back: %s", err.Error(), rberr.Error())
		}
		m.commitAPP(old.Version)
		log.WithField(openedge.OTAKeyStep, openedge.OTARolledBack).Infof("app is rolled back")
		return fmt.Errorf("failed to start app: %s", err.Error())
	}
	m.commitAPP(cur.Version)
	if isOTA {
		log.WithField(openedge.OTAKeyStep, openedge.OTAUpdated).Infof("app is updated")
	}
	return nil
}

func (m *Master) loadAPPConfig(target string) (cur, old openedge.AppConfig, err error) {
	if target != "" {
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

		if utils.FileExists(target) {
			// copy {target} to application.yml
			err = utils.CopyFile(target, appConfigFile)
		} else {
			// copy {target}/application.yml to application.yml
			err = utils.CopyFile(path.Join(target, openedge.AppConfFileName), appConfigFile)
		}
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

func (m *Master) rollBackAPP() error {
	if !utils.FileExists(appBackupFile) {
		return nil
	}
	// application.yml.old --> application.yml
	return os.Rename(appBackupFile, appConfigFile)
}

func (m *Master) commitAPP(ver string) {
	defer m.log.Infof("app version (%s) committed", ver)

	// update config version
	m.infostats.setVersion(ver)
	// remove application.yml.old
	err := os.RemoveAll(appBackupFile)
	if err != nil {
		logger.WithError(err).Errorf("failed to remove backup file (%s)", appBackupFile)
	}
}

// UpdateMST updates master
func (m *Master) UpdateMST(trace, target, backup string) (err error) {
	log := logger.New(m.cfg.OTALog, openedge.OTAKeyTrace, trace, openedge.OTAKeyType, openedge.OTAMST)

	if err = m.check(target); err != nil {
		log.WithField(openedge.OTAKeyStep, openedge.OTAFailure).WithError(err).Infof("failed to check master")
		return fmt.Errorf("failed to check master: %s", err.Error())
	}

	log.WithField(openedge.OTAKeyStep, openedge.OTAUpdating).Infof("master is updating")
	if err = apply(target, backup); err != nil {
		log.WithField(openedge.OTAKeyStep, openedge.OTARollingBack).WithError(err).Infof("failed to apply master")
		rberr := RollBackMST()
		if rberr != nil {
			log.WithField(openedge.OTAKeyStep, openedge.OTAFailure).WithError(rberr).Infof("failed to roll back")
			return fmt.Errorf("failed to apply master: %s; failed to roll back: %s", err.Error(), rberr.Error())
		}
		log.WithField(openedge.OTAKeyStep, openedge.OTARolledBack).Infof("master is rolled back")
		return fmt.Errorf("failed to apply master: %s", err.Error())
	}

	log.WithField(openedge.OTAKeyStep, openedge.OTARestarting).Infof("master is restarting")
	return m.Close()
}

// RollBackMST rolls back master
func RollBackMST() error {
	if !utils.FileExists(openedge.DefaultBinBackupFile) {
		return nil
	}
	err := apply(openedge.DefaultBinBackupFile, "")
	if err != nil {
		logger.WithError(err).Errorf("failed to apply backup master")
	}
	err = os.RemoveAll(openedge.DefaultBinBackupFile)
	if err != nil {
		logger.WithError(err).Errorf("failed to remove backup file (%s)", openedge.DefaultBinBackupFile)
	}
	return nil
}

// CommitMST commits master
func CommitMST() bool {
	if !utils.FileExists(openedge.DefaultBinBackupFile) {
		return false
	}
	err := os.RemoveAll(openedge.DefaultBinBackupFile)
	if err != nil {
		logger.WithError(err).Errorf("failed to remove backup file (%s)", openedge.DefaultBinBackupFile)
	}
	return true
}

func apply(target, backup string) error {
	f, err := os.Open(target)
	if err != nil {
		return fmt.Errorf("failed to open binary: %s", err.Error())
	}
	defer f.Close()
	err = update.Apply(f, update.Options{OldSavePath: backup})
	if err != nil {
		return fmt.Errorf("failed to apply binary: %s", err.Error())
	}
	return nil
}

func (m *Master) check(target string) error {
	m.log.Debugf("new binary: %s", target)
	os.Chmod(target, 0755)
	cmd := exec.Command(target, "check", "-w", m.pwd, "-c", m.cfg.File)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("check result: %s", err.Error())
	}
	if !strings.Contains(string(out), openedge.CheckOK) {
		return fmt.Errorf("check result: OK expected, but get %s", string(out))
	}
	return nil
}
