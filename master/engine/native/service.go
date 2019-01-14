package native

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"syscall"
	"time"

	openedge "github.com/baidu/openedge/api/go"

	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/utils"
)

type nativeService struct {
	name    string
	wdir    string
	stop    chan struct{}
	done    chan *os.ProcessState
	retry   int
	backoff time.Duration
	e       *nativeEngine
	si      *openedge.ServiceInfo
	w       string
	p       *os.Process
}

func (s *nativeService) Name() string {
	return s.name
}

func (s *nativeService) Info() *openedge.ServiceInfo {
	return s.si
}

func (s *nativeService) Instances() []engine.Instance {
	return []engine.Instance{}
}

func (s *nativeService) Scale(replica int, grace time.Duration) error {
	return errors.New("not implemented yet")
}

func (s *nativeService) Stop(grace time.Duration) error {
	defer os.RemoveAll(s.wdir)
	s.stop <- struct{}{}
	for {
		select {
		case <-s.done:
			return nil
		case <-time.After(grace):
			s.p.Kill()
		}
	}
}

// Run new service
func (e *nativeEngine) Run(name string, si *openedge.ServiceInfo) (engine.Service, error) {
	wdir := path.Join(e.wdir, "var", "run", "openedge", "service", name)
	err := os.MkdirAll(wdir, 0755)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(path.Join(wdir, "etc", "openedge"), 0755)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	err = os.Symlink(
		path.Join(e.wdir, "var/db/openedge/service", name, "service.yml"),
		path.Join(wdir, "etc/openedge/service.yml"),
	)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	err = os.MkdirAll(path.Join(wdir, "var", "log"), 0755)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	err = os.Symlink(
		path.Join(e.wdir, "var/log/openedge", fmt.Sprintf("%s.log", name)),
		path.Join(wdir, "var/log/openedge-service.log"),
	)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	err = e.mount(wdir, si.Mounts)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	p, err := startProcess(e.wdir, wdir, name, si)
	if err != nil {
		e.log.Errorln("start process fail:", err.Error())
		os.RemoveAll(wdir)
		return nil, err
	}
	// FIX need correct default value
	if si.Restart.Backoff.Factor < 1.0 {
		si.Restart.Backoff.Factor = 1.0
	}
	if si.Restart.Backoff.Min < time.Second {
		si.Restart.Backoff.Min = time.Second
	}
	s := &nativeService{
		name:    name,
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
func (e *nativeEngine) RunWithConfig(name string, si *openedge.ServiceInfo, cfg []byte) (engine.Service, error) {
	wdir := path.Join(e.wdir, "var", "run", "openedge", "service", name)
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
	err = os.MkdirAll(path.Join(wdir, "var", "log"), 0755)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	err = os.Symlink(
		path.Join(e.wdir, "var/log/openedge", fmt.Sprintf("%s.log", name)),
		path.Join(wdir, "var/log/openedge-service.log"),
	)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	err = e.mount(wdir, si.Mounts)
	if err != nil {
		os.RemoveAll(wdir)
		return nil, err
	}
	p, err := startProcess(e.wdir, wdir, name, si)
	if err != nil {
		e.log.Errorln("start process fail:", err.Error())
		os.RemoveAll(wdir)
		return nil, err
	}
	// FIX need correct default value
	if si.Restart.Backoff.Factor < 1.0 {
		si.Restart.Backoff.Factor = 1.0
	}
	if si.Restart.Backoff.Min < time.Second {
		si.Restart.Backoff.Min = time.Second
	}
	s := &nativeService{
		name:    name,
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

func (e *nativeEngine) mount(wdir string, ms []openedge.MountInfo) error {
	for _, m := range ms {
		src := path.Join(e.wdir, "var", "db", "openedge", "service", m.Volume)
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

func (s *nativeService) supervise() {
	go s.waitProcess()
	for {
		select {
		case <-s.stop:
			s.p.Signal(syscall.SIGTERM)
			return
		case ps := <-s.done:
			if s.si.Restart.Policy == openedge.RestartNo {
				<-s.stop
				s.done <- nil
				return
			}
			if s.si.Restart.Policy == openedge.RestartOnFailure && ps != nil && ps.Success() {
				<-s.stop
				s.done <- nil
				return
			}
			for {
				if s.si.Restart.Retry.Max > 0 && s.retry > s.si.Restart.Retry.Max {
					<-s.stop
					s.done <- nil
					return
				}
				select {
				case <-s.stop:
					s.done <- nil
					return
				case <-time.After(s.backoff):
				}
				p, err := startProcess(s.e.wdir, s.w, s.name, s.si)
				s.backoff = time.Duration(int64(float64(s.backoff) * s.si.Restart.Backoff.Factor))
				if int64(s.si.Restart.Backoff.Max) > 0 && s.backoff > s.si.Restart.Backoff.Max {
					s.backoff = s.si.Restart.Backoff.Max
				}
				s.retry++
				if err != nil {
					s.e.log.Errorln("restart process fail:", err.Error())
					continue
				}
				s.p = p
				go s.waitProcess()
				break
			}
		}
	}
}

func startProcess(cdir string, wdir string, name string, si *openedge.ServiceInfo) (*os.Process, error) {
	openedge.Debugln("start process", cdir, wdir)
	pkgdir := path.Join(cdir, "lib", "openedge", "packages", si.Image)
	var pkg Package
	err := utils.LoadYAML(path.Join(pkgdir, PackageConfigPath), &pkg)
	if err != nil {
		return nil, err
	}
	args := make([]string, 0)
	args = append(args, fmt.Sprintf("openedge-service-%s", name))
	for _, p := range si.Params {
		args = append(args, p)
	}
	return os.StartProcess(
		path.Join(pkgdir, pkg.Entry),
		args,
		&os.ProcAttr{
			Dir: wdir,
			Env: utils.AppendEnv(si.Env, true),
			Files: []*os.File{
				os.Stdin,
				os.Stdout,
				os.Stderr,
			},
		},
	)
}

func (s *nativeService) waitProcess() {
	ps, err := s.p.Wait()
	if err != nil {
		s.e.log.Errorln("wait process fail:", err.Error())
	}
	s.done <- ps
}
