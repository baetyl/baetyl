package ami

import (
	"os"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-go/log"
	bh "github.com/timshannon/bolthold"
	"k8s.io/client-go/kubernetes"
	appv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientset "k8s.io/metrics/pkg/client/clientset/versioned"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

type kubeImpl struct {
	knn   string // kube node name
	cli   *Client
	store *bh.Store
	log   *log.Logger
}

// TODO: move store and shadow to engine. kubemodel only implement the interfaces of omi
func NewKubeImpl(cfg config.KubernetesConfig, sto *bh.Store) (AMI, error) {
	cli, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	knn := os.Getenv("KUBE_NODE_NAME")
	model := &kubeImpl{
		cli:   cli,
		store: sto,
		knn:   knn,
		log:   log.With(log.Any("ami", "kube")),
	}
	return model, nil
}

type Client struct {
	Namespace string
	Core      corev1.CoreV1Interface
	App       appv1.AppsV1Interface
	Metrics   metricsv1beta1.MetricsV1beta1Interface
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
