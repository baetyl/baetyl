package ami

import (
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/utils"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/jinzhu/copier"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (k *kubeModel) ApplyApplications(apps map[string]interface{}) error {
	deploys := map[string]*appv1.Deployment{}
	var services []*corev1.Service
	configs := map[string]*corev1.ConfigMap{}
	deployInterface := k.cli.App.Deployments(k.cli.Namespace)
	for name, ver := range apps {
		if ver.(string) == "" {
			err := deployInterface.Delete(name, &metav1.DeleteOptions{})
			if err != nil {
				return err
			}
			continue
		}
		var app v1.Application
		key := utils.MakeKey(common.Application, name, ver.(string))
		err := k.store.Get(key, &app)
		if err != nil {
			return err
		}
		deploy, svcs, err := toDeployAndService(&app)
		if err != nil {
			return err
		}
		deploys[deploy.Name] = deploy
		services = append(services, svcs...)
		for _, v := range app.Volumes {
			if cfg := v.Config; cfg != nil {
				key := utils.MakeKey(common.Configuration, cfg.Name, cfg.Version)
				var config v1.Configuration
				err := k.store.Get(key, &config)
				if err != nil {
					return err
				}
				configMap, err := toConfigMap(&config)
				configs[config.Name] = configMap
			}
		}
	}

	configMapInterface := k.cli.Core.ConfigMaps(k.cli.Namespace)
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

	serviceInterface := k.cli.Core.Services(k.cli.Namespace)
	for _, s := range services {
		service, err := serviceInterface.Get(s.Name, metav1.GetOptions{})
		if err == nil {
			s.ResourceVersion = service.ResourceVersion
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
		appsVersionResource{Name: common.DefaultAppsKey, Value: apps})
}

func toConfigMap(config *v1.Configuration) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	err := copier.Copy(configMap, config)
	if err != nil {
		return nil, err
	}
	return configMap, nil
}

func toDeployAndService(app *v1.Application) (*appv1.Deployment, []*corev1.Service, error) {
	deploy := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
		Spec: appv1.DeploymentSpec{
			Strategy: appv1.DeploymentStrategy{
				Type: appv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"baetyl": app.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
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

	for _, v := range app.Volumes {
		if config := v.Config; config != nil {
			volume := corev1.Volume{
				Name:         v.Name,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: config.Name},
					},
				},
			}
			deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, volume)
		} else if secret := v.Secret; secret != nil {
			volume := corev1.Volume{
				Name:         v.Name,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource {
						SecretName: secret.Name,
					},
				},
			}
			deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, volume)
		}
	}

	var services []*corev1.Service
	for _, svc := range app.Services {
		if len(svc.Ports) == 0 {
			continue
		}
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      svc.Name,
				Namespace: app.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"baetyl": app.Name,
				},
				ClusterIP: "None",
			},
		}
		for _, p := range svc.Ports {
			port := corev1.ServicePort{
				Port:       p.ContainerPort,
				TargetPort: intstr.IntOrString{IntVal: p.ContainerPort},
			}
			service.Spec.Ports = append(service.Spec.Ports, port)
		}
		services = append(services, service)
	}
	return deploy, services, nil
}
