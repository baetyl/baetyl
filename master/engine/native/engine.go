package native

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
	"github.com/orcaman/concurrent-map"
)

// NAME of this engine
const NAME = "native"

func init() {
	engine.Factories()[NAME] = New
}

// New native engine
func New(grace time.Duration, pwd string) (engine.Engine, error) {
	e := &nativeEngine{
		pwd:   pwd,
		grace: grace,
		log:   logger.WithField("engine", NAME),
	}
	e.initHistory()
	return e, nil
}

type nativeEngine struct {
	pwd   string // work directory
	grace time.Duration
	log   logger.Logger
	mut   sync.Mutex
}

// Name of engine
func (e *nativeEngine) Name() string {
	return NAME
}

// Close engine
func (e *nativeEngine) Close() error {
	e.cleanHistory()
	return nil
}

// Prepare prepares all images
func (e *nativeEngine) Prepare([]openedge.ServiceInfo) {
	// do nothing in native mode
}

// Run new service
func (e *nativeEngine) Run(cfg openedge.ServiceInfo, vs map[string]openedge.VolumeInfo) (engine.Service, error) {
	spwd := path.Join(e.pwd, "var", "run", "openedge", "services", cfg.Name)
	err := mount(e.pwd, spwd, cfg.Mounts, vs)
	if err != nil {
		os.RemoveAll(spwd)
		return nil, err
	}
	var pkg packageConfig
	image := strings.Replace(strings.TrimSpace(cfg.Image), ":", "/", -1)
	pkgDir := path.Join(spwd, "lib", "openedge", image)
	err = utils.LoadYAML(path.Join(pkgDir, packageConfigPath), &pkg)
	if err != nil {
		os.RemoveAll(spwd)
		return nil, err
	}
	argv := make([]string, 0)
	argv = append(argv, cfg.Args...)
	params := processConfigs{
		exec: path.Join(pkgDir, pkg.Entry),
		argv: argv,
		env:  utils.AppendEnv(cfg.Env, true),
		pwd:  spwd,
	}
	s := &nativeService{
		cfg:       cfg,
		engine:    e,
		params:    params,
		instances: cmap.New(),
		log:       e.log.WithField("service", cfg.Name),
	}
	err = s.Start()
	if err != nil {
		s.Stop()
		return nil, err
	}
	return s, nil
}

func mount(epwd, spwd string, ms []openedge.MountInfo, vs map[string]openedge.VolumeInfo) error {
	for _, m := range ms {
		v, ok := vs[m.Name]
		if !ok {
			return fmt.Errorf("volume '%s' not found", m.Name)
		}
		src := path.Join(epwd, path.Clean(v.Path))
		err := os.MkdirAll(src, 0755)
		if err != nil {
			return err
		}
		dst := path.Join(spwd, path.Clean(strings.TrimSpace(m.Path)))
		err = os.MkdirAll(path.Dir(dst), 0755)
		if err != nil {
			return err
		}
		err = os.RemoveAll(dst)
		if err != nil {
			return err
		}
		err = os.Symlink(src, dst)
		if err != nil {
			return err
		}
	}
	return nil
}
