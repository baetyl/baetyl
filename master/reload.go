package master

import (
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/utils"
	"github.com/mholt/archiver"
	cmap "github.com/orcaman/concurrent-map"
)

var serviceDir = path.Join("var", "db", "openedge", "service")
var serviceOldDir = path.Join("var", "db", "openedge", "service-old")
var configFile = path.Join(serviceDir, "config.yml")

func (m *Master) startServices() error {
	m.cleanServices()
	for k, v := range m.dyncfg.Services {
		s, err := m.engine.Run(k, v)
		if err != nil {
			return err
		}
		m.services.Set(k, s)
	}
	return nil
}

func (m *Master) cleanServices() {
	svcs := m.services
	m.services = cmap.New()
	var wg sync.WaitGroup
	for _, s := range svcs.Items() {
		wg.Add(1)
		go func(s engine.Service, wg *sync.WaitGroup) {
			defer wg.Done()
			s.Stop(m.dyncfg.Grace)
		}(s.(engine.Service), &wg)
	}
	wg.Wait()
}

func (m *Master) reload(file string) error {
	if !utils.FileExists(file) {
		return fmt.Errorf("no file: %s", file)
	}
	err := m.backup()
	if err != nil {
		return fmt.Errorf("failed to backup old service directory: %s", err.Error())
	}
	defer m.clean(serviceOldDir)
	err = m.unpack(file)
	if err != nil {
		err1 := m.rollback()
		if err1 != nil {
			return fmt.Errorf("%s ;failed to rollback old service config: %s", err.Error(), err1.Error())
		}
		return fmt.Errorf("failed to unpack new service config: %s", err.Error())
	}
	err = m.load()
	if err != nil {
		m.log.WithError(err).Infof("failed to load new service config, rollback")
		err1 := m.rollback()
		if err1 != nil {
			return fmt.Errorf("%s ;failed to rollback old service config: %s", err.Error(), err1.Error())
		}
		err1 = m.load()
		if err1 != nil {
			return fmt.Errorf("%s ;failed to load old service config: %s", err.Error(), err1.Error())
		}
		return fmt.Errorf("failed to load new service config: %s", err.Error())
	}
	err = m.startServices()
	if err != nil {
		m.log.WithError(err).Infof("failed to load new service config, rollback")
		err1 := m.rollback()
		if err1 != nil {
			return fmt.Errorf("%s ;failed to rollback old service config: %s", err.Error(), err1.Error())
		}
		err1 = m.load()
		if err1 != nil {
			return fmt.Errorf("%s ;failed to load old service config: %s", err.Error(), err1.Error())
		}
		err = m.startServices()
		if err1 != nil {
			err = fmt.Errorf(err.Error() + ";failed to start modules with old config" + err.Error())
			return err
		}
	}
	return nil
}

func (m *Master) backup() error {
	m.clean(serviceOldDir)
	if !utils.DirExists(serviceDir) {
		return nil
	}
	return os.Rename(serviceDir, serviceOldDir)
}

func (m *Master) rollback() error {
	m.clean(serviceDir)
	if !utils.DirExists(serviceOldDir) {
		return nil
	}
	return os.Rename(serviceOldDir, serviceDir)
}

func (m *Master) clean(dir string) {
	err := os.RemoveAll(dir)
	if err != nil && !os.IsNotExist(err) {
		m.log.WithError(err).Warnf("failed to remove directory: %s", dir)
	}
}

func (m *Master) unpack(file string) error {
	return archiver.Zip.Open(file, m.wdir)
}

func (m *Master) load() error {
	m.dyncfg = DynamicConfig{}
	if !utils.FileExists(configFile) {
		return nil
	}
	return utils.LoadYAML(configFile, &m.dyncfg)
}
