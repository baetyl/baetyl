package native

import (
	"os"
	"path"
	"sync"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/orcaman/concurrent-map"
)

const packageConfigPath = "package.yml"

type packageConfig struct {
	Entry string `yaml:"entry" json:"entry"`
}

type nativeService struct {
	info      engine.ServiceInfo
	cfgs      processConfigs
	engine    *nativeEngine
	instances cmap.ConcurrentMap
	wdir      string
	log       logger.Logger
}

func (s *nativeService) Name() string {
	return s.info.Name
}

func (s *nativeService) Stats() openedge.ServiceStatus {
	return openedge.ServiceStatus{}
}

func (s *nativeService) Stop() {
	defer os.RemoveAll(s.cfgs.pwd)
	var wg sync.WaitGroup
	for _, v := range s.instances.Items() {
		wg.Add(1)
		go func(i *nativeInstance, wg *sync.WaitGroup) {
			defer wg.Done()
			i.Close()
		}(v.(*nativeInstance), &wg)
	}
	wg.Wait()
}

func (s *nativeService) start() error {
	mounts := make([]engine.MountInfo, 0)
	datasetsDir := path.Join("var", "db", "openedge", "datasets")
	for _, m := range s.info.Datasets {
		v := path.Join(datasetsDir, m.Name, m.Version)
		mounts = append(mounts, engine.MountInfo{Volume: v, Target: m.Target, ReadOnly: m.ReadOnly})
	}
	for _, m := range s.info.Volumes {
		mounts = append(mounts, m)
	}

	for _, m := range mounts {
		src := path.Join(s.engine.pwd, m.Volume)
		err := os.MkdirAll(src, 0755)
		if err != nil {
			return err
		}
		dst := path.Join(s.cfgs.pwd, m.Target)
		err = os.MkdirAll(path.Dir(dst), 0755)
		if err != nil {
			return err
		}
		err = os.Symlink(src, dst)
		if err != nil {
			return err
		}
	}
	return s.startInstance()
}
