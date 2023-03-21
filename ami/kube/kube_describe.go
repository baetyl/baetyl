package kube

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/describe"
)

var (
	SchemaKindPod         = schema.GroupKind{Group: corev1.GroupName, Kind: "Pod"}
	SchemaKindSecret      = schema.GroupKind{Group: corev1.GroupName, Kind: "Secret"}
	SchemaKindService     = schema.GroupKind{Group: corev1.GroupName, Kind: "Service"}
	SchemaKindNode        = schema.GroupKind{Group: corev1.GroupName, Kind: "Node"}
	SchemaKindConfigMap   = schema.GroupKind{Group: corev1.GroupName, Kind: "ConfigMap"}
	SchemaKindJob         = schema.GroupKind{Group: batchv1.GroupName, Kind: "Job"}
	SchemaKindCronJob     = schema.GroupKind{Group: batchv1.GroupName, Kind: "CronJob"}
	SchemaKindStatefulSet = schema.GroupKind{Group: appsv1.GroupName, Kind: "StatefulSet"}
	SchemaKindDeployment  = schema.GroupKind{Group: appsv1.GroupName, Kind: "Deployment"}
	SchemaKindDaemonSet   = schema.GroupKind{Group: appsv1.GroupName, Kind: "DaemonSet"}
)

func (k *kubeImpl) RemoteDescribe(tp, ns, n string) (string, error) {
	var sm schema.GroupKind
	switch strings.TrimSpace(strings.ToLower(tp)) {
	case "pod", "pods", "po":
		sm = SchemaKindPod
	case "deploy", "deployment", "deployments":
		sm = SchemaKindDeployment
	case "sts", "statefulsets", "statefulset":
		sm = SchemaKindStatefulSet
	case "ds", "daemonsets", "daemonset":
		sm = SchemaKindDaemonSet
	case "cm", "configmap", "configmaps":
		sm = SchemaKindConfigMap
	case "secrets", "secret":
		sm = SchemaKindSecret
	case "svc", "services", "service":
		sm = SchemaKindService
	case "job", "jobs":
		sm = SchemaKindJob
	case "cronjobs", "cronjob":
		sm = SchemaKindCronJob
	case "no", "nodes", "node":
		sm = SchemaKindNode
	default:
		return "", fmt.Errorf("describe schema type (%s) not support", tp)
	}
	desc, ok := describe.DescriberFor(sm, k.cli.kubeConfig)
	if !ok {
		return "", fmt.Errorf("describe type (%+v) not support", sm)
	}
	return desc.Describe(ns, n, describe.DescriberSettings{
		ShowEvents: true,
		ChunkSize:  cmdutil.DefaultChunkSize,
	})
}
