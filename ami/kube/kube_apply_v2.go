package kube

import (
	"fmt"
	"strings"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/jinzhu/copier"
	appv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/baetyl/baetyl/v2/ami"
)

func (k *kubeImpl) applyApplicationV2(ns string, app specv1.Application, imagePullSecs []string) error {
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

	if app.Workload == "" {
		// compatible with the original one service corresponding to one workload
		if len(app.Services) > 0 && app.Services[0].Type != "" {
			app.Workload = app.Services[0].Type
		} else {
			app.Workload = specv1.WorkloadDeployment
		}
	}

	if app.Labels == nil {
		app.Labels = map[string]string{}
	}
	app.Labels[AppName] = app.Name
	app.Labels[AppVersion] = app.Version

	switch app.Workload {
	case specv1.WorkloadDaemonSet:
		if daemon, err := prepareDaemonV2(ns, &app, imagePullSecrets); err != nil {
			return errors.Trace(err)
		} else {
			daemons[daemon.Name] = daemon
		}
	case specv1.WorkloadDeployment:
		if deploy, err := prepareDeployV2(ns, &app, imagePullSecrets); err != nil {
			return errors.Trace(err)
		} else {
			deploys[deploy.Name] = deploy
		}
	case specv1.WorkloadJob:
		if job, err := prepareJobV2(ns, &app, imagePullSecrets); err != nil {
			return errors.Trace(err)
		} else {
			jobs[job.Name] = job
		}
	default:
		k.log.Warn("service type not support", log.Any("type", app.Workload), log.Any("name", app.Name))
	}

	if service := k.prepareServiceV2(ns, app); service != nil {
		services[service.Name] = service
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
	k.log.Info("ami apply apps", log.Any("apps", app))
	return nil
}

func (k *kubeImpl) prepareServiceV2(ns string, app specv1.Application) *corev1.Service {
	var ports []corev1.ServicePort
	for _, svc := range app.Services {
		if len(svc.Ports) == 0 {
			continue
		}
		for i, p := range svc.Ports {
			port := corev1.ServicePort{
				Name:       fmt.Sprintf("%s-%d", svc.Name, i),
				Port:       p.ContainerPort,
				TargetPort: intstr.IntOrString{IntVal: p.ContainerPort},
			}
			ports = append(ports, port)
		}
	}
	if len(ports) == 0 {
		return nil
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
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

func prepareJobV2(ns string, app *specv1.Application, imagePullSecrets []corev1.LocalObjectReference) (*batchv1.Job, error) {
	podSpec, err := prepareInfoV2(app, imagePullSecrets)
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
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: ns,
			Labels:    labels,
		},
		Spec: jobSpec,
	}, nil
}

func prepareDaemonV2(ns string, app *specv1.Application, imagePullSecrets []corev1.LocalObjectReference) (*appv1.DaemonSet, error) {
	podSpec, err := prepareInfoV2(app, imagePullSecrets)
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
	}, nil
}

func prepareDeployV2(ns string, app *specv1.Application, imagePullSecrets []corev1.LocalObjectReference) (*appv1.Deployment, error) {
	podSpec, err := prepareInfoV2(app, imagePullSecrets)
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
	for k, v := range app.Labels {
		deploy.Spec.Template.Labels[k] = v
	}
	if strings.Contains(app.Name, specv1.BaetylCore) || strings.Contains(app.Name, specv1.BaetylInit) {
		deploy.Spec.Template.Spec.ServiceAccountName = ServiceAccountName
	}
	return deploy, nil
}

func prepareInfoV2(app *specv1.Application, imagePullSecrets []corev1.LocalObjectReference) (*corev1.PodSpec, error) {
	var containers []corev1.Container
	var volumes []corev1.Volume

	for _, svc := range app.Services {
		var c corev1.Container
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
			volume.VolumeSource.HostPath = &corev1.HostPathVolumeSource{
				Path: v.VolumeSource.HostPath.Path,
			}
		}
		volumes = append(volumes, volume)
	}

	return &corev1.PodSpec{
		Volumes:          volumes,
		Containers:       containers,
		ImagePullSecrets: imagePullSecrets,
		HostNetwork:      app.HostNetwork,
	}, nil
}
