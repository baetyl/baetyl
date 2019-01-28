package native

import (
	"errors"
	"os"
	"path"
	"syscall"
	"time"

	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/master/engine"
	"github.com/baidu/openedge/utils"
)

type nativeService struct {
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

func (s *nativeService) Info() *openedge.ServiceInfo {
	return s.si
}

func (s *nativeService) Instances() []engine.Instance {
	return []engine.Instance{}
}

func (s *nativeService) Scale(replica int, grace time.Duration) error {
	return errors.New("not implemented yet")
}

func (s *nativeService) Stats() openedge.ServiceStatus {
	return openedge.ServiceStatus{}
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
				p, err := startProcess(s.e.wdir, s.w, s.si)
				s.backoff = time.Duration(int64(float64(s.backoff) * s.si.Restart.Backoff.Factor))
				if int64(s.si.Restart.Backoff.Max) > 0 && s.backoff > s.si.Restart.Backoff.Max {
					s.backoff = s.si.Restart.Backoff.Max
				}
				s.retry++
				if err != nil {
					s.e.log.Errorln("failed to restart process:", err.Error())
					continue
				}
				s.p = p
				go s.waitProcess()
				break
			}
		}
	}
}

func startProcess(cdir string, wdir string, si *openedge.ServiceInfo) (*os.Process, error) {
	openedge.Debugln("start process", cdir, wdir)
	pkgdir := path.Join(cdir, "lib", "openedge", "packages", si.Image)
	var pkg Package
	err := utils.LoadYAML(path.Join(pkgdir, PackageConfigPath), &pkg)
	if err != nil {
		return nil, err
	}
	args := make([]string, 0)
	args = append(args, si.Name) // add prefix "openedge-service-"?
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
		s.e.log.Errorln("failed to wait process:", err.Error())
	}
	s.done <- ps
}
