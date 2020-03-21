package ami

import (
	"time"

	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/spec/crd"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/jinzhu/copier"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kl "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/reference"
	"k8s.io/kubectl/pkg/scheme"
)

func (k *kubeModel) CollectInfo() (specv1.Report, error) {
	var apps specv1.Desire
	err := k.store.Get("apps", &apps)
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
	appStatus, err := k.collectAppStatus(apps.AppInfos())
	if err != nil {
		k.log.Error("failed to collect app status", log.Error(err))
	}

	return specv1.Report{
		"time":     time.Now(),
		"node":     nodeInfo,
		"nodestat": nodeStats,
		"apps":     apps.AppInfos,
		"appstats": appStatus,
	}, nil
}

func (k *kubeModel) collectNodeInfo(node *corev1.Node) (specv1.NodeInfo, error) {
	ni := node.Status.NodeInfo
	nodeInfo := specv1.NodeInfo{
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

func (k *kubeModel) collectNodeStats(node *corev1.Node) (specv1.NodeStatus, error) {
	nodeStats := specv1.NodeStatus{
		Usage:    map[string]*specv1.ResourceInfo{},
		Capacity: map[string]*specv1.ResourceInfo{},
	}
	for res, quantity := range node.Status.Capacity {
		nodeStats.Capacity[string(res)] = &specv1.ResourceInfo{
			Name:  string(res),
			Value: quantity.String(),
		}
	}
	nodeMetric, err := k.cli.Metrics.NodeMetricses().Get(k.nodeName, metav1.GetOptions{})
	if err != nil {
		return nodeStats, err
	}
	for res, quantity := range nodeMetric.Usage {
		nodeStats.Usage[string(res)] = &specv1.ResourceInfo{
			Name:  string(res),
			Value: quantity.String(),
		}
	}
	return nodeStats, nil
}

func (k *kubeModel) collectAppStatus(apps []specv1.AppInfo) ([]specv1.AppStatus, error) {
	var res []specv1.AppStatus
	for _, app := range apps {
		status := specv1.AppStatus{AppInfo: app}
		deploy, err := k.cli.App.Deployments(k.cli.Namespace).Get(app.Name, metav1.GetOptions{})
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
		var appinfo crd.Application
		err = k.store.Get(makeKey(crd.KindApplication, app.Name, app.Version), &appinfo)
		if err != nil {
			return nil, err
		}
		status.ServiceInfos = map[string]*specv1.ServiceInfo{}
		for _, svc := range appinfo.Services {
			status.ServiceInfos[svc.Name], err = k.collectServiceInfo(svc.Name)
			if err != nil {
				return nil, err
			}
		}
		status.VolumeInfos = map[string]*specv1.VolumeInfo{}
		for _, v := range appinfo.Volumes {
			status.VolumeInfos[v.Name], err = k.collectVolumeInfo(v)
			if err != nil {
				return nil, err
			}
		}
		res = append(res, status)
	}
	return res, nil
}

func (k *kubeModel) collectServiceInfo(name string) (*specv1.ServiceInfo, error) {
	info := &specv1.ServiceInfo{Name: name, Usage: map[string]*specv1.ResourceInfo{}}
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
				info.Usage[string(res)] = &specv1.ResourceInfo{
					Name:  string(res),
					Value: quantity.String(),
				}
			}
		}
	}
	info.Status = string(pod.Status.Phase)
	return info, nil
}

func (k *kubeModel) collectVolumeInfo(volume crd.Volume) (*specv1.VolumeInfo, error) {
	info := &specv1.VolumeInfo{Name: volume.Name}
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
