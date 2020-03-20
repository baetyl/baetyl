package ami

import (
	"github.com/baetyl/baetyl-go/spec/api"
	"github.com/baetyl/baetyl-go/spec/crd"
	"github.com/jinzhu/copier"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (k *kubeModel) ApplyApplications(apps *api.ReportResponse) error {
	deploys := map[string]*appv1.Deployment{}
	var services []*corev1.Service
	configs := map[string]*corev1.ConfigMap{}
	deployInterface := k.cli.App.Deployments(k.cli.Namespace)
	for _, app := range apps.AppInfos {
		key := makeKey(crd.KindApplication, app.Name, app.Version)

		var appdata crd.Application
		err := k.store.Get(key, &appdata)
		if err != nil {
			return err
		}
		deploy, svcs, err := toDeployAndService(&appdata)
		if err != nil {
			return err
		}
		deploys[deploy.Name] = deploy
		services = append(services, svcs...)
		for _, v := range appdata.Volumes {
			if cfg := v.Config; cfg != nil {
				key := makeKey(crd.KindConfiguration, cfg.Name, cfg.Version)
				var config crd.Configuration
				err := k.store.Get(key, &config)
				if err != nil {
					return err
				}
				configMap, err := toConfigMap(&config)
				configs[config.Name] = configMap
			}
		}
	}

	// TODO: delete removed services

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

	return k.store.Upsert("apps", apps)
}

func toConfigMap(config *crd.Configuration) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	err := copier.Copy(configMap, config)
	if err != nil {
		return nil, err
	}
	return configMap, nil
}

func toDeployAndService(app *crd.Application) (*appv1.Deployment, []*corev1.Service, error) {
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
				Name: v.Name,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{Name: config.Name},
					},
				},
			}
			deploy.Spec.Template.Spec.Volumes = append(deploy.Spec.Template.Spec.Volumes, volume)
		} else if secret := v.Secret; secret != nil {
			volume := corev1.Volume{
				Name: v.Name,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
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

func makeKey(kind crd.Kind, name, ver string) string {
	if name == "" || ver == "" {
		return ""
	}
	return string(kind) + "/" + name + "/" + ver
}
