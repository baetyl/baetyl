package kube

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

func (k *kubeImpl) FetchLog(ns, pod, container string, tailLines, sinceSeconds int64) ([]byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := k.cli.core.Pods(ns).GetLogs(pod, k.toLogOptions(container, tailLines, sinceSeconds)).Do(ctx)
	return result.Raw()
}

func (k *kubeImpl) toLogOptions(container string, tailLines, sinceSeconds int64) *corev1.PodLogOptions {
	logOptions := &corev1.PodLogOptions{
		Follow:     k.conf.LogConfig.Follow,
		Previous:   k.conf.LogConfig.Previous,
		Timestamps: k.conf.LogConfig.TimeStamps,
	}
	if tailLines > 0 {
		logOptions.TailLines = &tailLines
	}
	if sinceSeconds > 0 {
		logOptions.SinceSeconds = &sinceSeconds
	}
	if container != "" {
		logOptions.Container = container
	}
	return logOptions
}
