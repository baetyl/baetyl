package master

import (
	"io/ioutil"
	"path"
	"sync"

	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/utils"
	cmap "github.com/orcaman/concurrent-map"
)

// Reload services
func (m *Master) Reload(dir string) error {
	data, err := ioutil.ReadFile(path.Join(dir, "config.yml"))
	if err != nil {
		return err
	}
	dyncfg := &DynamicConfig{}
	err = utils.UnmarshalYAML(data, dyncfg)
	if err != nil {
		return err
	}
	// Very simple policy, stop all service and start new
	if m.dyncfg != nil {
		m.cleanServices()
	}
	m.dyncfg = dyncfg
	for k, v := range dyncfg.Services {
		s, err := m.engine.Run(k, v)
		if err != nil {
			return err
		}
		m.svcs.Set(k, s)
	}
	return nil
}

func (m *Master) cleanServices() {
	svcs := m.svcs
	m.svcs = cmap.New()
	var wg sync.WaitGroup
	for _, s := range svcs.Items() {
		wg.Add(1)
		go func(s engine.Service, wg *sync.WaitGroup) {
			s.Stop(m.dyncfg.Grace)
			wg.Done()
		}(s.(engine.Service), &wg)
	}
	wg.Wait()
}

/*
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
		openedge.WithError(err).Infof("failed to load new config, rollback")
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
	err = m.engine.StartAll(m.cfg.Modules)
	if err != nil {
		openedge.WithError(err).Infof("failed to load new config, rollback")
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
		err1 = m.engine.StartAll(m.cfg.Modules)
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
		openedge.WithError(err).Errorf("failed to remove backup file")
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

	data, err := ioutil.ReadFile(confFile)
	if err != nil {
		return err
	}
	return utils.UnmarshalYAML(data, &m.cfg)
}
*/
