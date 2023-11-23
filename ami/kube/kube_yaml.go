package kube

import (
	"context"
	"strings"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/baetyl/baetyl/v2/utils"
)

const (
	CustomYamlAppInfo = "baetyl-yaml-app-info"
)

type CustomInfo struct {
	specv1.AppInfo `yaml:",inline" json:",inline" mapstructure:",squash"`
	Namespace      string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
}

type YamlAppInfo struct {
	AppInfo map[string]CustomInfo
}

func (k *kubeImpl) ApplyYaml(app specv1.Application, cfgs map[string]specv1.Configuration) error {
	if customNs, ok := app.Labels[specv1.CustomAppNsLabel]; customNs != "" && ok {
		k.log.Info("user yaml app custom ns", log.Any("ns", customNs))
		if !app.PreserveUpdates {
			err := k.deleteYamlApp(customNs, app.Name, cfgs)
			if err != nil {
				return errors.Trace(err)
			}
		}
		return k.applyYamlApp(customNs, cfgs)
	}
	return nil
}

func (k *kubeImpl) StatsYaml() ([]specv1.AppStats, error) {
	var res []specv1.AppStats
	customInfo := &YamlAppInfo{}
	err := k.store.Get(CustomYamlAppInfo, customInfo)
	if err != nil {
		k.log.Debug("no yaml app info found", log.Any("err", err))
		return nil, nil
	}
	for name, info := range customInfo.AppInfo {
		custom, err := k.collectCustomStats(name, info)
		if err != nil {
			return nil, errors.Trace(err)
		}
		res = append(res, custom)
	}
	return res, nil
}

func (k *kubeImpl) DeleteYaml(app *specv1.Application) error {
	// get yaml config data
	cfgs := make(map[string]specv1.Configuration)
	for _, v := range app.Volumes {
		if cfg := v.VolumeSource.Config; cfg != nil {
			key := utils.MakeKey(specv1.KindConfiguration, cfg.Name, cfg.Version)
			if key == "" {
				return errors.Errorf("failed to get config name: (%s) version: (%s)", cfg.Name, cfg.Version)
			}
			var config specv1.Configuration
			if err := k.store.Get(key, &config); err != nil {
				return errors.Errorf("failed to get config name: (%s) version: (%s) with error: %s", cfg.Name, cfg.Version, err.Error())
			}
			cfgs[config.Name] = config
		}
	}
	// delete yaml info
	yamlAppInfo := &YamlAppInfo{}
	err := k.store.Get(CustomYamlAppInfo, yamlAppInfo)
	if err == nil {
		delete(yamlAppInfo.AppInfo, app.Name)
		err = k.store.Upsert(CustomYamlAppInfo, yamlAppInfo)
		if err != nil {
			k.log.Error("failed to store yaml app info", log.Any("error", err))
			return err
		}
	}
	// delete yaml app
	if customNs, ok := app.Labels[specv1.CustomAppNsLabel]; customNs != "" && ok {
		return k.deleteYamlApp(customNs, app.Name, cfgs)
	}
	return nil
}

func (k *kubeImpl) collectCustomStats(name string, info CustomInfo) (specv1.AppStats, error) {
	stats := specv1.AppStats{
		AppInfo:    info.AppInfo,
		DeployType: specv1.WorkloadCustom,
	}

	pods, err := k.cli.core.Pods(info.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return stats, err
	}
	if pods == nil || len(pods.Items) == 0 {
		stats.Status = specv1.Running
		return stats, nil
	}
	if stats.InstanceStats == nil {
		stats.InstanceStats = map[string]specv1.InstanceStats{}
	}
	insStats := map[string]specv1.InstanceStats{}
	for _, pod := range pods.Items {
		stats.InstanceStats[pod.Name] = k.collectInstanceStats(info.Namespace, name, map[string]interface{}{}, &pod)
		insStats[pod.Name] = stats.InstanceStats[pod.Name]

	}
	stats.Status = getAppStatus(stats.Status, int32(len(pods.Items)), insStats)

	return stats, nil
}

func (k *kubeImpl) deleteYamlApp(ns string, appName string, cfgs map[string]specv1.Configuration) error {
	for _, cfg := range cfgs {
		for name, data := range cfg.Data {
			if !strings.Contains(name, "yaml") && !strings.Contains(name, "yml") {
				continue
			}
			objs := parseK8SYaml(data)
			if len(objs) == 0 {
				k.log.Info("no k8s object found in cfg data")
				continue
			}
			for _, obj := range objs {
				meta, err := meta.Accessor(obj)
				if err != nil {
					k.log.Info("failed to transfer k8s obj to metav1 obj", log.Any("error", err))
					continue
				}
				rsc := k.getResourceMapping(obj, k.cli.discovery)
				if rsc == nil {
					k.log.Info("failed to get k8s server resource for obj")
					continue
				}
				deletePolicy := metav1.DeletePropagationForeground
				err = k.cli.dynamic.Resource(*rsc).Namespace(ns).Delete(context.Background(), meta.GetName(), metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
				if err != nil {
					k.log.Info("no k8s resource found to delete", log.Any("name", meta.GetName()), log.Any("error", err))
					continue
				}
			}
		}
	}

	return nil
}

func (k *kubeImpl) applyYamlApp(ns string, cfgs map[string]specv1.Configuration) error {
	for _, cfg := range cfgs {
		for name, data := range cfg.Data {
			if !strings.Contains(name, "yaml") && !strings.Contains(name, "yml") {
				continue
			}
			objs := parseK8SYaml(data)
			if len(objs) == 0 {
				continue
			}
			for _, obj := range objs {
				metaAccessor, err := meta.Accessor(obj)
				if err != nil {
					k.log.Error("failed to transfer k8s obj to metav1 obj", log.Any("error", err))
					continue
				}
				metaAccessor.SetNamespace(ns)
				u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
				if err != nil {
					k.log.Error("failed to transfer k8s obj to unstructured obj", log.Any("error", err))
					continue
				}
				us := &unstructured.Unstructured{
					Object: u,
				}
				rsc := k.getResourceMapping(obj, k.cli.discovery)
				if rsc == nil {
					k.log.Info("failed to get k8s server resource for obj")
					continue
				}
				applyOptions := metav1.ApplyOptions{
					FieldManager: "my-controller-name",
				}
				_, err = k.cli.dynamic.Resource(*rsc).Namespace(ns).Apply(context.Background(), metaAccessor.GetName(), us, applyOptions)
				if err != nil {
					k.log.Error("failed to apply k8s resource with dynamic client", log.Any("error", err))
					continue
				}
			}
		}
	}
	k.log.Info("Apply yaml app success")
	return nil
}

func parseK8SYaml(fileR string) []runtime.Object {
	sepYamlfiles := strings.Split(fileR, "---")
	res := make([]runtime.Object, 0, len(sepYamlfiles))
	for _, f := range sepYamlfiles {
		if s := strings.TrimSpace(f); s == "" {
			// ignore empty cases
			continue
		}
		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode([]byte(f), nil, nil)
		if err != nil {
			continue
		}
		res = append(res, obj)
	}
	return res
}

func (k *kubeImpl) getResourceMapping(obj runtime.Object, discovery discovery.DiscoveryInterface) *schema.GroupVersionResource {
	gvk := obj.GetObjectKind().GroupVersionKind()
	apiResources, err := discovery.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		k.log.Error("failed to get server api resources", log.Any("error", err))
		return nil
	}
	for _, apiResource := range apiResources.APIResources {
		if apiResource.Kind == gvk.Kind {
			return &schema.GroupVersionResource{
				Group:    gvk.Group,
				Version:  gvk.Version,
				Resource: apiResource.Name,
			}
		}
	}
	return nil
}
