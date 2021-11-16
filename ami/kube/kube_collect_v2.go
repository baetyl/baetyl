package kube

import (
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/reference"
	"k8s.io/kubectl/pkg/scheme"
)

func (k *kubeImpl) collectDeploymentStatsV2(ns string, qps map[string]interface{}) ([]specv1.AppStats, error) {
	deploys, err := k.cli.app.Deployments(ns).List(metav1.ListOptions{})
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
		err = k.collectAppStatsV2(appStats, qps, ns, info)
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

func (k *kubeImpl) collectDaemonSetStatsV2(ns string, qps map[string]interface{}) ([]specv1.AppStats, error) {
	daemons, err := k.cli.app.DaemonSets(ns).List(metav1.ListOptions{})
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
		err = k.collectAppStatsV2(appStats, qps, ns, info)
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

func (k *kubeImpl) collectJobStatsV2(ns string, qps map[string]interface{}) ([]specv1.AppStats, error) {
	jobs, err := k.cli.batch.Jobs(ns).List(metav1.ListOptions{})
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
		err = k.collectAppStatsV2(appStats, qps, ns, info)
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

func (k *kubeImpl) collectAppStatsV2(appStats map[string]specv1.AppStats, qps map[string]interface{}, ns string, info appInfo) error {
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
	pods, err := k.cli.core.Pods(ns).List(metav1.ListOptions{LabelSelector: labels.SelectorFromSet(info.set).String()})
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
		stats.InstanceStats[pod.Name] = k.collectInstanceStatsV2(ns, info.name, qps, &pod)
		insStats[pod.Name] = stats.InstanceStats[pod.Name]

	}
	stats.Status = getAppStatus(stats.Status, info.replicas, insStats)
	appStats[info.name] = stats
	return nil
}

func (k *kubeImpl) collectInstanceStatsV2(ns, appName string, qps map[string]interface{}, pod *corev1.Pod) specv1.InstanceStats {
	stats := specv1.InstanceStats{Name: pod.Name, AppName: appName, Usage: map[string]string{}}
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
		stats.Containers = append(stats.Containers, containerInfo)
	}

	for key, val := range usageTotal {
		stats.Usage[key] = val.String()
	}

	if qpsStats, ok := qps[pod.Name]; ok {
		stats.Extension = qpsStats
	}

	stats.IP = pod.Status.PodIP
	stats.NodeName = pod.Spec.NodeName
	return stats
}
