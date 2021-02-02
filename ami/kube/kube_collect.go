package kube

import (
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/reference"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"

	"github.com/baetyl/baetyl/v2/ami"
)

func (k *kubeImpl) GetMasterNodeName() string {
	return k.knn
}

func (k *kubeImpl) CollectNodeInfo() (map[string]interface{}, error) {
	nodes, err := k.cli.core.Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}
	infos := map[string]interface{}{}
	for _, node := range nodes.Items {
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
			Role:             "worker",
		}
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeHostName {
				nodeInfo.Hostname = addr.Address
			}
		}
		for k, _ := range node.GetLabels() {
			if k == MasterRole {
				nodeInfo.Role = "master"
				break
			}
		}
		infos[node.Name] = nodeInfo
	}
	return infos, nil
}

func (k *kubeImpl) CollectNodeStats() (map[string]interface{}, error) {
	nodes, err := k.cli.core.Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}
	nodeMetrics, err := k.cli.metrics.NodeMetricses().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}

	metrics := map[string]v1beta1.NodeMetrics{}
	for _, metric := range nodeMetrics.Items {
		metrics[metric.Name] = metric
	}
	var exts map[string]interface{}
	if extension, ok := ami.Hooks[ami.BaetylStatsExtension]; ok {
		collectStatsExt, ok := extension.(ami.CollectStatsExtFunc)
		if ok {
			exts, err = collectStatsExt()
			if err != nil {
				k.log.Warn("failed to collect extended stats", log.Error(errors.Trace(err)))
			}
			k.log.Debug("collect extended stats successfully", log.Any("stats", exts))
		} else {
			k.log.Warn("invalid collecting stats function")
		}
	}
	infos := map[string]interface{}{}
	for _, node := range nodes.Items {
		nodeStats := &specv1.NodeStats{
			Usage:    map[string]string{},
			Capacity: map[string]string{},
		}
		nodeMetric, ok := metrics[node.Name]
		if !ok {
			k.log.Warn("failed to collect node metric")
		} else {
			for res, quan := range nodeMetric.Usage {
				nodeStats.Usage[string(res)] = quan.String()
			}
		}
		for res, quan := range node.Status.Capacity {
			if _, ok := nodeStats.Usage[string(res)]; ok {
				nodeStats.Capacity[string(res)] = quan.String()
			}
		}
		for _, cond := range node.Status.Conditions {
			if cond.Status == corev1.ConditionTrue {
				switch cond.Type {
				case corev1.NodeDiskPressure:
					nodeStats.DiskPressure = true
				case corev1.NodeMemoryPressure:
					nodeStats.MemoryPressure = true
				case corev1.NodeReady:
					nodeStats.Ready = true
				case corev1.NodePIDPressure:
					nodeStats.PIDPressure = true
				case corev1.NodeNetworkUnavailable:
					nodeStats.NetworkUnavailable = true
				default:
				}
			}
		}
		if len(exts) > 0 {
			if ext, ok := exts[node.Name]; ok {
				nodeStats.Extension = ext
			}
		}
		infos[node.Name] = nodeStats
	}
	return infos, nil
}

func (k *kubeImpl) collectDeploymentStats(ns string) ([]specv1.AppStats, error) {
	deploys, err := k.cli.app.Deployments(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}
	appStats := map[string]specv1.AppStats{}
	for _, deploy := range deploys.Items {
		appName := deploy.Labels[AppName]
		appVersion := deploy.Labels[AppVersion]
		serviceName := deploy.Labels[ServiceName]
		err = k.collectAppStats(appStats, ns, specv1.AppDeployTypeDeployment,
			appName, appVersion, serviceName, deploy.Spec.Selector.MatchLabels)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	var res []specv1.AppStats
	for _, stats := range appStats {
		res = append(res, stats)
	}
	return res, nil
}

func (k *kubeImpl) collectDaemonSetStats(ns string) ([]specv1.AppStats, error) {
	daemons, err := k.cli.app.DaemonSets(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}
	appStats := map[string]specv1.AppStats{}
	for _, daemon := range daemons.Items {
		appName := daemon.Labels[AppName]
		appVersion := daemon.Labels[AppVersion]
		serviceName := daemon.Labels[ServiceName]
		err = k.collectAppStats(appStats, ns, specv1.AppDeployTypeDaemonSet,
			appName, appVersion, serviceName, daemon.Spec.Selector.MatchLabels)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}
	var res []specv1.AppStats
	for _, stats := range appStats {
		res = append(res, stats)
	}
	return res, nil
}

func (k *kubeImpl) collectAppStats(appStats map[string]specv1.AppStats,
	ns, tp, appName, appVersion, serviceName string, set labels.Set) error {
	if appName == "" || serviceName == "" {
		return nil
	}
	stats, ok := appStats[appName]
	if !ok {
		stats = specv1.AppStats{
			AppInfo:    specv1.AppInfo{Name: appName, Version: appVersion},
			DeployType: tp,
		}
	}
	selector := labels.SelectorFromSet(set)
	pods, err := k.cli.core.Pods(ns).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return errors.Trace(err)
	}
	for _, pod := range pods.Items {
		if stats.InstanceStats == nil {
			stats.InstanceStats = map[string]specv1.InstanceStats{}
		}
		stats.InstanceStats[pod.Name] = k.collectInstanceStats(ns, serviceName, &pod)
	}
	if pods == nil || len(pods.Items) == 0 {
		return nil
	}
	stats.Status = getAppStatus(stats.InstanceStats)
	appStats[appName] = stats
	return nil
}

func getAppStatus(infos map[string]specv1.InstanceStats) specv1.Status {
	var pending = false
	for _, info := range infos {
		if info.Status == specv1.Pending {
			pending = true
		} else if info.Status == specv1.Failed {
			return info.Status
		}
	}
	if pending {
		return specv1.Pending
	}
	return specv1.Running
}

func (k *kubeImpl) collectInstanceStats(ns, serviceName string, pod *corev1.Pod) specv1.InstanceStats {
	stats := specv1.InstanceStats{Name: pod.Name, ServiceName: serviceName, Usage: map[string]string{}}
	stats.CreateTime = pod.CreationTimestamp.Local()
	stats.Status = specv1.Status(pod.Status.Phase)
	if stats.Status != specv1.Running {
		ref, err := reference.GetReference(scheme.Scheme, pod)
		if err != nil {
			k.log.Warn("failed to get service reference", log.Error(err))
			return stats
		}
		events, _ := k.cli.core.Events(ns).Search(scheme.Scheme, ref)
		if l := len(events.Items); l > 0 {
			if e := events.Items[l-1]; e.Type == "Warning" {
				stats.Cause += e.Message
			}
		}
	}

	for _, st := range pod.Status.ContainerStatuses {
		if st.State.Waiting != nil {
			stats.Status = specv1.Status(corev1.PodPending)
		}
	}
	podMetric, err := k.cli.metrics.PodMetricses(ns).Get(pod.Name, metav1.GetOptions{})
	if err != nil {
		k.log.Warn("failed to collect pod metrics", log.Error(err))
		return stats
	}
	for _, cont := range podMetric.Containers {
		if cont.Name == serviceName {
			for res, quan := range cont.Usage {
				stats.Usage[string(res)] = quan.String()
			}
		}
	}
	stats.IP = pod.Status.PodIP
	stats.NodeName = pod.Spec.NodeName
	return stats
}
