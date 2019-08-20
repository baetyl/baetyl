package native

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/shirou/gopsutil/process"
)

// NAME of this engine
const NAME = "native"

func init() {
	engine.Factories()[NAME] = New
}

// New native engine
func New(grace time.Duration, pwd string, stats engine.InfoStats) (engine.Engine, error) {
	e := &nativeEngine{
		InfoStats: stats,
		pwd:       pwd,
		grace:     grace,
		log:       logger.WithField("engine", NAME),
	}
	return e, nil
}

type nativeEngine struct {
	engine.InfoStats
	pwd   string // work directory
	grace time.Duration
	log   logger.Logger
}

// Name of engine
func (e *nativeEngine) Name() string {
	return NAME
}

// Recover recover old services when master restart
func (e *nativeEngine) Recover() {
	// clean old services in native mode
	e.clean()
}

// Prepare prepares all images
func (e *nativeEngine) Prepare(openedge.AppConfig) {
	// do nothing in native mode
}

// Clean clean all old instances
func (e *nativeEngine) clean() {
	sss := map[string]map[string]attribute{}
	if e.LoadStats(&sss) {
		for sn, instances := range sss {
			for in, instance := range instances {
				id := int32(instance.Process.ID)
				if id == 0 {
					e.log.Warnf("[%s][%s] process id not found, maybe running mode changed", sn, in)
					continue
				}
				name := instance.Process.Name
				p, err := process.NewProcess(id)
				if err != nil {
					e.log.WithError(err).Warnf("[%s][%s] failed to get old process (%d)", sn, in, id)
					continue
				}
				pn, err := p.Name()
				if err != nil {
					e.log.WithError(err).Warnf("[%s][%s] failed to get name of old process (%d)", sn, in, id)
					continue
				}
				if pn != name {
					e.log.Debugf("[%s][%s] name of old process (%d) not matched, %s -> %s", sn, in, id, name, pn)
					continue
				}
				err = p.Kill()
				if err != nil {
					e.log.Warnf("[%s][%s] failed to stop the old process (%d)", sn, in, id)
				} else {
					e.log.Infof("[%s][%s] old process (%d) stopped", sn, in, id)
				}
			}
		}
	}
}

// Run new service
func (e *nativeEngine) Run(cfg openedge.ServiceInfo, vs map[string]openedge.VolumeInfo) (engine.Service, error) {
	spwd := path.Join(e.pwd, "var", "run", "openedge", "services", cfg.Name)
	err := os.RemoveAll(spwd)
	if err != nil {
		return nil, err
	}
	err = mountAll(e.pwd, spwd, cfg.Mounts, vs)
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
	params := processConfigs{
		exec: path.Join(pkgDir, pkg.Entry),
		env:  utils.AppendEnv(cfg.Env, true),
		argv: cfg.Args,
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

// Close engine
func (e *nativeEngine) Close() error {
	return nil
}

func mountAll(epwd, spwd string, ms []openedge.MountInfo, vs map[string]openedge.VolumeInfo) error {
	for _, m := range ms {
		v, ok := vs[m.Name]
		if !ok {
			return fmt.Errorf("volume '%s' not found", m.Name)
		}
		err := mount(v.Path, path.Join(spwd, strings.TrimSpace(m.Path)))
		if err != nil {
			return err
		}
	}
	sock := utils.GetEnv(openedge.EnvMasterHostSocket)
	if sock != "" {
		return mount(sock, path.Join(spwd, openedge.DefaultSockFile))
	}
	return nil
}

func mount(src, dst string) error {
	// if it is a file mapping, the file must exist, otherwise it
	// will be used as a dir mapping and make the dir.
	if !utils.PathExists(src) {
		err := os.MkdirAll(src, 0755)
		if err != nil {
			return err
		}
	}
	err := os.MkdirAll(path.Dir(dst), 0755)
	if err != nil {
		return err
	}
	return os.Symlink(src, dst)
}
