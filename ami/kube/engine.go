package kube

import (
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/kube"
	"github.com/baetyl/baetyl-core/models"
	"github.com/baetyl/baetyl-core/store"
	"github.com/jinzhu/copier"
	v1 "k8s.io/api/core/v1"
)

type Engine interface {
	UpdateApp(apps map[string]string) error
	Start() error
}

type kubeEngine struct {
	cli   *kube.Client
	store store.Store
}

func NewEngine(cli *kube.Client, store store.Store) *kubeEngine {
	return &kubeEngine{cli: cli, store: store}
}

func (k *kubeEngine) Start() error {
	var apps map[string]string
	err := k.store.Get(common.DefaultAppsKey, &apps)
	if err != nil {
		return err
	}
	// TODO: why to get here, then to update in update.go?
	return k.UpdateApp(apps)
}

func configurationToConfigMap(config *models.Configuration) (*v1.ConfigMap, error) {
	configMap := &v1.ConfigMap{}
	err := copier.Copy(configMap, config)
	if err != nil {
		return nil, err
	}
	return configMap, nil
}
