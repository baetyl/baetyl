package native

import (
	"io/ioutil"
	"os"
	"path"

	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/master/engine"
)

// NAME of this engine
const NAME = "native"

func init() {
	engine.Factories()[NAME] = New
}

type nativeEngine struct {
	wdir string
	log  openedge.Logger
}

// Close engine
func (e *nativeEngine) Close() error {
	return nil
}

// Name of engine
func (e *nativeEngine) Name() string {
	return NAME
}

// New native engine
func New(wdir string) (engine.Engine, error) {
	return &nativeEngine{
		wdir: wdir,
		log:  openedge.WithField("engine", NAME),
	}, nil
}

// Run new service
func (e *nativeEngine) Run(si *openedge.ServiceInfo) (engine.Service, error) {
	wdir := path.Join(e.wdir, "var", "run", "openedge", "service", si.Name)
	err := os.MkdirAll(wdir, 0755)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(path.Join(wdir, "etc"), 0755)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	err = os.Symlink(
		path.Join(e.wdir, "var", "db", "openedge", "service", si.Name),
		path.Join(wdir, "etc", "openedge"),
	)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	err = e.mklog(si.Name, wdir)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	err = e.mount(wdir, si.Mounts)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	p, err := startProcess(e.wdir, wdir, si)
	if err != nil {
		e.log.Errorln("failed to start process:", err.Error())
		os.RemoveAll(wdir)
		return nil, err
	}
	s := &nativeService{
		wdir:    wdir,
		stop:    make(chan struct{}),
		done:    make(chan *os.ProcessState),
		backoff: si.Restart.Backoff.Min,
		e:       e,
		si:      si,
		w:       wdir,
		p:       p,
	}
	go s.supervise()
	return s, nil
}

// RunWithConfig new service
func (e *nativeEngine) RunWithConfig(si *openedge.ServiceInfo, cfg []byte) (engine.Service, error) {
	wdir := path.Join(e.wdir, "var", "run", "openedge", "service", si.Name)
	err := os.MkdirAll(wdir, 0755)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(path.Join(wdir, "etc", "openedge"), 0755)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	err = ioutil.WriteFile(path.Join(wdir, "etc/openedge/service.yml"), cfg, 0644)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	err = e.mklog(si.Name, wdir)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	err = e.mount(wdir, si.Mounts)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	p, err := startProcess(e.wdir, wdir, si)
	if err != nil {
		e.log.Errorln("failed to start process:", err.Error())
		os.RemoveAll(wdir)
		return nil, err
	}
	s := &nativeService{
		wdir:    wdir,
		stop:    make(chan struct{}),
		done:    make(chan *os.ProcessState),
		backoff: si.Restart.Backoff.Min,
		e:       e,
		si:      si,
		w:       wdir,
		p:       p,
	}
	go s.supervise()
	return s, nil
}

func (e *nativeEngine) mklog(name, wdir string) error {
	logdir := path.Join(e.wdir, "var", "log", "openedge", name)
	err := os.MkdirAll(logdir, 0755)
	if err != nil {
		return err
	}
	vardir := path.Join(wdir, "var", "log")
	err = os.MkdirAll(vardir, 0755)
	if err != nil {
		return err
	}
	return os.Symlink(logdir, path.Join(vardir, "openedge"))
}

func (e *nativeEngine) mount(wdir string, ms []openedge.MountInfo) error {
	for _, m := range ms {
		src := path.Join(e.wdir, m.Volume)
		err := os.MkdirAll(src, 0755)
		if err != nil {
			return err
		}
		dst := path.Join(wdir, m.Target)
		err = os.MkdirAll(path.Dir(dst), 0755)
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
