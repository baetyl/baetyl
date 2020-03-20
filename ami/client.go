package ami

import (
	"github.com/baetyl/baetyl-core/config"
	"k8s.io/client-go/kubernetes"
	appv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientset "k8s.io/metrics/pkg/client/clientset/versioned"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type Client struct {
	Core      corev1.CoreV1Interface
	App       appv1.AppsV1Interface
	Metrics   metricsv1beta1.MetricsV1beta1Interface
	Namespace string
}

func NewClient(cfg config.KubernetesConfig) (*Client, error) {
	kubeConfig, err := func() (*rest.Config, error) {
		if cfg.InCluster {
			return rest.InClusterConfig()

		}
		return clientcmd.BuildConfigFromFlags(
			"", cfg.ConfigPath)
	}()
	if err != nil {
		return nil, err
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}

	metricsCli, err := clientset.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return &Client{
		Core:      kubeClient.CoreV1(),
		App:       kubeClient.AppsV1(),
		Metrics:   metricsCli.MetricsV1beta1(),
		Namespace: "default", // TODO: check
	}, nil
}
