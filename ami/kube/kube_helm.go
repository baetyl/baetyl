// Package kube implements a kubernetes client
// kube_helm.go is the implementation of helm for connecting to the kubernetes cluster
package kube

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/baetyl/baetyl-go/v2/errors"
	logv2 "github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	BaetylHelmVersion    = "baetyl-version"
	DefaultHelmNamespace = "default"
	HelmDriver           = "HELM_DRIVER"

	ErrNotHelmApp = "not a helm app"
)

// collectHelmStats collects the stats of helm by namespace and app names
func (k *kubeImpl) collectHelmStats(appStats map[string]specv1.AppStats, ns string, apps []string) error {
	pods, err := k.cli.core.Pods(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return errors.Trace(err)
	}
	if pods == nil || len(pods.Items) == 0 {
		return nil
	}
	for _, app := range apps {
		stats := appStats[app]
		if stats.InstanceStats == nil {
			stats.InstanceStats = map[string]specv1.InstanceStats{}
		}
		insStats := map[string]specv1.InstanceStats{}
		for _, pod := range pods.Items {
			// check if the pod is a helm app by pod name
			if strings.HasPrefix(pod.Name, app) {
				stats.InstanceStats[pod.Name] = k.collectInstanceStats(ns, app, map[string]interface{}{}, &pod)
				insStats[pod.Name] = stats.InstanceStats[pod.Name]
			}
		}
		appStats[app] = stats
	}
	return nil
}

// StatsHelm collects the stats of helm pods
func (k *kubeImpl) StatsHelm(_ string) ([]specv1.AppStats, error) {
	helms, err := k.ListHelm()
	if err != nil {
		return nil, errors.Trace(err)
	}
	if helms == nil || len(helms) == 0 {
		return nil, nil
	}
	appStats := map[string]specv1.AppStats{}
	nsMap := make(map[string][]string)
	for _, h := range helms {
		if version, ok := h.Labels[BaetylHelmVersion]; !ok || version == "" {
			continue
		}
		appStats[h.Name] = specv1.AppStats{
			AppInfo: specv1.AppInfo{
				Name:    h.Name,
				Version: h.Labels[BaetylHelmVersion],
			},
			Status:     transStatus(h.Info.Status),
			DeployType: specv1.WorkloadCustom,
		}
		if _, ok := nsMap[h.Namespace]; !ok {
			nsMap[h.Namespace] = []string{h.Name}
		} else {
			nsMap[h.Namespace] = append(nsMap[h.Namespace], h.Name)
		}
	}
	for ns, apps := range nsMap {
		err = k.collectHelmStats(appStats, ns, apps)
		if err != nil {
			continue
		}
	}
	var res []specv1.AppStats
	for _, stats := range appStats {
		res = append(res, stats)
	}
	return res, nil
}

// ListHelm lists the helm releases using stored configuration
func (k *kubeImpl) ListHelm() ([]*release.Release, error) {
	cli := action.NewList(k.helm)
	return cli.Run()
}

// GetHelm gets the helm release
func (k *kubeImpl) GetHelm(cfg *action.Configuration, app string) (*release.Release, error) {
	cli := action.NewGet(cfg)
	return cli.Run(app)
}

// DeleteHelmByCfg delete helm by helm configuration
func (k *kubeImpl) DeleteHelmByCfg(cfg *action.Configuration, app string) error {
	cli := action.NewUninstall(cfg)
	result, err := cli.Run(app)
	if result != nil && result.Release != nil {
		k.log.Debug("helm uninstall", logv2.Any("release", result.Release.Name))
	}
	return err
}

// ApplyHelm apply the helm release
func (k *kubeImpl) ApplyHelm(ns string, app specv1.Application, cfgs map[string]specv1.Configuration) error {
	ns, ok := app.Labels[specv1.CustomAppNsLabel]
	if !ok || ns == "" {
		ns = DefaultHelmNamespace
	}
	helmCfg := new(action.Configuration)
	if err := helmCfg.Init(&genericclioptions.ConfigFlags{Namespace: &ns}, ns, os.Getenv(HelmDriver), log.Printf); err != nil {
		return errors.Trace(err)
	}
	old, err := k.GetHelm(helmCfg, app.Name)
	// already exists, check version
	if err == nil {
		if version, ok := old.Labels[BaetylHelmVersion]; !ok || version == app.Version {
			k.log.Warn("helm release already exists", logv2.Any("release", app.Name))
			return nil
		} else if !app.PreserveUpdates {
			err = k.DeleteHelmByCfg(helmCfg, app.Name)
			if err != nil {
				return errors.Trace(err)
			}
		} else {
			return k.UpdateHelm(helmCfg, app, cfgs, old)
		}
	}
	cli := action.NewInstall(helmCfg)
	cli.Namespace = ns
	cli.ReleaseName = app.Name
	cli.CreateNamespace = true
	cli.Labels = map[string]string{BaetylHelmVersion: app.Version}
	if len(app.Services) != 1 || len(app.Volumes) < 1 {
		return errors.Trace(errors.New("helm chart only support one service"))
	}
	if app.Services[0].Runtime != "" {
		if len(app.Volumes) < 2 {
			return errors.Trace(errors.New("no value has been selected"))
		}
	}
	dir, vals, err := setChartValues(app, cfgs)
	if err != nil {
		return errors.Trace(err)
	}
	chart, err := loader.Load(dir)
	if err != nil {
		return errors.Trace(err)
	}
	rel, err := cli.Run(chart, vals)
	if rel != nil {
		k.log.Debug("helm install", logv2.Any("release", rel.Name))
	}
	return err
}

// UpdateHelm updates the helm release
func (k *kubeImpl) UpdateHelm(cfg *action.Configuration, app specv1.Application, cfgs map[string]specv1.Configuration, old *release.Release) error {
	cli := action.NewUpgrade(cfg)
	old.Labels[BaetylHelmVersion] = app.Version
	cli.Labels = old.Labels
	if len(app.Services) != 1 || len(app.Volumes) < 1 {
		return errors.Trace(errors.New("helm chart only support one service"))
	}
	if app.Services[0].Runtime != "" {
		if len(app.Volumes) < 2 {
			return errors.Trace(errors.New("no value has been selected"))
		}
	}
	dir, vals, err := setChartValues(app, cfgs)
	if err != nil {
		return errors.Trace(err)
	}
	chart, err := loader.Load(dir)
	if err != nil {
		return errors.Trace(err)
	}
	rel, err := cli.Run(app.Name, chart, vals)
	if rel != nil {
		k.log.Debug("helm upgrade", logv2.Any("release", rel.Name))
	}
	return err
}

// DeleteHelm check if the helm release exists, if not delete it
func (k *kubeImpl) DeleteHelm(app string) error {
	helms, err := k.ListHelm()
	if err != nil || len(helms) == 0 {
		return errors.New(ErrNotHelmApp)
	}
	ns := ""
	for _, h := range helms {
		if h.Name == app {
			ns = h.Namespace
		}
	}
	if ns == "" {
		return errors.New(ErrNotHelmApp)
	}
	helmCfg := new(action.Configuration)
	if err = helmCfg.Init(&genericclioptions.ConfigFlags{Namespace: &ns}, ns, os.Getenv(HelmDriver), log.Printf); err != nil {
		return errors.Trace(err)
	}
	return k.DeleteHelmByCfg(helmCfg, app)
}

// setChartValues get helm chart path and values from app service
func setChartValues(app specv1.Application, cfgs map[string]specv1.Configuration) (string, map[string]interface{}, error) {
	var dir string
	var valueConfig string
	svc := app.Services[0]

	source1 := app.Volumes[0].VolumeSource

	// No value has been selected, return
	if app.Services[0].Runtime == "" {
		if source1.HostPath != nil {
			return source1.HostPath.Path + "/" + svc.Image, nil, nil
		} else {
			return "", nil, errors.Trace(errors.New("no chart tar has been selected"))
		}
	}
	source2 := app.Volumes[1].VolumeSource
	if source1.HostPath != nil {
		dir = source1.HostPath.Path + "/" + svc.Image
		if source2.Config == nil {
			return "", nil, errors.Trace(errors.New("no value has been selected"))
		}
		valueConfig = source2.Config.Name
	} else if source2.HostPath != nil {
		dir = source2.HostPath.Path + "/" + svc.Image
		if source1.Config == nil {
			return "", nil, errors.Trace(errors.New("no value has been selected"))
		}
		valueConfig = source1.Config.Name
	} else {
		return "", nil, errors.Trace(errors.New("no value has been selected"))
	}
	cfg, ok := cfgs[valueConfig]
	if !ok {
		return "", nil, errors.Trace(errors.New("config not exist"))
	}
	data, ok := cfg.Data[svc.Runtime]
	if !ok {
		return "", nil, errors.Trace(errors.New("config not exist"))
	}
	var result map[string]interface{}
	err := yaml.Unmarshal([]byte(data), &result)
	if err != nil {
		return "", nil, errors.Trace(err)
	}
	return dir, result, nil
}

// transStatus transform helm status to baetyl app status
func transStatus(status release.Status) specv1.Status {
	switch status {
	case release.StatusUnknown:
		return specv1.Unknown
	case release.StatusDeployed:
		return specv1.Running
	case release.StatusFailed:
		return specv1.Failed
	default:
		return specv1.Pending
	}
}
