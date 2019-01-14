package master

import (
	"fmt"
	"os"
	"path"

	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/logger"
	"github.com/baidu/openedge/module/utils"
	"github.com/mholt/archiver"
)

// backupFile backup file name
const backupFile = "module.zip"

// backupDir dir to backup
var backupDir = path.Join("var", "db", "openedge", "module")

// confFile config file path
var confFile = path.Join(backupDir, "module.yml")

func (m *Master) reload(file string) error {
	err := m.backupDir()
	if err != nil {
		return fmt.Errorf("failed to backup old config: %s", err.Error())
	}
	defer m.cleanBackupFile()
	err = m.unpackConfigFile(file)
	if err != nil {
		return fmt.Errorf("failed to unpack new config: %s", err.Error())
	}
	err = m.loadConfig()
	if err != nil {
		logger.Log.WithError(err).Infof("failed to load new config, rollback")
		err1 := m.unpackBackupFile()
		if err1 != nil {
			err = fmt.Errorf(err.Error() + ";failed to unpack old config backup file: " + err1.Error())
			return err
		}
		err1 = m.loadConfig()
		if err1 != nil {
			err = fmt.Errorf(err.Error() + ";failed to load old config: " + err1.Error())
			return err
		}
		return fmt.Errorf("failed to load new config: %s", err.Error())
	}
	m.engine.StopAll()
	err = m.engine.StartAll(m.conf.Modules)
	if err != nil {
		logger.Log.WithError(err).Infof("failed to load new config, rollback")
		err1 := m.unpackBackupFile()
		if err1 != nil {
			err = fmt.Errorf(err.Error() + ";failed to unpack old config backup file" + err1.Error())
			return err
		}
		err1 = m.loadConfig()
		if err1 != nil {
			err = fmt.Errorf(err.Error() + ";failed to load old config" + err1.Error())
			return err
		}
		m.engine.StopAll()
		err1 = m.engine.StartAll(m.conf.Modules)
		if err1 != nil {
			err = fmt.Errorf(err.Error() + ";failed to start modules with old config" + err.Error())
			return err
		}
	}
	return nil
}

func (m *Master) backupDir() error {
	if !utils.DirExists(backupDir) {
		return nil
	}
	return archiver.Zip.Make(backupFile, []string{backupDir})
}

func (m *Master) cleanBackupFile() {
	err := os.RemoveAll(backupFile)
	if err != nil {
		logger.Log.WithError(err).Errorf("failed to remove backup file")
	}
}

func (m *Master) unpackConfigFile(file string) error {
	err := archiver.Zip.Open(file, m.context.PWD)
	return err
}

func (m *Master) unpackBackupFile() error {
	if !utils.FileExists(backupFile) {
		return os.RemoveAll(backupDir)
	}
	err := archiver.Zip.Open(backupFile, path.Dir(backupDir))
	return err
}

func (m *Master) loadConfig() error {
	if !utils.FileExists(confFile) {
		m.conf.Version = ""
		m.conf.Modules = []config.Module{}
		return nil
	}

	return module.Load(&m.conf, confFile)
}
