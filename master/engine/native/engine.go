package native

import (
	"path"
	"time"

	"github.com/orcaman/concurrent-map"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/utils"
)

// NAME of this engine
const NAME = "native"

func init() {
	engine.Factories()[NAME] = New
}

// New native engine
func New(grace time.Duration, pwd string) (engine.Engine, error) {
	return &nativeEngine{
		pwd:   pwd,
		grace: grace,
		log:   logger.WithField("engine", NAME),
	}, nil
}

type nativeEngine struct {
	pwd   string // work directory
	grace time.Duration
	log   logger.Logger
}

// Name of engine
func (e *nativeEngine) Name() string {
	return NAME
}

// Close engine
func (e *nativeEngine) Close() error {
	return nil
}

// Run new service
func (e *nativeEngine) Run(cfg engine.ServiceInfo) (engine.Service, error) {
	var pkg packageConfig
	pkgDir := path.Join(e.pwd, "lib", "openedge", "packages", cfg.Image)
	err := utils.LoadYAML(path.Join(pkgDir, packageConfigPath), &pkg)
	if err != nil {
		return nil, err
	}
	argv := make([]string, 0)
	argv = append(argv, cfg.Name) // add prefix "openedge-service-"?
	for _, p := range cfg.Params {
		argv = append(argv, p)
	}
	cfgs := processConfigs{
		exec: path.Join(pkgDir, pkg.Entry),
		argv: argv,
		env:  utils.AppendEnv(cfg.Env, true),
		pwd:  path.Join(e.pwd, "var", "run", "openedge", "services", cfg.Name),
	}
	s := &nativeService{
		info:      cfg,
		engine:    e,
		cfgs:      cfgs,
		instances: cmap.New(),
		log:       e.log.WithField("service", cfg.Name),
	}
	err = s.start()
	if err != nil {
		s.Stop()
		return nil, err
	}
	return s, nil
}
