package ami

import (
	"github.com/baetyl/baetyl-go/errors"
	"github.com/baetyl/baetyl-go/log"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/reference"
	"k8s.io/kubectl/pkg/scheme"
)

func (k *kubeImpl) CollectNodeInfo() (*specv1.NodeInfo, error) {
	node, err := k.cli.core.Nodes().Get(k.knn, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}
	ni := node.Status.NodeInfo
	nodeInfo := &specv1.NodeInfo{
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
	return nodeInfo, nil
}

func (k *kubeImpl) CollectNodeStats() (*specv1.NodeStatus, error) {
	node, err := k.cli.core.Nodes().Get(k.knn, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}
	nodeStats := &specv1.NodeStatus{
		Usage:    map[string]string{},
		Capacity: map[string]string{},
	}
	nodeMetric, err := k.cli.metrics.NodeMetricses().Get(k.knn, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Trace(err)
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

func (k *kubeImpl) CollectAppStatus(ns string) ([]specv1.AppStatus, error) {
	deploys, err := k.cli.app.Deployments(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}
	appStatuses := map[string]*specv1.AppStatus{}
	for _, deploy := range deploys.Items {
		appName := deploy.Labels[AppName]
		appVersion := deploy.Labels[AppVersion]
		serviceName := deploy.Labels[ServiceName]
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
		dRef, err := reference.GetReference(scheme.Scheme, &deploy)
		if err != nil {
			return nil, errors.Trace(err)
		}
		events, _ := k.cli.core.Events(ns).Search(scheme.Scheme, dRef)
		if l := len(events.Items); l > 0 {
			if e := events.Items[l-1]; e.Type == "Warning" {
				status.Cause += e.Message
			}
		}
		selector := labels.SelectorFromSet(deploy.Spec.Selector.MatchLabels)
		pods, err := k.cli.core.Pods(ns).List(metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return nil, errors.Trace(err)
		}
		if pods == nil || len(pods.Items) == 0 {
			continue
		}
		pod := pods.Items[0]
		if status.ServiceInfos == nil {
			status.ServiceInfos = map[string]*specv1.ServiceInfo{}
		}
		status.ServiceInfos[serviceName] = k.collectServiceInfo(ns, serviceName, &pod)
		status.Status = getDeployStatus(status.ServiceInfos)
		appStatuses[appName] = status
	}
	var res []specv1.AppStatus
	for _, status := range appStatuses {
		res = append(res, *status)
	}
	return res, nil
}

func getDeployStatus(infos map[string]*specv1.ServiceInfo) string {
	var pending = false
	for _, info := range infos {
		if info.Status == string(corev1.PodPending) {
			pending = true
		} else if info.Status == string(corev1.PodFailed) {
			return info.Status
		}
	}
	if pending {
		return string(corev1.PodPending)
	}
	return string(corev1.PodRunning)
}

func (k *kubeImpl) collectServiceInfo(ns, serviceName string, pod *corev1.Pod) *specv1.ServiceInfo {
	info := &specv1.ServiceInfo{Name: serviceName, Usage: map[string]string{}}
	ref, err := reference.GetReference(scheme.Scheme, pod)
	if err != nil {
		k.log.Warn("failed to get service reference", log.Error(err))
		return info
	}
	events, _ := k.cli.core.Events(ns).Search(scheme.Scheme, ref)
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
	info.Status = string(pod.Status.Phase)
	for _, st := range pod.Status.ContainerStatuses {
		if st.State.Waiting != nil {
			info.Status = string(corev1.PodPending)
		}
	}
	podMetric, err := k.cli.metrics.PodMetricses(ns).Get(pod.Name, metav1.GetOptions{})
	if err != nil {
		k.log.Warn("failed to collect pod metrics", log.Error(err))
		return info
	}
	for _, cont := range podMetric.Containers {
		if cont.Name == serviceName {
			for res, quan := range cont.Usage {
				info.Usage[string(res)] = quan.String()
			}
		}
	}
	return info
}
