package ami

import (
	"github.com/baetyl/baetyl-go/spec/crd"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/jinzhu/copier"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kl "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	AppName     = "baetyl-app-name"
	AppVersion  = "baetyl-app-version"
	ServiceName = "baetyl-service-name"

	RegistryAddress  = "address"
	RegistryUsername = "username"
	RegistryPassword = "password"
)

func (k *kubeImpl) Apply(appInfos []specv1.AppInfo) error {
	appMap := map[string]string{}
	configs := map[string]*corev1.ConfigMap{}
	secrets := map[string]*corev1.Secret{}
	services := map[string]*corev1.Service{}
	deploys := map[string]*appv1.Deployment{}
	for _, info := range appInfos {
		appMap[info.Name] = info.Version
		key := makeKey(crd.KindApplication, info.Name, info.Version)
		var app crd.Application
		err := k.store.Get(key, &app)
		if err != nil {
			return err
		}

		var imagePullSecrets []corev1.LocalObjectReference
		for _, v := range app.Volumes {
			if cfg := v.Config; cfg != nil {
				key := makeKey(crd.KindConfiguration, cfg.Name, cfg.Version)
				var config crd.Configuration
				err := k.store.Get(key, &config)
				if err != nil {
					return err
				}
				configMap, err := k.toConfigMap(&config)
				if err != nil {
					return err
				}
				configs[config.Name] = configMap
			}

			if sec := v.Secret; sec != nil {
				key := makeKey(crd.KindSecret, sec.Name, sec.Version)
				var secret crd.Secret
				err := k.store.Get(key, &secret)
				if err != nil {
					return err
				}

				if isRegistrySecret(&secret) {
					imagePullSecrets = append(imagePullSecrets,
						corev1.LocalObjectReference{
							Name: secret.Name,
						})
					v.Secret = nil
				}

				kSecret, err := k.toSecret(&secret)
				if err != nil {
					return err
				}
				secrets[kSecret.Name] = kSecret

			}
		}

		for _, svc := range app.Services {
			deploy, err := k.toDeploy(&app, &svc, app.Volumes, imagePullSecrets)
			if err != nil {
				return err
			}
			deploys[deploy.Name] = deploy
			service, err := k.toService(&svc)
			if err != nil {
				return err
			}
			if service != nil {
				services[service.Name] = service
			}
		}

	}

	if err := k.applyConfigMaps(configs); err != nil {
		return err
	}
	if err := k.applySecrets(secrets); err != nil {
		return err
	}
	if err := k.applyDeploys(deploys); err != nil {
		return err
	}
	if err := k.applyServices(services); err != nil {
		return err
	}
	return nil
}

func (k *kubeImpl) applyDeploys(deploys map[string]*appv1.Deployment) error {
	deployInterface := k.cli.App.Deployments(k.cli.Namespace)
	ls := kl.Set{}
	selector := map[string]string{
		"baetyl": "baetyl",
	}
	err := copier.Copy(&ls, &selector)
	if err != nil {
		return err
	}
	deployList, err := k.cli.App.Deployments(k.cli.Namespace).List(metav1.ListOptions{
		LabelSelector: ls.String(),
	})
	if err != nil {
		return err
	}
	deletes := map[string]struct{}{}
	if deployList != nil {
		for _, d := range deployList.Items {
			if _, ok := deploys[d.Name]; !ok {
				deletes[d.Name] = struct{}{}
			}
		}
	}
	for n := range deletes {
		err := deployInterface.Delete(n, &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	for _, d := range deploys {
		deploy, err := deployInterface.Get(d.Name, metav1.GetOptions{})
		if deploy != nil && err == nil {
			d.ResourceVersion = deploy.ResourceVersion
			_, err = deployInterface.Update(d)
			if err != nil {
				return err
			}
		} else {
			_, err = deployInterface.Create(d)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (k *kubeImpl) applyServices(services map[string]*corev1.Service) error {
	serviceInterface := k.cli.Core.Services(k.cli.Namespace)
	for _, s := range services {
		service, err := serviceInterface.Get(s.Name, metav1.GetOptions{})
		if service != nil && err == nil {
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
	return nil
}

func (k *kubeImpl) applyConfigMaps(configMaps map[string]*corev1.ConfigMap) error {
	configMapInterface := k.cli.Core.ConfigMaps(k.cli.Namespace)
	for _, cfg := range configMaps {
		config, err := configMapInterface.Get(cfg.Name, metav1.GetOptions{})
		if config != nil && err == nil {
			cfg.ResourceVersion = config.ResourceVersion
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
	return nil
}

func (k *kubeImpl) applySecrets(secrets map[string]*corev1.Secret) error {
	secretInterface := k.cli.Core.Secrets(k.cli.Namespace)
	for _, sec := range secrets {
		secret, err := secretInterface.Get(sec.Name, metav1.GetOptions{})
		if secret != nil && err == nil {
			sec.ResourceVersion = secret.ResourceVersion
			_, err := secretInterface.Update(sec)
			if err != nil {
				return err
			}
		} else {
			_, err := secretInterface.Create(sec)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (k *kubeImpl) toDeploy(app *crd.Application, service *crd.Service, vols []crd.Volume,
	imagePullSecrets []corev1.LocalObjectReference) (*appv1.Deployment, error) {
	volMap := map[string]crd.Volume{}
	for _, v := range vols {
		volMap[v.Name] = v
	}
	var c corev1.Container
	err := copier.Copy(&c, &service)
	if err != nil {
		return nil, err
	}
	if service.Resources != nil {
		c.Resources.Limits = corev1.ResourceList{}
		for n, value := range service.Resources.Limits {
			quantity, err := resource.ParseQuantity(value)
			if err != nil {
				return nil, err
			}
			c.Resources.Limits[corev1.ResourceName(n)] = quantity
		}
	}
	var containers []corev1.Container
	containers = append(containers, c)
	var volumes []corev1.Volume

	for _, v := range service.VolumeMounts {
		vol := volMap[v.Name]
		volume := corev1.Volume{
			Name: v.Name,
		}
		if vol.Config != nil {
			volume.VolumeSource.ConfigMap = &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: vol.VolumeSource.Config.Name},
			}
		} else if vol.Secret != nil {
			volume.VolumeSource.Secret = &corev1.SecretVolumeSource{
				SecretName: vol.VolumeSource.Secret.Name,
			}
		} else if vol.HostPath != nil {
			volume.VolumeSource.HostPath = &corev1.HostPathVolumeSource{
				Path: vol.VolumeSource.HostPath.Path,
			}
		}
		volumes = append(volumes, volume)
	}
	restartPolicy := corev1.RestartPolicyAlways
	if service.Restart != nil {
		restartPolicy = corev1.RestartPolicy(service.Restart.Policy)
	}
	replica := new(int32)
	*replica = int32(service.Replica)
	deploy := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.Name,
			Namespace: k.cli.Namespace,
			Labels: map[string]string{
				"baetyl": "baetyl",
			},
		},
		Spec: appv1.DeploymentSpec{
			Replicas: replica,
			Strategy: appv1.DeploymentStrategy{
				Type: appv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					ServiceName: service.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
					"baetyl":    "baetyl",
					AppName:     app.Name,
					AppVersion:  app.Version,
					ServiceName: service.Name,
				}},
				Spec: corev1.PodSpec{
					Volumes:          volumes,
					Containers:       containers,
					RestartPolicy:    restartPolicy,
					ImagePullSecrets: imagePullSecrets,
				},
			},
		},
	}
	return deploy, nil
}

func (k *kubeImpl) toConfigMap(config *crd.Configuration) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	err := copier.Copy(configMap, config)
	if err != nil {
		return nil, err
	}
	configMap.Namespace = k.cli.Namespace
	return configMap, nil
}

func (k *kubeImpl) toSecret(sec *crd.Secret) (*corev1.Secret, error) {
	// secret for docker config
	if isRegistrySecret(sec) {
		return k.generateRegistrySecret(sec.Name, string(sec.Data[RegistryAddress]),
			string(sec.Data[RegistryUsername]), string(sec.Data[RegistryPassword]))
	}

	// common secret
	secret := &corev1.Secret{}
	err := copier.Copy(secret, sec)
	if err != nil {
		return nil, err
	}
	secret.Namespace = k.cli.Namespace
	return secret, nil
}

func (k *kubeImpl) toService(svc *crd.Service) (*corev1.Service, error) {
	if len(svc.Ports) == 0 {
		return nil, nil
	}
	var ports []corev1.ServicePort
	for _, p := range svc.Ports {
		port := corev1.ServicePort{
			Port:       p.ContainerPort,
			TargetPort: intstr.IntOrString{IntVal: p.ContainerPort},
		}
		ports = append(ports, port)
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svc.Name,
			Namespace: k.cli.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"baetyl": svc.Name,
			},
			ClusterIP: "None",
			Ports:     ports,
		},
	}
	return service, nil
}

func makeKey(kind crd.Kind, name, ver string) string {
	if name == "" || ver == "" {
		return ""
	}
	return string(kind) + "-" + name + "-" + ver
}

func isRegistrySecret(secret *crd.Secret) bool {
	registry, ok := secret.Labels[crd.SecretLabel]

	return ok && registry == crd.SecretRegistry
}
