package kube

import (
	"github.com/baetyl/baetyl-go/v2/errors"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	appv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	v2 "k8s.io/client-go/kubernetes/typed/autoscaling/v2"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientset "k8s.io/metrics/pkg/client/clientset/versioned"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"

	"github.com/baetyl/baetyl/v2/config"
)

type client struct {
	kubeConfig *rest.Config
	core       corev1.CoreV1Interface
	app        appv1.AppsV1Interface
	batch      batchv1.BatchV1Interface
	metrics    metricsv1beta1.MetricsV1beta1Interface
	discovery  discovery.DiscoveryInterface
	autoscale  v2.AutoscalingV2Interface
}

func newClient(cfg config.KubeConfig) (*client, error) {
	kubeConfig, err := func() (*rest.Config, error) {
		if !cfg.OutCluster {
			return rest.InClusterConfig()
		}
		return clientcmd.BuildConfigFromFlags("", cfg.ConfPath)
	}()
	if err != nil {
		return nil, errors.Trace(err)
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.Trace(err)
	}

	metricsCli, err := clientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &client{
		kubeConfig: kubeConfig,
		core:       kubeClient.CoreV1(),
		app:        kubeClient.AppsV1(),
		batch:      kubeClient.BatchV1(),
		metrics:    metricsCli.MetricsV1beta1(),
		discovery:  kubeClient.Discovery(),
		autoscale:  kubeClient.AutoscalingV2(),
	}, nil
}
