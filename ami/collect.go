package ami

import (
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/utils"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/spec/v1"
	"github.com/jinzhu/copier"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kl "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/reference"
	"k8s.io/kubectl/pkg/scheme"
	"time"
)

type appsVersionResource struct {
	Name  string                 `yaml:"name" json:"name"`
	Value map[string]interface{} `yaml:"value" json:"value"`
}

func (k *kubeModel) CollectInfo() (map[string]interface{}, error) {
	var apps appsVersionResource
	err := k.store.Get(common.DefaultAppsKey, &apps)
	if err != nil {
		return nil, err
	}
	node, err := k.cli.Core.Nodes().Get(k.nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	nodeInfo, err := k.collectNodeInfo(node)
	if err != nil {
		k.log.Error("failed to collect node info", log.Error(err))
	}
	nodeStats, err := k.collectNodeStats(node)
	if err != nil {
		k.log.Error("failed to collect node status", log.Error(err))
	}
	appStatus, err := k.collectAppStatus(apps.Value)
	if err != nil {
		k.log.Error("failed to collect app status", log.Error(err))
	}

	info := map[string]interface{}{
		"time":      time.Now(),
		"node":      nodeInfo,
		"nodestats": nodeStats,
		"apps":      apps.Value,
		"appstats":  appStatus,
	}
	return info, nil
}

func (k *kubeModel) collectNodeInfo(node *corev1.Node) (*v1.NodeInfo, error) {
	ni := node.Status.NodeInfo
	nodeInfo := &v1.NodeInfo{
		Arch:             ni.Architecture,
		KernelVersion:    ni.KernelVersion,
		OS:               ni.OperatingSystem,
		ContainerRuntime: ni.ContainerRuntimeVersion,
		MachineID:        ni.MachineID,
		OSImage:          ni.OSImage,
	}

	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			nodeInfo.Address = addr.Address
		} else if addr.Type == corev1.NodeHostName {
			nodeInfo.Hostname = addr.Address
		}
	}
	return nodeInfo, nil
}

func (k *kubeModel) collectNodeStats(node *corev1.Node) (*v1.NodeStatus, error) {
	nodeStats := &v1.NodeStatus{
		Usage:    map[string]*v1.ResourceInfo{},
		Capacity: map[string]*v1.ResourceInfo{},
	}
	for res, quantity := range node.Status.Capacity {
		nodeStats.Capacity[string(res)] = &v1.ResourceInfo{
			Name:  string(res),
			Value: quantity.String(),
		}
	}
	nodeMetric, err := k.cli.Metrics.NodeMetricses().Get(k.nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	for res, quantity := range nodeMetric.Usage {
		nodeStats.Usage[string(res)] = &v1.ResourceInfo{
			Name:  string(res),
			Value: quantity.String(),
		}
	}
	return nodeStats, nil
}

func (k *kubeModel) collectAppStatus(apps map[string]interface{}) ([]*v1.AppStatus, error) {
	var res []*v1.AppStatus
	for name, ver := range apps {
		status := &v1.AppStatus{}
		status.Name = name
		status.Version = ver.(string)
		deploy, err := k.cli.App.Deployments(k.cli.Namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		ref, err := reference.GetReference(scheme.Scheme, deploy)
		events, _ := k.cli.Core.Events(k.cli.Namespace).Search(scheme.Scheme, ref)
		for _, e := range events.Items {
			if e.Type == "Warning" {
				status.Cause += e.Message + "\n"
			}
		}
		var app v1.Application
		err = k.store.Get(utils.MakeKey(common.Application, name, ver.(string)), &app)
		if err != nil {
			return nil, err
		}
		status.ServiceInfos = map[string]*v1.ServiceInfo{}
		for _, svc := range app.Services {
			status.ServiceInfos[svc.Name], err = k.collectServiceInfo(svc.Name)
			if err != nil {
				return nil, err
			}
		}
		status.VolumeInfos = map[string]*v1.VolumeInfo{}
		for _, v := range app.Volumes {
			status.VolumeInfos[v.Name], err = k.collectVolumeInfo(v)
			if err != nil {
				return nil, err
			}
		}
		res = append(res, status)
	}
	return res, nil
}

func (k *kubeModel) collectServiceInfo(name string) (*v1.ServiceInfo, error) {
	info := &v1.ServiceInfo{Name: name, Usage: map[string]*v1.ResourceInfo{}}
	svc, err := k.cli.Core.Services(k.cli.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	ls := kl.Set{}
	err = copier.Copy(&ls, &svc.Spec.Selector)
	if err != nil {
		return nil, err
	}
	pods, err := k.cli.Core.Pods(k.cli.Namespace).List(metav1.ListOptions{
		LabelSelector: ls.String(),
		// limit 1 temporarily
		Limit: 1,
	})
	if err != nil {
		return nil, err
	}
	pod := pods.Items[0]
	ref, err := reference.GetReference(scheme.Scheme, &pod)
	events, _ := k.cli.Core.Events(k.cli.Namespace).Search(scheme.Scheme, ref)
	for _, e := range events.Items {
		if e.Type == "Warning" {
			info.Cause += e.Message + "\n"
		}
	}
	info.CreateTime = pod.CreationTimestamp.Local()
	for _, cont := range pod.Status.ContainerStatuses {
		if cont.Name == name {
			info.Container.Name = name
			info.Container.ID = cont.ContainerID
		}
	}
	podMetric, err := k.cli.Metrics.PodMetricses(k.cli.Namespace).Get(pod.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	for _, cont := range podMetric.Containers {
		if cont.Name == name {
			for res, quantity := range cont.Usage {
				info.Usage[string(res)] = &v1.ResourceInfo{
					Name:  string(res),
					Value: quantity.String(),
				}
			}
		}
	}
	info.Status = string(pod.Status.Phase)
	return info, nil
}

func (k *kubeModel) collectVolumeInfo(volume v1.Volume) (*v1.VolumeInfo, error) {
	info := &v1.VolumeInfo{Name: volume.Name}
	if config := volume.VolumeSource.Config; config != nil {
		configMap, err := k.cli.Core.ConfigMaps(k.cli.Namespace).Get(config.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		info.Version = configMap.ResourceVersion
	} else if secret := volume.VolumeSource.Secret; secret != nil {
		secret, err := k.cli.Core.Secrets(k.cli.Namespace).Get(secret.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		info.Version = secret.ResourceVersion
	}
	return info, nil
}
