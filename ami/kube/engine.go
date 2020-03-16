package kube

import (
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/kube"
	"github.com/baetyl/baetyl-core/models"
	"github.com/baetyl/baetyl-core/store"
	"github.com/jinzhu/copier"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	var apps config.AppsVersionResource
	err := k.store.Get(common.DefaultAppsKey, &apps)
	if err != nil {
		return err
	}
	// TODO: why to get here, then to update in update.go?
	return k.UpdateApp(apps.Value)
}

func ToConfigMap(config *models.Configuration) (*v1.ConfigMap, error) {
	configMap := &v1.ConfigMap{}
	err := copier.Copy(configMap, config)
	if err != nil {
		return nil, err
	}
	return configMap, nil
}

func ToDeployAndService(app *models.Application) (*appv1.Deployment, []*v1.Service, error) {
	deploy := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:                       app.Name,
			Namespace:                  app.Namespace,
		},
		Spec: appv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"baetyl": app.Name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
					"baetyl": app.Name,
				}},
			},
		},
	}

	err := copier.Copy(&deploy.Spec.Template.Spec.Containers, &app.Services)
	if err != nil {
		return nil, nil, err
	}
	err = copier.Copy(&deploy.Spec.Template.Spec.Volumes, &app.Volumes)
	if err != nil {
		return nil, nil, err
	}

	var services []*v1.Service
	for _, svc := range app.Services {
		if len(svc.Ports) == 0 {
			continue
		}
		service := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      svc.Name,
				Namespace: app.Namespace,
			},
			Spec: v1.ServiceSpec{
				Selector: map[string]string{
					"baetyl": app.Name,
				},
				ClusterIP: "None",
			},
		}
		for _, p := range svc.Ports {
			port := v1.ServicePort{
				Port:       p.ContainerPort,
				TargetPort: intstr.IntOrString{IntVal: p.ContainerPort},
			}
			service.Spec.Ports = append(service.Spec.Ports, port)
		}
		services = append(services, service)
	}
	return deploy, services, nil
}
