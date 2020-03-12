package kube

import (
	"encoding/json"
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
	cli    *kube.Client
	driver store.Driver
}

func NewEngine(cli *kube.Client, driver store.Driver) *kubeEngine {
	return &kubeEngine{cli: cli, driver: driver}
}

func (k *kubeEngine) Start() error {
	data, err := k.driver.Get([]byte(common.DefaultAppsKey))
	if err != nil {
		return err
	}
	var apps map[string]string
	err = json.Unmarshal(data, &apps)
	if err != nil {
		return err
	}

	err = k.UpdateApp(apps)
	if err != nil {
		return err
	}
	return nil
}

func configurationToConfigMap(config *models.Configuration) (*v1.ConfigMap, error) {
	configMap := &v1.ConfigMap{}
	err := copier.Copy(configMap, config)
	if err != nil {
		return nil, err
	}
	return configMap, nil
}
