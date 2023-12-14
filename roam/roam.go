// Package roam 大文件同步模块，用于集群环境部署下大文件在集群中的传播
package roam

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/config"
)

type Roam interface {
	RoamObject(dir, file, md5, unpack string) error
}

type roamImpl struct {
	cfg  config.Config
	mode string
	ami  ami.AMI
	cli  *http.Client
	wg   *sync.WaitGroup
	mx   *sync.Mutex
	log  *log.Logger
}

func NewRoam(cfg config.Config) (Roam, error) {
	mode := context.RunMode()

	am, err := ami.NewAMI(mode, cfg.AMI, nil)
	if err != nil {
		return nil, errors.Trace(err)
	}

	ops, err := cfg.Roam.ToClientOptions()
	if err != nil {
		return nil, errors.Trace(err)
	}
	r := &roamImpl{
		cfg:  cfg,
		mode: mode,
		ami:  am,
		cli:  http.NewClient(ops),
		wg:   &sync.WaitGroup{},
		mx:   &sync.Mutex{},
		log:  log.L().With(log.Any("baetyl", "roam")),
	}
	r.log.Info("app running mode", log.Any("mode", mode))
	return r, nil
}

func (r *roamImpl) RoamObject(dir, file, md5, unpack string) error {
	r.log.Debug("roam object", log.Any("dir", dir),
		log.Any("file", file), log.Any("md5", md5),
		log.Any("unpack", unpack))
	apps, err := r.ami.StatsApps(context.EdgeSystemNamespace())
	if err != nil {
		return errors.Trace(err)
	}

	errs := []string{}
	for _, app := range apps {
		r.log.Debug("roam object app", log.Any("name", app.Name))
		if app.DeployType != specv1.WorkloadDaemonSet || !strings.Contains(app.Name, specv1.BaetylAgent) {
			continue
		}
		for _, instance := range app.InstanceStats {
			fr := &fileRoamer{
				dir:    dir,
				file:   file,
				md5:    md5,
				unpack: unpack,
				url: url.URL{
					Scheme: "http",
					Host:   fmt.Sprintf("%s:%s", instance.IP, r.cfg.Roam.Port),
					Path:   r.cfg.Roam.Path,
				},
				cli: r.cli,
				wg:  r.wg,
				mx:  r.mx,
				log: r.log,
			}

			r.log.Debug("roam object fr", log.Any("fr", fr))
			r.mx.Lock()
			fr.errs = &errs
			r.mx.Unlock()

			r.wg.Add(1)
			go fr.Uploading()
		}
		break
	}
	r.log.Debug("roam object before wait")
	r.wg.Wait()
	if len(errs) > 0 {
		return errors.Trace(errors.New(strings.Join(errs, "\n")))
	}
	r.log.Debug("roam object success")
	return nil
}
