package kube

import (
	"fmt"
	"strings"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/jinzhu/copier"
	appv1 "k8s.io/api/apps/v1"
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

	RegistryAddress  = "address"
	RegistryUsername = "username"
	RegistryPassword = "password"

	ServiceAccountName = "baetyl-edge-system-service-account"
	MasterRole         = "node-role.kubernetes.io/master"
	BaetylSetPodSpec   = "baetyl_set_pod_spec"
)

type SetPodSpecFunc func(*corev1.PodSpec, *specv1.Application) (*corev1.PodSpec, error)

var (
	ErrSetPodSpec = errors.New("failed to convert SetPodSpec function")
)

func (k *kubeImpl) createNamespace(ns string) (*corev1.Namespace, error) {
	defer utils.Trace(k.log.Debug, "applyNamespace")()
	return k.cli.core.Namespaces().Create(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: ns},
	})
}

func (k *kubeImpl) getNamespace(ns string) (*corev1.Namespace, error) {
	defer utils.Trace(k.log.Debug, "getNamespace")()
	return k.cli.core.Namespaces().Get(ns, metav1.GetOptions{})
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

func (k *kubeImpl) deleteApplication(ns, name string) error {
	set := lb.Set{AppName: name}
	selector := lb.SelectorFromSet(set)
	deploys, err := k.cli.app.Deployments(ns).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return errors.Trace(err)
	}
	daemons, err := k.cli.app.DaemonSets(ns).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return errors.Trace(err)
	}
	services, err := k.cli.core.Services(ns).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return errors.Trace(err)
	}
	jobs, err := k.cli.batch.Jobs(ns).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return errors.Trace(err)
	}
	deployInterface := k.cli.app.Deployments(ns)
	for _, d := range deploys.Items {
		if err := deployInterface.Delete(d.Name, &metav1.DeleteOptions{}); err != nil {
			return errors.Trace(err)
		}
	}
	daemonInterface := k.cli.app.DaemonSets(ns)
	for _, d := range daemons.Items {
		if err := daemonInterface.Delete(d.Name, &metav1.DeleteOptions{}); err != nil {
			return errors.Trace(err)
		}
	}
	svcInterface := k.cli.core.Services(ns)
	for _, s := range services.Items {
		if err := svcInterface.Delete(s.Name, &metav1.DeleteOptions{}); err != nil {
			return errors.Trace(err)
		}
	}
	jobInterface := k.cli.batch.Jobs(ns)
	policy := metav1.DeletePropagationBackground
	for _, j := range jobs.Items {
		if err = jobInterface.Delete(j.Name, &metav1.DeleteOptions{PropagationPolicy: &policy}); err != nil {
			return errors.Trace(err)
		}
	}
	k.log.Info("ami delete app", log.Any("name", name))
	return nil
}

// applyApplication Compatible with the previous logic, the service in each app is used as a separate deployment
// Deprecated: Use applyApplicationV2 instead
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
	for _, svc := range app.Services {
		if svc.Type == "" {
			svc.Type = specv1.WorkloadDeployment
		}
		switch svc.Type {
		case specv1.WorkloadDaemonSet:
			if daemon, err := prepareDaemon(ns, &app, svc, imagePullSecrets); err != nil {
				return errors.Trace(err)
			} else {
				daemons[daemon.Name] = daemon
			}
		case specv1.WorkloadDeployment:
			if deploy, err := prepareDeploy(ns, &app, svc, imagePullSecrets); err != nil {
				return errors.Trace(err)
			} else {
				deploys[deploy.Name] = deploy
			}
		case specv1.WorkloadJob:
			if job, err := prepareJob(ns, &app, svc, imagePullSecrets); err != nil {
				return errors.Trace(err)
			} else {
				jobs[job.Name] = job
			}
		default:
			k.log.Warn("service type not support", log.Any("type", svc.Type), log.Any("name", svc.Name))
		}

		if service := k.prepareService(ns, app.Name, &svc); service != nil {
			services[service.Name] = service
		}
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
	k.log.Info("ami apply apps by service", log.Any("apps", app))
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

func (k *kubeImpl) applyDaemons(ns string, daemons map[string]*appv1.DaemonSet) error {
	daemonInterface := k.cli.app.DaemonSets(ns)
	for _, d := range daemons {
		daemon, err := daemonInterface.Get(d.Name, metav1.GetOptions{})
		if daemon != nil && err == nil {
			d.ResourceVersion = daemon.ResourceVersion
			if _, err = daemonInterface.Update(d); err != nil {
				return errors.Trace(err)
			}
		} else {
			if _, err = daemonInterface.Create(d); err != nil {
				return errors.Trace(err)
			}
		}
	}
	return nil
}

func (k *kubeImpl) applyJobs(ns string, jobs map[string]*batchv1.Job) error {
	jobInterface := k.cli.batch.Jobs(ns)
	for _, j := range jobs {
		job, err := jobInterface.Get(j.Name, metav1.GetOptions{})
		if job != nil && err == nil {
			policy := metav1.DeletePropagationBackground
			err = jobInterface.Delete(job.Name, &metav1.DeleteOptions{PropagationPolicy: &policy})
			if err != nil {
				return errors.Trace(err)
			}
		}
		if _, err = jobInterface.Create(j); err != nil {
			return errors.Trace(err)
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
			svc.Spec.ClusterIP = osvc.Spec.ClusterIP
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

// Deprecated: Use prepareDeployV2 instead.
// Change from one workload for each service to one workload for one app, and each service as a container
func prepareDeploy(ns string, app *specv1.Application, service specv1.Service,
	imagePullSecrets []corev1.LocalObjectReference) (*appv1.Deployment, error) {
	podSpec, err := prepareInfo(app, service, imagePullSecrets)
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
	*replica = int32(service.Replica)

	labels := map[string]string{}
	for k, v := range app.Labels {
		labels[k] = v
	}
	deploy := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.Name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: appv1.DeploymentSpec{
			Replicas: replica,
			Strategy: appv1.DeploymentStrategy{
				Type: appv1.RecreateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{ServiceName: service.Name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{ServiceName: service.Name}},
				Spec:       *podSpec,
			},
		},
	}
	for k, v := range service.Labels {
		deploy.Spec.Template.Labels[k] = v
	}
	if strings.Contains(app.Name, specv1.BaetylCore) || strings.Contains(app.Name, specv1.BaetylInit) {
		deploy.Spec.Template.Spec.ServiceAccountName = ServiceAccountName
	}
	return deploy, nil
}

// Deprecated: Use prepareDaemonV2 instead.
// Change from one workload for each service to one workload for one app, and each service as a container
func prepareDaemon(ns string, app *specv1.Application, service specv1.Service,
	imagePullSecrets []corev1.LocalObjectReference) (*appv1.DaemonSet, error) {
	podSpec, err := prepareInfo(app, service, imagePullSecrets)
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
	return &appv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.Name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: appv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{ServiceName: service.Name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{ServiceName: service.Name}},
				Spec:       *podSpec,
			},
		},
	}, nil
}

// Deprecated: Use prepareJobV2 instead.
// Change from one workload for each service to one workload for one app, and each service as a container
func prepareJob(ns string, app *specv1.Application, service specv1.Service,
	imagePullSecrets []corev1.LocalObjectReference) (*batchv1.Job, error) {
	podSpec, err := prepareInfo(app, service, imagePullSecrets)
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
	if service.JobConfig != nil {
		parallelism := int32(service.JobConfig.Parallelism)
		completions := int32(service.JobConfig.Completions)
		backoffLimit := int32(service.JobConfig.BackoffLimit)
		podSpec.RestartPolicy = corev1.RestartPolicy(service.JobConfig.RestartPolicy)
		jobSpec.Parallelism = &parallelism
		jobSpec.Completions = &completions
		jobSpec.BackoffLimit = &backoffLimit
	}
	jobSpec.Template.Spec = *podSpec
	labels := map[string]string{}
	for k, v := range app.Labels {
		labels[k] = v
	}
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      service.Name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: jobSpec,
	}, nil
}

// Deprecated: Use prepareInfoV2 instead.
// Change from one workload for each service to one workload for one app, and each service as a container
func prepareInfo(app *specv1.Application, service specv1.Service,
	imagePullSecrets []corev1.LocalObjectReference) (*corev1.PodSpec, error) {
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
	if sc := service.SecurityContext; sc != nil {
		c.SecurityContext = &corev1.SecurityContext{
			Privileged: &sc.Privileged,
		}
	}
	c.Env = append(c.Env, corev1.EnvVar{
		Name:      KubeNodeName,
		ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
	})
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
	if app.Labels == nil {
		app.Labels = map[string]string{}
	}
	app.Labels[AppName] = app.Name
	app.Labels[AppVersion] = app.Version
	app.Labels[ServiceName] = service.Name

	return &corev1.PodSpec{
		Volumes:          volumes,
		Containers:       containers,
		ImagePullSecrets: imagePullSecrets,
		HostNetwork:      service.HostNetwork,
	}, nil
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

// Deprecated: Use prepareServiceV2 instead.
// Change from one workload for each service to one workload for one app, and each service as a container
func (k *kubeImpl) prepareService(ns, appName string, svc *specv1.Service) *corev1.Service {
	if len(svc.Ports) == 0 {
		return nil
	}
	var ports []corev1.ServicePort
	for i, p := range svc.Ports {
		port := corev1.ServicePort{
			Name:       fmt.Sprintf("%s-%d", svc.Name, i),
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
			Selector: map[string]string{ServiceName: svc.Name},
			Ports:    ports,
		},
	}
	return service
}

func isRegistrySecret(secret specv1.Secret) bool {
	registry, ok := secret.Labels[specv1.SecretLabel]
	return ok && registry == specv1.SecretRegistry
}
