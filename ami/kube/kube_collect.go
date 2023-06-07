package kube

import (
	"context"

	gctx "github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/imdario/mergo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"

	"github.com/baetyl/baetyl/v2/ami"
)

type appInfo struct {
	name    string
	version string
	// Deprecated: Field svcName is no longer used.
	svcName  string
	typ      string
	replicas int32
	set      labels.Set
}

func (k *kubeImpl) GetModeInfo() (interface{}, error) {
	info, err := k.cli.discovery.ServerVersion()
	if err != nil {
		return nil, err
	}
	return info.String(), nil
}

func (k *kubeImpl) CollectNodeInfo() (map[string]interface{}, error) {
	nodes, err := k.cli.core.Nodes().List(context.TODO(), metav1.ListOptions{})
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
			Labels:           node.GetLabels(),
		}
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeHostName {
				nodeInfo.Hostname = addr.Address
			}
			if addr.Type == corev1.NodeInternalIP {
				nodeInfo.Address = addr.Address
			}
		}
		for k := range node.GetLabels() {
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
	nodes, err := k.cli.core.Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}
	nodeMetrics, err := k.cli.metrics.NodeMetricses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}

	metrics := map[string]v1beta1.NodeMetrics{}
	for _, metric := range nodeMetrics.Items {
		metrics[metric.Name] = metric
	}
	var gpuExts map[string]interface{}
	if extension, ok := ami.Hooks[ami.BaetylGPUStatsExtension]; ok {
		collectStatsExt, ok := extension.(ami.CollectStatsExtFunc)
		if ok {
			gpuExts, err = collectStatsExt(gctx.RunModeKube)
			if err != nil {
				k.log.Warn("failed to collect gpu stats", log.Error(errors.Trace(err)))
			}
			k.log.Debug("collect gpu stats successfully", log.Any("gpuStats", gpuExts))
		} else {
			k.log.Warn("invalid collecting gpu stats function")
		}
	}
	var nodeExts map[string]interface{}
	if nodeExtHook, ok := ami.Hooks[ami.BaetylNodeStatsExtension]; ok {
		nodeStatsExt, ok := nodeExtHook.(ami.CollectStatsExtFunc)
		if ok {
			nodeExts, err = nodeStatsExt(gctx.RunModeKube)
			if err != nil {
				k.log.Warn("failed to collect node stats", log.Error(errors.Trace(err)))
			}
			k.log.Debug("collect node stats successfully", log.Any("nodeStats", nodeExts))
		} else {
			k.log.Warn("invalid collecting node stats function")
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

		var nodeStatsMerge map[string]interface{}
		if len(gpuExts) > 0 {
			if ext, ok := gpuExts[node.Name]; ok && ext != nil {
				nodeStatsMerge = ext.(map[string]interface{})
			}
		}
		if len(nodeExts) > 0 {
			if ext, ok := nodeExts[node.Name]; ok && ext != nil {
				if nodeStatsMerge == nil {
					nodeStatsMerge = make(map[string]interface{}, 0)
				}
				if err := mergo.Merge(&nodeStatsMerge, ext.(map[string]interface{})); err != nil {
					k.log.Warn("fail to merge node stats and node gpu stats", log.Error(err))
				}
			}
		}

		nodeStats.Extension = nodeStatsMerge
		infos[node.Name] = nodeStats
	}
	return infos, nil
}

func (k *kubeImpl) collectAppStats(appStats map[string]specv1.AppStats, qps map[string]interface{}, ns string, info appInfo) error {
	if info.name == "" {
		return nil
	}
	stats, ok := appStats[info.name]
	if !ok {
		stats = specv1.AppStats{
			AppInfo:    specv1.AppInfo{Name: info.name, Version: info.version},
			DeployType: info.typ,
		}
	}
	pods, err := k.cli.core.Pods(ns).List(context.TODO(), metav1.ListOptions{LabelSelector: labels.SelectorFromSet(info.set).String()})
	if err != nil {
		return errors.Trace(err)
	}
	if pods == nil || len(pods.Items) == 0 {
		return nil
	}
	if stats.InstanceStats == nil {
		stats.InstanceStats = map[string]specv1.InstanceStats{}
	}
	insStats := map[string]specv1.InstanceStats{}
	for _, pod := range pods.Items {
		stats.InstanceStats[pod.Name] = k.collectInstanceStats(ns, info.name, qps, &pod)
		insStats[pod.Name] = stats.InstanceStats[pod.Name]

	}
	stats.Status = getAppStatus(stats.Status, info.replicas, insStats)
	appStats[info.name] = stats
	return nil
}

func (k *kubeImpl) collectDeploymentStats(ns string, qps map[string]interface{}) ([]specv1.AppStats, error) {
	deploys, err := k.cli.app.Deployments(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}
	appStats := map[string]specv1.AppStats{}
	info := appInfo{typ: specv1.WorkloadDeployment}
	for _, deploy := range deploys.Items {
		info.name = deploy.Labels[AppName]
		info.version = deploy.Labels[AppVersion]
		info.replicas = *deploy.Spec.Replicas
		info.set = deploy.Spec.Selector.MatchLabels
		err = k.collectAppStats(appStats, qps, ns, info)
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

func (k *kubeImpl) collectDaemonSetStats(ns string, qps map[string]interface{}) ([]specv1.AppStats, error) {
	daemons, err := k.cli.app.DaemonSets(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}
	appStats := map[string]specv1.AppStats{}
	info := appInfo{typ: specv1.WorkloadDaemonSet}
	for _, daemon := range daemons.Items {
		info.name = daemon.Labels[AppName]
		info.version = daemon.Labels[AppVersion]
		info.set = daemon.Spec.Selector.MatchLabels
		info.replicas = daemon.Status.DesiredNumberScheduled
		err = k.collectAppStats(appStats, qps, ns, info)
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

func (k *kubeImpl) collectJobStats(ns string, qps map[string]interface{}) ([]specv1.AppStats, error) {
	jobs, err := k.cli.batch.Jobs(ns).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Trace(err)
	}
	appStats := map[string]specv1.AppStats{}
	info := appInfo{typ: specv1.WorkloadJob}
	for _, job := range jobs.Items {
		info.name = job.Labels[AppName]
		info.version = job.Labels[AppVersion]
		info.replicas = *job.Spec.Completions
		info.set = job.Spec.Selector.MatchLabels
		err = k.collectAppStats(appStats, qps, ns, info)
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

func (k *kubeImpl) collectInstanceStats(ns, appName string, qps map[string]interface{}, pod *corev1.Pod) specv1.InstanceStats {
	stats := specv1.InstanceStats{Name: pod.Name, AppName: appName, Usage: map[string]string{}}
	stats.CreateTime = pod.CreationTimestamp.Local()
	stats.Status = specv1.Status(pod.Status.Phase)
	stats.Cause = pod.Status.Reason

	for _, initStatus := range pod.Status.InitContainerStatuses {
		containerInfo := specv1.ContainerInfo{Name: initStatus.Name}
		containerInfo.State, containerInfo.Reason = getContainerStatus(&initStatus)
		stats.InitContainers = append(stats.InitContainers, containerInfo)
	}

	metricsStatus := make(map[string]specv1.ContainerInfo)
	podMetric, err := k.cli.metrics.PodMetricses(ns).Get(context.TODO(), pod.Name, metav1.GetOptions{})
	if err != nil {
		k.log.Warn("failed to collect pod metrics", log.Error(err))
	} else {
		usageTotal := map[string]*resource.Quantity{}
		for _, cont := range podMetric.Containers {
			containerInfo := specv1.ContainerInfo{
				Name:  cont.Name,
				Usage: map[string]string{},
			}
			for res, quan := range cont.Usage {
				containerInfo.Usage[string(res)] = quan.String()
				if _, ok := usageTotal[string(res)]; ok {
					usageTotal[string(res)].Add(quan)
				} else {
					var v resource.Quantity
					quan.DeepCopyInto(&v)
					usageTotal[string(res)] = &v
				}
			}
			metricsStatus[cont.Name] = containerInfo
		}
		for key, val := range usageTotal {
			stats.Usage[key] = val.String()
		}
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		containerInfo := specv1.ContainerInfo{Name: containerStatus.Name}
		containerInfo.State, containerInfo.Reason = getContainerStatus(&containerStatus)
		if metrics, ok := metricsStatus[containerStatus.Name]; ok {
			containerInfo.Usage = metrics.Usage
		}
		stats.Containers = append(stats.Containers, containerInfo)
	}

	if qpsStats, ok := qps[pod.Name]; ok {
		stats.Extension = qpsStats
	}

	stats.IP = pod.Status.PodIP
	stats.NodeName = pod.Spec.NodeName
	return stats
}

func getContainerStatus(info *corev1.ContainerStatus) (specv1.ContainerState, string) {
	if info.State.Waiting != nil {
		return specv1.ContainerWaiting, info.State.Waiting.Reason
	}
	if info.State.Running != nil {
		return specv1.ContainerRunning, ""
	}
	if info.State.Terminated != nil {
		return specv1.ContainerTerminated, info.State.Terminated.Reason
	}
	return specv1.ContainerWaiting, "status unknown"
}

func getAppStatus(status specv1.Status, replicas int32, insStats map[string]specv1.InstanceStats) specv1.Status {
	var cnt int32
	pending, unknown := false, false
	for _, ins := range insStats {
		switch ins.Status {
		case specv1.Pending, specv1.Failed:
			pending = true
		case specv1.Unknown:
			unknown = true
		case specv1.Running, specv1.Succeeded:
			cnt++
		default:
		}
	}
	var res = specv1.Pending
	if cnt == replicas {
		res = specv1.Running
	} else {
		if pending {
			res = specv1.Pending
		}
		if unknown {
			res = specv1.Unknown
		}
	}
	if status == "" || status == specv1.Running {
		return res
	} else {
		return status
	}
}
