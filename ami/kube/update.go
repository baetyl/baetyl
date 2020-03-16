package kube

import (
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/models"
	"github.com/baetyl/baetyl-core/utils"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *kubeEngine) UpdateApp(apps map[string]string) error {
	deploys := map[string]*appv1.Deployment{}
	var services []*v1.Service
	configs := map[string]*v1.ConfigMap{}
	for name, ver := range apps {
		var app models.Application
		key := utils.MakeKey(common.Application, name, ver)
		err := k.store.Get(key, &app)
		if err != nil {
			return err
		}
		deploy, svcs, err := ToDeployAndService(&app)
		if err != nil {
			return err
		}
		deploys[deploy.Name] = deploy
		services = append(services, svcs...)
		for _, v := range app.Volumes {
			if cfg := v.Configuration; cfg != nil {
				key := utils.MakeKey(common.Configuration, cfg.Name, cfg.Version)
				var config models.Configuration
				err := k.store.Get(key, &config)
				if err != nil {
					return err
				}
				configMap, err := ToConfigMap(&config)
				configs[config.Name] = configMap
			}
		}
	}

	configMapInterface := k.cli.CoreV1.ConfigMaps(k.cli.Namespace)
	for _, cfg := range configs {
		_, err := configMapInterface.Get(cfg.Name, metav1.GetOptions{})
		if err == nil {
			_, err := configMapInterface.Update(cfg)
			if err != nil {
				return err
			}
		} else {
			_, err := configMapInterface.Create(cfg)
			if err != nil {
				return err
			}
		}
	}

	deployInterface := k.cli.AppV1.Deployments(k.cli.Namespace)
	for _, d := range deploys {
		_, err := deployInterface.Get(d.Name, metav1.GetOptions{})
		if err == nil {
			_, err := deployInterface.Update(d)
			if err != nil {
				return err
			}
		} else {
			_, err := deployInterface.Create(d)
			if err != nil {
				return err
			}
		}
	}

	serviceInterface := k.cli.CoreV1.Services(k.cli.Namespace)
	for _, s := range services {
		_, err := serviceInterface.Get(s.Name, metav1.GetOptions{})
		if err == nil {
			_, err := serviceInterface.Update(s)
			if err != nil {
				return err
			}
		} else {
			_, err := serviceInterface.Create(s)
			if err != nil {
				return err
			}
		}
	}

	return k.store.Upsert(common.DefaultAppsKey,
		config.AppsVersionResource{Name: common.DefaultAppsKey, Value: apps})
}
