package ami

import (
	"strings"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/jinzhu/copier"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	KubeNodeName = "KUBE_NODE_NAME"
	AppName      = "baetyl-app-name"
	AppVersion   = "baetyl-app-version"
	ServiceName  = "baetyl-service-name"

	RegistryAddress  = "address"
	RegistryUsername = "username"
	RegistryPassword = "password"

	ServiceAccountName = "baetyl-edge-system-service-account"
	BaetylCore         = "baetyl-core"
)

func (k *kubeImpl) ApplyConfigurations(ns string, cfgs map[string]specv1.Configuration) error {
	for _, cfg := range cfgs {
		cm := &corev1.ConfigMap{}
		if err := copier.Copy(cm, &cfg); err != nil {
			return errors.Trace(err)
		}
		cm.Namespace = ns
		cmInterface := k.cli.core.ConfigMaps(ns)
		ocm, err := cmInterface.Get(cfg.Name, metav1.GetOptions{})
		if ocm != nil && err == nil {
			cm.ResourceVersion = ocm.ResourceVersion
			if _, err := cmInterface.Update(cm); err != nil {
				return errors.Trace(err)
			}
		} else {
			if _, err := cmInterface.Create(cm); err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

func (k *kubeImpl) ApplySecrets(ns string, secs map[string]specv1.Secret) error {
	for _, sec := range secs {
		secret := &corev1.Secret{}
		// secret for docker repository authentication
		if isRegistrySecret(sec) {
			var err error
			secret, err = k.generateRegistrySecret(ns, sec.Name, string(sec.Data[RegistryAddress]),
				string(sec.Data[RegistryUsername]), string(sec.Data[RegistryPassword]))
			if err != nil {
				return errors.Trace(err)
			}
		} else {
			if err := copier.Copy(secret, &sec); err != nil {
				return errors.Trace(err)
			}
		}
		secret.Namespace = ns
		secretInterface := k.cli.core.Secrets(ns)
		osec, err := secretInterface.Get(sec.Name, metav1.GetOptions{})
		if osec != nil && err == nil {
			secret.ResourceVersion = osec.ResourceVersion
			_, err := secretInterface.Update(secret)
			if err != nil {
				return errors.Trace(err)
			}
		} else {
			_, err := secretInterface.Create(secret)
			if err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

func (k *kubeImpl) DeleteApplication(ns, name string) error {
	set := labels.Set{AppName: name}
	selector := labels.SelectorFromSet(set)
	deploys, err := k.cli.app.Deployments(ns).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return errors.Trace(err)
	}
	services, err := k.cli.core.Services(ns).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return errors.Trace(err)
	}
	deployInterface := k.cli.app.Deployments(ns)
	for _, d := range deploys.Items {
		if err := deployInterface.Delete(d.Name, &metav1.DeleteOptions{}); err != nil {
			return errors.Trace(err)
		}
	}
	svcInterface := k.cli.core.Services(ns)
	for _, s := range services.Items {
		if err := svcInterface.Delete(s.Name, &metav1.DeleteOptions{}); err != nil {
			return errors.Trace(err)
		}
	}
	k.log.Info("ami delete app", log.Any("name", name))
	return nil
}

func (k *kubeImpl) ApplyApplication(ns string, app specv1.Application, imagePullSecs []string) error {
	var imagePullSecrets []corev1.LocalObjectReference
	secs := make(map[string]struct{})
	for _, sec := range imagePullSecs {
		imagePullSecrets = append(imagePullSecrets,
			corev1.LocalObjectReference{
				Name: sec,
			})
		secs[sec] = struct{}{}
	}
	// remove app's secrets which are image-pull secret actually
	for i, v := range app.Volumes {
		if v.Secret != nil {
			if _, ok := secs[v.Secret.Name]; ok {
				app.Volumes[i].Secret = nil
			}
		}
	}
	services := make(map[string]*corev1.Service)
	deploys := make(map[string]*appv1.Deployment)
	for _, svc := range app.Services {
		if deploy, err := k.prepareDeploy(ns, app, svc, imagePullSecrets); err == nil {
			deploys[deploy.Name] = deploy
		} else {
			return errors.Trace(err)
		}
		if service := k.prepareService(ns, app.Name, &svc); service != nil {
			services[service.Name] = service
		}
	}
	if err := k.applyDeploys(ns, deploys); err != nil {
		return errors.Trace(err)
	}
	if err := k.applyServices(ns, services); err != nil {
		return errors.Trace(err)
	}
	k.log.Info("ami apply apps", log.Any("apps", app))
	return nil
}

func (k *kubeImpl) applyDeploys(ns string, deploys map[string]*appv1.Deployment) error {
	deployInterface := k.cli.app.Deployments(ns)
	for _, d := range deploys {
		deploy, err := deployInterface.Get(d.Name, metav1.GetOptions{})
		if deploy != nil && err == nil {
			d.ResourceVersion = deploy.ResourceVersion
			if _, err = deployInterface.Update(d); err != nil {
				return errors.Trace(err)
			}
		} else {
			if _, err = deployInterface.Create(d); err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

func (k *kubeImpl) applyServices(ns string, svcs map[string]*corev1.Service) error {
	svcInterface := k.cli.core.Services(ns)
	for _, svc := range svcs {
		osvc, err := svcInterface.Get(svc.Name, metav1.GetOptions{})
		if osvc != nil && err == nil {
			svc.ResourceVersion = osvc.ResourceVersion
			if _, err := svcInterface.Update(svc); err != nil {
				return errors.Trace(err)
			}
		} else {
			if _, err := svcInterface.Create(svc); err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

func (k *kubeImpl) prepareDeploy(ns string, app specv1.Application, service specv1.Service,
	imagePullSecrets []corev1.LocalObjectReference) (*appv1.Deployment, error) {
	var c corev1.Container
	var volumes []corev1.Volume
	if err := copier.Copy(&c, &service); err != nil {
		return nil, errors.Trace(err)
	}
	if service.Resources != nil {
		c.Resources.Limits = corev1.ResourceList{}
		for n, value := range service.Resources.Limits {
			quantity, err := resource.ParseQuantity(value)
			if err != nil {
				return nil, errors.Trace(err)
			}
			c.Resources.Limits[corev1.ResourceName(n)] = quantity
		}
	}
	env := corev1.EnvVar{
		Name:  KubeNodeName,
		Value: k.knn,
	}
	c.Env = append(c.Env, env)
	if sc := service.SecurityContext; sc != nil {
		c.SecurityContext = &corev1.SecurityContext{
			Privileged: &sc.Privileged,
		}
	}
	var containers []corev1.Container
	containers = append(containers, c)

	for _, v := range app.Volumes {
		volume := corev1.Volume{
			Name: v.Name,
		}
		if v.Config != nil {
			volume.VolumeSource.ConfigMap = &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: v.VolumeSource.Config.Name},
			}
		} else if v.Secret != nil {
			volume.VolumeSource.Secret = &corev1.SecretVolumeSource{
				SecretName: v.VolumeSource.Secret.Name,
			}
		} else if v.HostPath != nil {
			volume.VolumeSource.HostPath = &corev1.HostPathVolumeSource{
				Path: v.VolumeSource.HostPath.Path,
			}
		}
		volumes = append(volumes, volume)
	}
	replica := new(int32)
	*replica = int32(service.Replica)
	if app.Labels == nil {
		app.Labels = map[string]string{}
	}
	app.Labels[AppName] = app.Name
	app.Labels[AppVersion] = app.Version
	app.Labels[ServiceName] = service.Name
	deploy := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.Name,
			Namespace: ns,
			Labels:    app.Labels,
		},
		Spec: appv1.DeploymentSpec{
			Replicas: replica,
			Strategy: appv1.DeploymentStrategy{
				Type: appv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{ServiceName: service.Name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{ServiceName: service.Name}},
				Spec: corev1.PodSpec{
					Volumes:          volumes,
					Containers:       containers,
					ImagePullSecrets: imagePullSecrets,
					HostNetwork:      service.HostNetwork,
					NodeName:         k.knn,
				},
			},
		},
	}
	if strings.Contains(app.Name, BaetylCore) {
		deploy.Spec.Template.Spec.ServiceAccountName = ServiceAccountName
	}
	return deploy, nil
}

func (k *kubeImpl) prepareService(ns, appName string, svc *specv1.Service) *corev1.Service {
	if len(svc.Ports) == 0 {
		return nil
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
			Namespace: ns,
			Labels:    map[string]string{AppName: appName},
		},
		Spec: corev1.ServiceSpec{
			Selector:  map[string]string{ServiceName: svc.Name},
			ClusterIP: "None",
			Ports:     ports,
		},
	}
	return service
}

func isRegistrySecret(secret specv1.Secret) bool {
	registry, ok := secret.Labels[specv1.SecretLabel]
	return ok && registry == specv1.SecretRegistry
}
