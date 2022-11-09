package kube

import (
	"os"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	bh "github.com/timshannon/bolthold"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/config"
)

type kubeImpl struct {
	knn   string // kube node name
	cli   *client
	store *bh.Store
	conf  *config.KubeConfig
	log   *log.Logger
}

func init() {
	ami.Register("kube", newKubeImpl)
	ami.Register("kubernetes", newKubeImpl)
}

func newKubeImpl(cfg config.AmiConfig) (ami.AMI, error) {
	cli, err := newClient(cfg.Kube)
	if err != nil {
		return nil, err
	}
	knn := os.Getenv(KubeNodeName)
	model := &kubeImpl{
		knn:  knn,
		cli:  cli,
		conf: &cfg.Kube,
		log:  log.With(log.Any("ami", "kube")),
	}
	return model, nil
}

func (k *kubeImpl) ApplyApp(ns string, app specv1.Application, cfgs map[string]specv1.Configuration, secs map[string]specv1.Secret) error {
	err := k.checkAndCreateNamespace(ns)
	if err != nil {
		return errors.Trace(err)
	}
	if err := k.applyConfigurations(ns, cfgs); err != nil {
		return errors.Trace(err)
	}
	if err := k.applySecrets(ns, secs); err != nil {
		return errors.Trace(err)
	}
	var imagePullSecs []string
	for n, sec := range secs {
		if isRegistrySecret(sec) {
			imagePullSecs = append(imagePullSecs, n)
		}
	}
	err = k.deleteApplication(ns, app.Name)
	if err != nil {
		return errors.Trace(err)
	}
	if err := k.applyApplication(ns, app, imagePullSecs); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (k *kubeImpl) DeleteApp(ns string, app string) error {
	return k.deleteApplication(ns, app)
}

func (k *kubeImpl) StatsApps(ns string) ([]specv1.AppStats, error) {
	var res []specv1.AppStats
	var qpsExts map[string]interface{}
	var err error
	if extension, ok := ami.Hooks[ami.BaetylQPSStatsExtension]; ok {
		qpsStatsExt, ok := extension.(ami.CollectStatsExtFunc)
		if ok {
			qpsExts, err = qpsStatsExt(context.RunModeKube)
			if err != nil {
				k.log.Warn("failed to collect qps stats", log.Error(errors.Trace(err)))
			}
			k.log.Debug("collect qps stats successfully", log.Any("qpsStats", qpsExts))
		} else {
			k.log.Warn("invalid collecting qps stats function")
		}
	}
	dps, err := k.collectDeploymentStats(ns, qpsExts)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res = append(res, dps...)
	dss, err := k.collectDaemonSetStats(ns, qpsExts)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res = append(res, dss...)
	js, err := k.collectJobStats(ns, qpsExts)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res = append(res, js...)
	return res, nil
}
