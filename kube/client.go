package kube

import (
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/config"
	"k8s.io/client-go/kubernetes"
	appv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	CoreV1    corev1.CoreV1Interface
	AppV1     appv1.AppsV1Interface
	Namespace string
}

func NewClient(cfg config.APIServer) (*Client, error) {
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
	return &Client{
		CoreV1:    kubeClient.CoreV1(),
		AppV1:     kubeClient.AppsV1(),
		Namespace: common.DefaultNamespace,
	}, nil
}
