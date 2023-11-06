package kube

import (
	"log"
	"os"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	logv2 "github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	bh "github.com/timshannon/bolthold"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/config"
)

type kubeImpl struct {
	knn   string // kube node name
	cli   *client
	helm  *action.Configuration
	store *bh.Store
	conf  *config.KubeConfig
	log   *logv2.Logger
}

func init() {
	ami.Register("kube", newKubeImpl)
	ami.Register("kubernetes", newKubeImpl)
}

func newKubeImpl(cfg config.AmiConfig) (ami.AMI, error) {
	kubeCli, err := newClient(cfg.Kube)
	if err != nil {
		return nil, err
	}
	knn := os.Getenv(KubeNodeName)

	// init helm for just list
	actionConfig := new(action.Configuration)
	if err = actionConfig.Init(&genericclioptions.ConfigFlags{}, "", os.Getenv(HelmDriver), log.Printf); err != nil {
		return nil, err
	}

	model := &kubeImpl{
		knn:  knn,
		cli:  kubeCli,
		helm: actionConfig,
		conf: &cfg.Kube,
		log:  logv2.With(logv2.Any("ami", "kube")),
	}
	return model, nil
}

func (k *kubeImpl) ApplyApp(ns string, app specv1.Application, cfgs map[string]specv1.Configuration, secs map[string]specv1.Secret) error {
	if app.Type == specv1.AppTypeHelm {
		return k.ApplyHelm(ns, app, cfgs)
	}
	err := k.checkAndCreateNamespace(ns)
	if err != nil {
		return errors.Trace(err)
	}
	if err = k.applyConfigurations(ns, cfgs); err != nil {
		return errors.Trace(err)
	}
	if err = k.applySecrets(ns, secs); err != nil {
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
	if err = k.applyApplication(ns, app, imagePullSecs); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (k *kubeImpl) DeleteApp(ns string, app string) error {
	if ns == context.EdgeNamespace() {
		err := k.DeleteHelm(ns, app)
		// If delete helm success or err is not ErrNotHelmApp, return directly
		if err == nil || err.Error() != ErrNotHelmApp {
			return err
		}
	}
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
				k.log.Warn("failed to collect qps stats", logv2.Error(errors.Trace(err)))
			}
			k.log.Debug("collect qps stats successfully", logv2.Any("qpsStats", qpsExts))
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
	// Collect only in baetyl-edge namespace
	if ns == context.EdgeNamespace() {
		helmStats, err := k.StatsHelm(ns)
		if err != nil {
			return nil, errors.Trace(err)
		}
		res = append(res, helmStats...)
	}
	return res, nil
}
