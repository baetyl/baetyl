package ami

import (
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/models"
	"github.com/baetyl/baetyl-core/utils"
	"github.com/jinzhu/copier"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kl "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/reference"
	"k8s.io/kubectl/pkg/scheme"
	"time"
)

func (k *kubeModel) CollectInfo() (map[string]interface{}, error) {
	var apps models.AppsVersionResource
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
		return nil, err
	}
	nodeStats, err := k.collectNodeStats(node)
	if err != nil {
		return nil, err
	}
	appStatus, err := k.collectAppStatus(apps.Value)
	if err != nil {
		return nil, err
	}

	info := map[string]interface{}{
		"time":      time.Now(),
		"node":      nodeInfo,
		"nodestats": nodeStats,
		"appstats":  appStatus,
	}
	_, err = k.shadow.Report(info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (k *kubeModel) collectNodeInfo(node *v1.Node) (*models.NodeInfo, error) {
	ni := node.Status.NodeInfo
	nodeInfo := &models.NodeInfo{
		Arch:             ni.Architecture,
		KernelVersion:    ni.KernelVersion,
		OS:               ni.OperatingSystem,
		ContainerRuntime: ni.ContainerRuntimeVersion,
		MachineID:        ni.MachineID,
		OSImage:          ni.OSImage,
	}

	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			nodeInfo.Address = addr.Address
		} else if addr.Type == v1.NodeHostName {
			nodeInfo.Hostname = addr.Address
		}
	}
	return nodeInfo, nil
}

func (k *kubeModel) collectNodeStats(node *v1.Node) (*models.NodeStats, error) {
	nodeStats := &models.NodeStats{}
	for res, quantity := range node.Status.Capacity {
		nodeStats.Capacity[string(res)] = &models.ResourceInfo{
			Name:  string(res),
			Value: quantity.String(),
		}
	}
	nodeMetric, err := k.cli.Metrics.NodeMetricses().Get(k.nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	for res, quantity := range nodeMetric.Usage {
		nodeStats.Usage[string(res)] = &models.ResourceInfo{
			Name:  string(res),
			Value: quantity.String(),
		}
	}
	return nodeStats, nil
}

func (k *kubeModel) collectAppStatus(apps map[string]string) (map[string]*models.AppStats, error) {
	res := map[string]*models.AppStats{}
	for name, ver := range apps {
		status := &models.AppStats{
			Name:    name,
			Version: ver,
		}
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
		var app models.Application
		err = k.store.Get(utils.MakeKey(common.Application, app.Name, app.Version), &app)
		if err != nil {
			return nil, err
		}
		status.ServiceInfos = map[string]*models.ServiceInfo{}
		for _, svc := range app.Services {
			status.ServiceInfos[svc.Name], err = k.collectServiceInfo(svc.Name)
			if err != nil {
				return nil, err
			}
		}
		status.VolumeInfos = map[string]*models.VolumeInfo{}
		for _, v := range app.Volumes {
			status.VolumeInfos[v.Name], err = k.collectVolumeInfo(v)
			if err != nil {
				return nil, err
			}
		}
		res[name] = status
	}
	return res, nil
}

func (k *kubeModel) collectServiceInfo(name string) (*models.ServiceInfo, error) {
	info := &models.ServiceInfo{Name: name}
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
	info.Container = models.Container{
		// service only have one container
		Name: pod.Status.ContainerStatuses[0].Name,
		ID:   pod.Status.ContainerStatuses[0].ContainerID,
	}
	nodeMetric, err := k.cli.Metrics.NodeMetricses().Get(pod.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	for res, quantity := range nodeMetric.Usage {
		info.Usage[string(res)] = &models.ResourceInfo{
			Name:  string(res),
			Value: quantity.String(),
		}
	}
	info.Status = string(pod.Status.Phase)
	return info, nil
}

func (k *kubeModel) collectVolumeInfo(volume models.Volume) (*models.VolumeInfo, error) {
	info := &models.VolumeInfo{Name: volume.Name}
	if volume.VolumeSource.Configuration != nil {
		configMap, err := k.cli.Core.ConfigMaps(k.cli.Namespace).Get(volume.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		info.Version = configMap.ResourceVersion
	} else if volume.VolumeSource.Secret != nil {
		secret, err := k.cli.Core.Secrets(k.cli.Namespace).Get(volume.Name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		info.Version = secret.ResourceVersion
	}
	return info, nil
}
