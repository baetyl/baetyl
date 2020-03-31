package ami

import (
	"fmt"
	"github.com/jinzhu/copier"
	kl "k8s.io/apimachinery/pkg/labels"
	"time"

	"github.com/baetyl/baetyl-go/log"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/reference"
	"k8s.io/kubectl/pkg/scheme"
)

func (k *kubeImpl) Collect() (specv1.Report, error) {
	node, err := k.cli.Core.Nodes().Get(k.knn, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	nodeInfo := k.collectNodeInfo(node)
	nodeStats, err := k.collectNodeStats(node)
	if err != nil {
		k.log.Error("failed to collect node status", log.Error(err))
	}
	appStatus, err := k.collectAppStatus()
	if err != nil {
		k.log.Error("failed to collect app status", log.Error(err))
	}
	var apps []specv1.AppInfo
	for _, info := range appStatus {
		app := specv1.AppInfo{
			Name:    info.Name,
			Version: info.Version,
		}
		apps = append(apps, app)
	}
	return specv1.Report{
		"time":      time.Now(),
		"node":      nodeInfo,
		"nodestats": nodeStats,
		"apps":      apps,
		"appstats":  appStatus,
	}, nil
}

func (k *kubeImpl) collectNodeInfo(node *corev1.Node) specv1.NodeInfo {
	ni := node.Status.NodeInfo
	nodeInfo := specv1.NodeInfo{
		Arch:             ni.Architecture,
		KernelVersion:    ni.KernelVersion,
		OS:               ni.OperatingSystem,
		ContainerRuntime: ni.ContainerRuntimeVersion,
		MachineID:        ni.MachineID,
		OSImage:          ni.OSImage,
		BootID:           ni.BootID,
		SystemUUID:       ni.SystemUUID,
	}
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			nodeInfo.Address = addr.Address
		} else if addr.Type == corev1.NodeHostName {
			nodeInfo.Hostname = addr.Address
		}
	}
	return nodeInfo
}

func (k *kubeImpl) collectNodeStats(node *corev1.Node) (specv1.NodeStatus, error) {
	nodeStats := specv1.NodeStatus{
		Usage:    map[string]string{},
		Capacity: map[string]string{},
	}
	nodeMetric, err := k.cli.Metrics.NodeMetricses().Get(k.knn, metav1.GetOptions{})
	if err != nil {
		return nodeStats, err
	}
	for res, quan := range nodeMetric.Usage {
		nodeStats.Usage[string(res)] = quan.String()
	}
	for res, quan := range node.Status.Capacity {
		if _, ok := nodeStats.Usage[string(res)]; ok {
			nodeStats.Capacity[string(res)] = quan.String()
		}
	}
	return nodeStats, nil
}

func (k *kubeImpl) collectAppStatus() ([]specv1.AppStatus, error) {
	deploys, err := k.cli.App.Deployments(k.cli.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	appStatuses := map[string]*specv1.AppStatus{}
	if deploys == nil {
		return nil, nil
	}
	for _, deploy := range deploys.Items {
		ls := kl.Set{}
		selector := deploy.Spec.Selector.MatchLabels
		err := copier.Copy(&ls, &selector)
		pods, err := k.cli.Core.Pods(k.cli.Namespace).List(metav1.ListOptions{
			LabelSelector: ls.String(),
		})
		if pods == nil || len(pods.Items) > 1 {
			return nil, fmt.Errorf("no pod or more than one pod exists")
		}
		pod := pods.Items[0]
		appName := pod.Labels[AppName]
		appVersion := pod.Labels[AppVersion]
		serviceName := pod.Labels[ServiceName]
		if appName == "" || serviceName == "" {
			continue
		}
		var status *specv1.AppStatus
		if status = appStatuses[appName]; status == nil {
			status = &specv1.AppStatus{AppInfo: specv1.AppInfo{
				Name:    appName,
				Version: appVersion,
			}}
		}
		if err != nil {
			return nil, err
		}
		ref, err := reference.GetReference(scheme.Scheme, &deploy)
		events, _ := k.cli.Core.Events(k.cli.Namespace).Search(scheme.Scheme, ref)
		for _, e := range events.Items {
			if e.Type == "Warning" {
				status.Cause += e.Message + "\n"
			}
		}
		if status.ServiceInfos == nil {
			status.ServiceInfos = map[string]*specv1.ServiceInfo{}
		}
		status.ServiceInfos[serviceName], err = k.collectServiceInfo(serviceName, &pod)
		if err != nil {
			return nil, err
		}
		appStatuses[appName] = status
	}
	return transformAppStatus(appStatuses), nil
}

func transformAppStatus(appStatus map[string]*specv1.AppStatus) []specv1.AppStatus {
	var res []specv1.AppStatus
	for _, status := range appStatus {
		res = append(res, *status)
	}
	return res
}

func (k *kubeImpl) collectServiceInfo(serviceName string, pod *corev1.Pod) (*specv1.ServiceInfo, error) {
	info := &specv1.ServiceInfo{Name: serviceName, Usage: map[string]string{}}
	ref, err := reference.GetReference(scheme.Scheme, pod)
	events, _ := k.cli.Core.Events(k.cli.Namespace).Search(scheme.Scheme, ref)
	for _, e := range events.Items {
		if e.Type == "Warning" {
			info.Cause += e.Message + "\n"
		}
	}
	info.CreateTime = pod.CreationTimestamp.Local()
	for _, cont := range pod.Status.ContainerStatuses {
		if cont.Name == serviceName {
			info.Container.Name = serviceName
			info.Container.ID = cont.ContainerID
		}
	}
	podMetric, err := k.cli.Metrics.PodMetricses(k.cli.Namespace).Get(pod.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	for _, cont := range podMetric.Containers {
		if cont.Name == serviceName {
			for res, quan := range cont.Usage {
				info.Usage[string(res)] = quan.String()
			}
		}
	}
	info.Status = string(pod.Status.Phase)
	return info, nil
}
