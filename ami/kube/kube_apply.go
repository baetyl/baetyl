package kube

import (
	"context"
	"fmt"
	"strings"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/jinzhu/copier"
	appv1 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	lb "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/baetyl/baetyl/v2/ami"
)

const (
	KubeNodeName = "KUBE_NODE_NAME"
	AppName      = "baetyl-app-name"
	AppVersion   = "baetyl-app-version"
	ServiceName  = "baetyl-service-name"
	PrefixBaetyl = "baetyl-"

	RegistryAddress  = "address"
	RegistryUsername = "username"
	RegistryPassword = "password"

	ServiceAccountName = "baetyl-edge-system-service-account"
	MasterRole         = "node-role.kubernetes.io/master"
	BaetylSetPodSpec   = "baetyl_set_pod_spec"

	newBackendContainerPort = 54000
)

type SetPodSpecFunc func(*corev1.PodSpec, *specv1.Application) (*corev1.PodSpec, error)

var (
	ErrSetPodSpec = errors.New("failed to convert SetPodSpec function")
)

func (k *kubeImpl) createNamespace(ns string) (*corev1.Namespace, error) {
	defer utils.Trace(k.log.Debug, "applyNamespace")()
	return k.cli.core.Namespaces().Create(context.TODO(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: ns},
	}, metav1.CreateOptions{})
}

func (k *kubeImpl) getNamespace(ns string) (*corev1.Namespace, error) {
	defer utils.Trace(k.log.Debug, "getNamespace")()
	return k.cli.core.Namespaces().Get(context.TODO(), ns, metav1.GetOptions{})
}

func (k *kubeImpl) checkAndCreateNamespace(ns string) error {
	defer utils.Trace(k.log.Debug, "checkAndCreateNamespace")()
	_, err := k.getNamespace(ns)
	if err != nil && strings.Contains(err.Error(), "not found") {
		k.log.Debug("namespace not found, will be created", log.Any("ns", ns))
		_, err = k.createNamespace(ns)
		return errors.Trace(err)
	}
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (k *kubeImpl) applyConfigurations(ns string, cfgs map[string]specv1.Configuration) error {
	for _, cfg := range cfgs {
		cm := &corev1.ConfigMap{}
		if err := copier.Copy(cm, &cfg); err != nil {
			return errors.Trace(err)
		}
		cm.Namespace = ns
		cmInterface := k.cli.core.ConfigMaps(ns)
		ocm, err := cmInterface.Get(context.TODO(), cfg.Name, metav1.GetOptions{})
		if ocm != nil && err == nil {
			cm.ResourceVersion = ocm.ResourceVersion
			if _, err = cmInterface.Update(context.TODO(), cm, metav1.UpdateOptions{}); err != nil {
				return errors.Trace(err)
			}
		} else {
			if _, err = cmInterface.Create(context.TODO(), cm, metav1.CreateOptions{}); err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

func (k *kubeImpl) applySecrets(ns string, secs map[string]specv1.Secret) error {
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
		osec, err := secretInterface.Get(context.TODO(), sec.Name, metav1.GetOptions{})
		if osec != nil && err == nil {
			secret.ResourceVersion = osec.ResourceVersion
			_, err = secretInterface.Update(context.TODO(), secret, metav1.UpdateOptions{})
			if err != nil {
				return errors.Trace(err)
			}
		} else {
			_, err = secretInterface.Create(context.TODO(), secret, metav1.CreateOptions{})
			if err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

func (k *kubeImpl) applyApplication(ns string, app specv1.Application, imagePullSecs []string) error {
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
	daemons := make(map[string]*appv1.DaemonSet)
	jobs := make(map[string]*batchv1.Job)

	if app.Labels == nil {
		app.Labels = map[string]string{}
	}
	app.Labels[AppName] = app.Name
	app.Labels[AppVersion] = app.Version

	k.compatibleDeprecatedField(&app)

	switch app.Workload {
	case specv1.WorkloadDaemonSet:
		if daemon, err := prepareDaemon(ns, &app, imagePullSecrets); err != nil {
			return errors.Trace(err)
		} else {
			daemons[daemon.Name] = daemon
		}
	case specv1.WorkloadDeployment:
		if deploy, err := prepareDeploy(ns, &app, imagePullSecrets); err != nil {
			return errors.Trace(err)
		} else {
			deploys[deploy.Name] = deploy
		}
	case specv1.WorkloadJob:
		if job, err := prepareJob(ns, &app, imagePullSecrets); err != nil {
			return errors.Trace(err)
		} else {
			jobs[job.Name] = job
		}
	default:
		k.log.Warn("service type not support", log.Any("type", app.Workload), log.Any("name", app.Name))
	}

	if service := k.prepareService(ns, app); service != nil {
		services[service.Name] = service
	}
	if nodePortSvc := k.prepareNodePortService(ns, app); nodePortSvc != nil {
		services[nodePortSvc.Name] = nodePortSvc
	}

	if err := k.applyDeploys(ns, deploys); err != nil {
		return errors.Trace(err)
	}
	if err := k.applyDaemons(ns, daemons); err != nil {
		return errors.Trace(err)
	}
	if err := k.applyServices(ns, services); err != nil {
		return errors.Trace(err)
	}
	if err := k.applyJobs(ns, jobs); err != nil {
		return errors.Trace(err)
	}
	if hpa := k.prepareHPA(ns, app); hpa != nil {
		err := k.applyHPA(ns, hpa)
		if err != nil {
			return errors.Trace(err)
		}
		k.log.Info("ami apply hpa", log.Any("hpa", hpa))
	}
	k.log.Info("ami apply apps", log.Any("apps", app))
	return nil
}

func (k *kubeImpl) prepareService(ns string, app specv1.Application) *corev1.Service {
	var ports []corev1.ServicePort
	for _, svc := range app.Services {
		if len(svc.Ports) == 0 {
			continue
		}
		for i, p := range svc.Ports {
			if p.ServiceType == string(corev1.ServiceTypeNodePort) {
				continue
			}
			if p.Protocol == "" {
				p.Protocol = string(corev1.ProtocolTCP)
			}
			port := corev1.ServicePort{
				Name:       fmt.Sprintf("%s-%d-%s", svc.Name, p.ContainerPort, strings.ToLower(p.Protocol)),
				Port:       p.ContainerPort,
				TargetPort: intstr.IntOrString{IntVal: p.ContainerPort},
				Protocol:   corev1.Protocol(p.Protocol),
			}
			if v, ok := app.Labels["baetyl-webhook"]; ok && v == "true" {
				port.TargetPort = intstr.IntOrString{IntVal: int32(newBackendContainerPort + i)}
			}
			ports = append(ports, port)
		}
	}
	if len(ports) == 0 {
		return nil
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cutSysServiceRandSuffix(app.Name),
			Namespace: ns,
			Labels:    map[string]string{AppName: app.Name},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{AppName: app.Name},
			Ports:    ports,
		},
	}
	return service
}

func (k *kubeImpl) prepareNodePortService(ns string, app specv1.Application) *corev1.Service {
	var ports []corev1.ServicePort
	for _, svc := range app.Services {
		if len(svc.Ports) == 0 {
			continue
		}
		for i, p := range svc.Ports {
			if p.ServiceType != string(corev1.ServiceTypeNodePort) {
				continue
			}
			if p.Protocol == "" {
				p.Protocol = string(corev1.ProtocolTCP)
			}
			port := corev1.ServicePort{
				Name:       fmt.Sprintf("%s-%d-%s", svc.Name, p.ContainerPort, strings.ToLower(p.Protocol)),
				Port:       p.ContainerPort,
				TargetPort: intstr.IntOrString{IntVal: p.ContainerPort},
				NodePort:   p.NodePort,
				Protocol:   corev1.Protocol(p.Protocol),
			}
			if v, ok := app.Labels["baetyl-webhook"]; ok && v == "true" {
				port.TargetPort = intstr.IntOrString{IntVal: int32(newBackendContainerPort + i)}
			}
			ports = append(ports, port)
		}
	}
	if len(ports) == 0 {
		return nil
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", cutSysServiceRandSuffix(app.Name), "nodeport"),
			Namespace: ns,
			Labels:    map[string]string{AppName: app.Name},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{AppName: app.Name},
			Ports:    ports,
			Type:     corev1.ServiceTypeNodePort,
		},
	}
	return service
}

func (k *kubeImpl) prepareHPA(ns string, app specv1.Application) *v2.HorizontalPodAutoscaler {
	if !k.hpaAvailable() || app.AutoScaleCfg == nil || len(app.AutoScaleCfg.Metrics) == 0 {
		return nil
	}

	var metrics []v2.MetricSpec
	for _, m := range app.AutoScaleCfg.Metrics {
		metric := v2.MetricSpec{
			Type: v2.MetricSourceType(m.Type),
			Resource: &v2.ResourceMetricSource{
				Name: corev1.ResourceName(m.Resource.Name),
				Target: v2.MetricTarget{
					Type: v2.MetricTargetType(m.Resource.TargetType),
				},
			},
		}
		if m.Resource.AverageUtilization != 0 {
			averageUtilization := int32(m.Resource.AverageUtilization)
			metric.Resource.Target.AverageUtilization = &averageUtilization
		}
		if m.Resource.AverageValue != "" {
			averageValue, err := resource.ParseQuantity(m.Resource.AverageValue)
			if err != nil {
				k.log.Error("failed to parse quantity", log.Error(err))
				return nil
			}
			metric.Resource.Target.AverageValue = &averageValue
		}
		if m.Resource.Value != "" {
			value, err := resource.ParseQuantity(m.Resource.Value)
			if err != nil {
				k.log.Error("failed to parse quantity", log.Error(err))
				return nil
			}
			metric.Resource.Target.Value = &value
		}

		metrics = append(metrics, metric)
	}

	minReplica := int32(app.AutoScaleCfg.MinReplicas)
	hpa := &v2.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: ns,
			Labels:    map[string]string{AppName: app.Name},
		},
		Spec: v2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: v2.CrossVersionObjectReference{
				Kind:       "Deployment",
				Name:       app.Name,
				APIVersion: "apps/v1",
			},
			MinReplicas: &minReplica,
			MaxReplicas: int32(app.AutoScaleCfg.MaxReplicas),
			Metrics:     metrics,
		},
	}
	return hpa
}

func prepareJob(ns string, app *specv1.Application, imagePullSecrets []corev1.LocalObjectReference) (*batchv1.Job, error) {
	podSpec, err := prepareInfo(app, imagePullSecrets)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if extension, ok := ami.Hooks[BaetylSetPodSpec]; ok {
		setPodSpecExt, ok := extension.(SetPodSpecFunc)
		if ok {
			if podSpec, err = setPodSpecExt(podSpec, app); err != nil {
				return nil, errors.Trace(err)
			}
		} else {
			return nil, errors.Trace(ErrSetPodSpec)
		}
	} else {
		return nil, errors.Trace(ErrSetPodSpec)
	}

	jobSpec := batchv1.JobSpec{}
	podSpec.RestartPolicy = corev1.RestartPolicyNever
	if app.JobConfig != nil {
		parallelism := int32(app.JobConfig.Parallelism)
		completions := int32(app.JobConfig.Completions)
		backoffLimit := int32(app.JobConfig.BackoffLimit)
		podSpec.RestartPolicy = corev1.RestartPolicy(app.JobConfig.RestartPolicy)
		jobSpec.Parallelism = &parallelism
		jobSpec.Completions = &completions
		jobSpec.BackoffLimit = &backoffLimit
	}
	jobSpec.Template.Spec = *podSpec
	labels := map[string]string{}
	for k, v := range app.Labels {
		labels[k] = v
	}
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: jobSpec,
	}
	job.Spec.Template.Labels = labels
	return job, nil
}

func prepareDaemon(ns string, app *specv1.Application, imagePullSecrets []corev1.LocalObjectReference) (*appv1.DaemonSet, error) {
	podSpec, err := prepareInfo(app, imagePullSecrets)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if extension, ok := ami.Hooks[BaetylSetPodSpec]; ok {
		setPodSpecExt, ok := extension.(SetPodSpecFunc)
		if ok {
			if podSpec, err = setPodSpecExt(podSpec, app); err != nil {
				return nil, errors.Trace(err)
			}
		} else {
			return nil, errors.Trace(ErrSetPodSpec)
		}
	} else {
		return nil, errors.Trace(ErrSetPodSpec)
	}

	labels := map[string]string{}
	for k, v := range app.Labels {
		labels[k] = v
	}
	ds := &appv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: appv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{AppName: app.Name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{AppName: app.Name}},
				Spec:       *podSpec,
			},
		},
	}
	if app.Labels != nil {
		for k, v := range app.Labels {
			ds.Spec.Template.Labels[k] = v
		}
	}
	return ds, nil
}

func prepareDeploy(ns string, app *specv1.Application, imagePullSecrets []corev1.LocalObjectReference) (*appv1.Deployment, error) {
	podSpec, err := prepareInfo(app, imagePullSecrets)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if extension, ok := ami.Hooks[BaetylSetPodSpec]; ok {
		setPodSpecExt, ok := extension.(SetPodSpecFunc)
		if ok {
			if podSpec, err = setPodSpecExt(podSpec, app); err != nil {
				return nil, errors.Trace(err)
			}
		} else {
			return nil, errors.Trace(ErrSetPodSpec)
		}
	} else {
		return nil, errors.Trace(ErrSetPodSpec)
	}
	replica := new(int32)
	*replica = int32(app.Replica)

	labels := map[string]string{}
	for k, v := range app.Labels {
		labels[k] = v
	}
	deploy := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: appv1.DeploymentSpec{
			Replicas: replica,
			Strategy: appv1.DeploymentStrategy{
				Type: appv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{AppName: app.Name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{AppName: app.Name}},
				Spec:       *podSpec,
			},
		},
	}
	if app.Labels != nil {
		for k, v := range app.Labels {
			deploy.Spec.Template.Labels[k] = v
		}
	}
	if strings.Contains(app.Name, specv1.BaetylCore) || strings.Contains(app.Name, specv1.BaetylInit) {
		deploy.Spec.Template.Spec.ServiceAccountName = ServiceAccountName
	}
	return deploy, nil
}

func prepareInfo(app *specv1.Application, imagePullSecrets []corev1.LocalObjectReference) (*corev1.PodSpec, error) {
	var containers []corev1.Container
	var initContainers []corev1.Container
	var volumes []corev1.Volume

	for _, initSvc := range app.InitServices {
		var c corev1.Container
		_, err := TransSvcToContainer(&initSvc, &c)
		if err != nil {
			return nil, errors.Trace(err)
		}
		initContainers = append(initContainers, c)
	}

	for _, svc := range app.Services {
		var c corev1.Container
		_, err := TransSvcToContainer(&svc, &c)
		if err != nil {
			return nil, errors.Trace(err)
		}
		containers = append(containers, c)
	}
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
			tp := corev1.HostPathType(v.VolumeSource.HostPath.Type)
			volume.VolumeSource.HostPath = &corev1.HostPathVolumeSource{
				Path: v.VolumeSource.HostPath.Path,
				Type: &tp,
			}
		} else if v.VolumeSource.EmptyDir != nil {
			volume.VolumeSource.EmptyDir = &corev1.EmptyDirVolumeSource{
				Medium: corev1.StorageMedium(v.VolumeSource.EmptyDir.Medium),
			}
			if len(v.VolumeSource.EmptyDir.SizeLimit) > 0 {
				quantity, err := resource.ParseQuantity(v.VolumeSource.EmptyDir.SizeLimit)
				if err != nil {
					return nil, errors.Trace(err)
				}
				volume.VolumeSource.EmptyDir.SizeLimit = &quantity
			}
		}
		volumes = append(volumes, volume)
	}

	return &corev1.PodSpec{
		Volumes:          volumes,
		InitContainers:   initContainers,
		Containers:       containers,
		ImagePullSecrets: imagePullSecrets,
		HostNetwork:      app.HostNetwork,
		DNSPolicy:        app.DNSPolicy,
	}, nil
}

func TransSvcToContainer(svc *specv1.Service, c *corev1.Container) (*corev1.Container, error) {
	if err := copier.Copy(&c, &svc); err != nil {
		return nil, errors.Trace(err)
	}
	if svc.Resources != nil {
		c.Resources.Limits = corev1.ResourceList{}
		for n, value := range svc.Resources.Limits {
			quantity, err := resource.ParseQuantity(value)
			if err != nil {
				return nil, errors.Trace(err)
			}
			c.Resources.Limits[corev1.ResourceName(n)] = quantity
		}
	}
	if sc := svc.SecurityContext; sc != nil {
		c.SecurityContext = &corev1.SecurityContext{
			Privileged: &sc.Privileged,
		}
	}
	c.Env = append(c.Env, corev1.EnvVar{
		Name:      KubeNodeName,
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
	})
	return c, nil
}

func (k *kubeImpl) deleteApplication(ns, name string) error {
	set := lb.Set{AppName: name}
	selector := lb.SelectorFromSet(set)
	deploys, err := k.cli.app.Deployments(ns).List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return errors.Trace(err)
	}
	daemons, err := k.cli.app.DaemonSets(ns).List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return errors.Trace(err)
	}
	services, err := k.cli.core.Services(ns).List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return errors.Trace(err)
	}
	jobs, err := k.cli.batch.Jobs(ns).List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return errors.Trace(err)
	}
	if k.hpaAvailable() {
		hpas, err := k.cli.autoscale.HorizontalPodAutoscalers(ns).List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return errors.Trace(err)
		}
		as := k.cli.autoscale.HorizontalPodAutoscalers(ns)
		for _, hpa := range hpas.Items {
			if err = as.Delete(context.TODO(), hpa.Name, metav1.DeleteOptions{}); err != nil {
				return errors.Trace(err)
			}
		}
	}

	deployInterface := k.cli.app.Deployments(ns)
	for _, d := range deploys.Items {
		if err = deployInterface.Delete(context.TODO(), d.Name, metav1.DeleteOptions{}); err != nil {
			return errors.Trace(err)
		}
	}
	daemonInterface := k.cli.app.DaemonSets(ns)
	for _, d := range daemons.Items {
		if err = daemonInterface.Delete(context.TODO(), d.Name, metav1.DeleteOptions{}); err != nil {
			return errors.Trace(err)
		}
	}
	svcInterface := k.cli.core.Services(ns)
	for _, s := range services.Items {
		if err = svcInterface.Delete(context.TODO(), s.Name, metav1.DeleteOptions{}); err != nil {
			return errors.Trace(err)
		}
	}
	jobInterface := k.cli.batch.Jobs(ns)
	policy := metav1.DeletePropagationBackground
	for _, j := range jobs.Items {
		if err = jobInterface.Delete(context.TODO(), j.Name, metav1.DeleteOptions{PropagationPolicy: &policy}); err != nil {
			return errors.Trace(err)
		}
	}

	k.log.Info("ami delete app", log.Any("name", name))
	return nil
}

func (k *kubeImpl) applyDeploys(ns string, deploys map[string]*appv1.Deployment) error {
	deployInterface := k.cli.app.Deployments(ns)
	for _, d := range deploys {
		deploy, err := deployInterface.Get(context.TODO(), d.Name, metav1.GetOptions{})
		if deploy != nil && err == nil {
			d.ResourceVersion = deploy.ResourceVersion
			if _, err = deployInterface.Update(context.TODO(), d, metav1.UpdateOptions{}); err != nil {
				return errors.Trace(err)
			}
		} else {
			if _, err = deployInterface.Create(context.TODO(), d, metav1.CreateOptions{}); err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

func (k *kubeImpl) applyDaemons(ns string, daemons map[string]*appv1.DaemonSet) error {
	daemonInterface := k.cli.app.DaemonSets(ns)
	for _, d := range daemons {
		daemon, err := daemonInterface.Get(context.TODO(), d.Name, metav1.GetOptions{})
		if daemon != nil && err == nil {
			d.ResourceVersion = daemon.ResourceVersion
			if _, err = daemonInterface.Update(context.TODO(), d, metav1.UpdateOptions{}); err != nil {
				return errors.Trace(err)
			}
		} else {
			if _, err = daemonInterface.Create(context.TODO(), d, metav1.CreateOptions{}); err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

func (k *kubeImpl) applyJobs(ns string, jobs map[string]*batchv1.Job) error {
	jobInterface := k.cli.batch.Jobs(ns)
	for _, j := range jobs {
		job, err := jobInterface.Get(context.TODO(), j.Name, metav1.GetOptions{})
		if job != nil && err == nil {
			policy := metav1.DeletePropagationBackground
			err = jobInterface.Delete(context.TODO(), job.Name, metav1.DeleteOptions{PropagationPolicy: &policy})
			if err != nil {
				return errors.Trace(err)
			}
		}
		if _, err = jobInterface.Create(context.TODO(), j, metav1.CreateOptions{}); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func (k *kubeImpl) applyServices(ns string, svcs map[string]*corev1.Service) error {
	svcInterface := k.cli.core.Services(ns)
	for _, svc := range svcs {
		osvc, err := svcInterface.Get(context.TODO(), svc.Name, metav1.GetOptions{})
		if osvc != nil && err == nil {
			svc.ResourceVersion = osvc.ResourceVersion
			svc.Spec.ClusterIP = osvc.Spec.ClusterIP
			if _, err = svcInterface.Update(context.TODO(), svc, metav1.UpdateOptions{}); err != nil {
				return errors.Trace(err)
			}
		} else {
			if _, err = svcInterface.Create(context.TODO(), svc, metav1.CreateOptions{}); err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

func (k *kubeImpl) applyHPA(ns string, hpa *v2.HorizontalPodAutoscaler) error {
	as := k.cli.autoscale.HorizontalPodAutoscalers(ns)

	h, err := as.Get(context.TODO(), hpa.Name, metav1.GetOptions{})
	if h != nil && err == nil {
		if _, err = as.Update(context.TODO(), hpa, metav1.UpdateOptions{}); err != nil {
			return errors.Trace(err)
		}
	} else {
		if _, err = as.Create(context.TODO(), hpa, metav1.CreateOptions{}); err != nil {
			return errors.Trace(err)
		}
	}

	return nil
}

func (k *kubeImpl) hpaAvailable() bool {
	info, err := k.cli.discovery.ServerVersion()
	if err != nil {
		return false
	}
	return info.Major == "1" && strings.Compare(info.Minor, "23") >= 0
}

func SetPodSpec(spec *corev1.PodSpec, _ *specv1.Application) (*corev1.PodSpec, error) {
	spec.Affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      MasterRole,
						Operator: corev1.NodeSelectorOpExists,
					},
				}}},
			},
		},
	}
	return spec, nil
}

func isRegistrySecret(secret specv1.Secret) bool {
	registry, ok := secret.Labels[specv1.SecretLabel]
	return ok && registry == specv1.SecretRegistry
}

func cutSysServiceRandSuffix(s string) string {
	if strings.HasPrefix(s, PrefixBaetyl) {
		sub := s[len(PrefixBaetyl):]
		if idx := strings.LastIndex(sub, "-"); idx != -1 {
			return PrefixBaetyl + sub[:idx]
		}
	}
	return s
}

func (k *kubeImpl) compatibleDeprecatedField(app *specv1.Application) {
	// Workload
	if app.Workload == "" {
		// compatible with the original one service corresponding to one workload
		if len(app.Services) > 0 && app.Services[0].Type != "" {
			k.log.Debug("workload is empty, use the services[0].Type ")
			app.Workload = app.Services[0].Type
		} else {
			app.Workload = specv1.WorkloadDeployment
		}
	}

	// HostNetwork
	if !app.HostNetwork && len(app.Services) > 0 && app.Services[0].HostNetwork {
		k.log.Debug("app.HostNetwork is false, use the services[0].HostNetwork true ")
		app.HostNetwork = true
	}

	// Replica
	if app.Replica == 0 {
		// compatible with the original one service corresponding to one workload
		if len(app.Services) > 0 && app.Services[0].Replica != 0 {
			k.log.Debug("app.Replica is 0, use the services[0].Replica", log.Any("replica", app.Services[0].Replica))
			app.Replica = app.Services[0].Replica
		}
	}

	// JobConfig
	if app.JobConfig == nil || app.JobConfig.RestartPolicy == "" {
		// compatible with the original one service corresponding to one workload
		if len(app.Services) > 0 && app.Services[0].JobConfig != nil {
			k.log.Debug("app.JobConfig is 0, use the services[0].JobConfig ")
			app.JobConfig = &specv1.AppJobConfig{
				Completions:   app.Services[0].JobConfig.Completions,
				Parallelism:   app.Services[0].JobConfig.Parallelism,
				BackoffLimit:  app.Services[0].JobConfig.BackoffLimit,
				RestartPolicy: app.Services[0].JobConfig.RestartPolicy,
			}
		} else {
			app.JobConfig = &specv1.AppJobConfig{RestartPolicy: "Never"}
		}
	}
}
