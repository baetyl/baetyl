// Package qpsstats qps监控实现
package qpsstats

import (
	"github.com/baetyl/baetyl-go/v2/errors"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type client struct {
	kubeConfig *rest.Config
	core       corev1.CoreV1Interface
}

func newClient(cfg KubeConfig) (*client, error) {
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

	return &client{
		kubeConfig: kubeConfig,
		core:       kubeClient.CoreV1(),
	}, nil
}
