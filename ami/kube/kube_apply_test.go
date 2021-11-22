package kube

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/stretchr/testify/assert"
	appv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/store"
)

func TestCreateNamespace(t *testing.T) {
	am := initApplyKubeAMI(t)
	ns := "ns"
	res, err := am.createNamespace(ns)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, ns, res.Name)
}

func TestGetNamespace(t *testing.T) {
	am := initApplyKubeAMI(t)
	ns := "baetyl-edge"
	res, err := am.getNamespace(ns)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, ns, res.Name)
}

func TestCheckAndCreateNamespace(t *testing.T) {
	am := initApplyKubeAMI(t)
	ns := "ns"
	err := am.checkAndCreateNamespace(ns)
	assert.NoError(t, err)
}

func TestApplyDeploys(t *testing.T) {
	ami := initApplyKubeAMI(t)
	ns := "baetyl-edge"
	lables := map[string]string{
		"baetyl": "baetyl",
	}
	ds := map[string]*appv1.Deployment{
		"d1": {
			ObjectMeta: metav1.ObjectMeta{Name: "d1", Namespace: ns, Labels: lables},
		},
		"d2": {
			ObjectMeta: metav1.ObjectMeta{Name: "d2", Namespace: ns, Labels: lables},
		},
	}
	err := ami.applyDeploys(ns, ds)
	assert.NoError(t, err)

	wrongDs := map[string]*appv1.Deployment{
		"d1": {
			ObjectMeta: metav1.ObjectMeta{Name: "d1", Namespace: "default", Labels: lables},
		},
		"d3": {
			ObjectMeta: metav1.ObjectMeta{Name: "d3", Namespace: "default", Labels: lables},
		},
	}
	err = ami.applyDeploys(ns, wrongDs)
	assert.Error(t, err)
}

func TestApplyDaemons(t *testing.T) {
	am := initApplyKubeAMI(t)
	ns := "baetyl-edge"
	lables := map[string]string{
		"baetyl": "baetyl",
	}
	ds := map[string]*appv1.DaemonSet{
		"d1": {
			ObjectMeta: metav1.ObjectMeta{Name: "d1", Namespace: ns, Labels: lables},
		},
		"d2": {
			ObjectMeta: metav1.ObjectMeta{Name: "d2", Namespace: ns, Labels: lables},
		},
	}
	err := am.applyDaemons(ns, ds)
	assert.NoError(t, err)

	wrongDs := map[string]*appv1.DaemonSet{
		"d1": {
			ObjectMeta: metav1.ObjectMeta{Name: "d1", Namespace: "default", Labels: lables},
		},
		"d3": {
			ObjectMeta: metav1.ObjectMeta{Name: "d3", Namespace: "default", Labels: lables},
		},
	}
	err = am.applyDaemons(ns, wrongDs)
	assert.Error(t, err)
}

func TestDeleteApplication(t *testing.T) {
	ami := initApplyKubeAMI(t)
	ns := "baetyl-edge"
	name := "svc1"
	app := specv1.Application{
		Name:      name,
		Namespace: ns,
		Version:   "a1",
		Services: []specv1.Service{{
			Name:  "s1",
			Image: "image1",
			Ports: []specv1.ContainerPort{{
				HostPort:      80,
				ContainerPort: 80,
			}},
		}},
	}
	err := ami.applyApplication(ns, app, nil)
	assert.NoError(t, err)
	d, err := ami.cli.app.Deployments(ns).Get(name, metav1.GetOptions{})
	assert.NotNil(t, d)
	assert.NoError(t, err)
	s, err := ami.cli.core.Services(ns).Get(name, metav1.GetOptions{})
	assert.NotNil(t, s)
	assert.NoError(t, err)

	err = ami.deleteApplication(ns, app.Name)
	assert.NoError(t, err)
	d, _ = ami.cli.app.Deployments(ns).Get(name, metav1.GetOptions{})
	assert.Nil(t, d)
	s, _ = ami.cli.core.Services(ns).Get(name, metav1.GetOptions{})
	assert.Nil(t, s)
}

func TestApplySecret(t *testing.T) {
	ami := initApplyKubeAMI(t)
	ns := "baetyl-edge"
	secs := map[string]specv1.Secret{
		"sec1": {
			Name: "sec1", Namespace: ns,
		},
		"sec2": {
			Name: "sec2", Namespace: ns,
		},
	}
	err := ami.applySecrets(ns, secs)
	assert.NoError(t, err)

	sec := specv1.Secret{
		Name:      "sec",
		Namespace: ns,
	}
	secKey := "sec-key"
	secVal := "sec-val"
	sec.Data = map[string][]byte{
		secKey: []byte(secVal),
	}
	secs = map[string]specv1.Secret{"sec": sec}
	err = ami.applySecrets(ns, secs)
	assert.NoError(t, err)
	expected := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "sec", Namespace: "baetyl-edge"},
	}
	expected.Data = map[string][]byte{
		secKey: []byte(secVal),
	}
	res, err := ami.cli.core.Secrets(ns).Get("sec", metav1.GetOptions{})
	assert.Equal(t, res, expected)

	// registry
	reg := specv1.Secret{Name: "registry"}
	reg.Labels = map[string]string{
		specv1.SecretLabel: specv1.SecretRegistry,
	}
	reg.Data = map[string][]byte{
		RegistryAddress:  []byte("server"),
		RegistryUsername: []byte("test"),
		RegistryPassword: []byte("1234"),
	}
	regs := map[string]specv1.Secret{"registry": reg}
	err = ami.applySecrets(ns, regs)
	assert.NoError(t, err)
	expected = &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "registry", Namespace: ns},
		Type:       v1.SecretTypeDockerConfigJson,
	}
	auths := map[string]interface{}{
		"auths": map[string]auth{
			"server": {
				Username: "test",
				Password: "1234",
				Auth:     base64.StdEncoding.EncodeToString([]byte("test:1234")),
			},
		},
	}
	data, _ := json.Marshal(auths)
	expected.Data = map[string][]byte{
		v1.DockerConfigJsonKey: data,
	}
	res, err = ami.cli.core.Secrets(ns).Get("registry", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.Equal(t, res, expected)
}

func TestApplyConfigMap(t *testing.T) {
	ami := initApplyKubeAMI(t)
	ns := "baetyl-edge"
	cfgs := map[string]specv1.Configuration{
		"cfg1": {
			Name: "cfg1", Namespace: ns,
		},
		"cfg2": {
			Name: "cfg2", Namespace: ns,
		},
	}
	err := ami.applyConfigurations(ns, cfgs)
	assert.NoError(t, err)
}

func TestApplyService(t *testing.T) {
	ami := initApplyKubeAMI(t)
	ns := "baetyl-edge"
	svcs := map[string]*v1.Service{
		"svc1": {
			ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: ns},
		},
		"svc2": {
			ObjectMeta: metav1.ObjectMeta{Name: "svc2", Namespace: ns},
		},
	}
	err := ami.applyServices(ns, svcs)
	assert.NoError(t, err)
	wrongSvcs := map[string]*v1.Service{
		"svc1": {
			ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: "default"},
		},
		"svc3": {
			ObjectMeta: metav1.ObjectMeta{Name: "svc3", Namespace: "default"},
		},
	}
	err = ami.applyServices(ns, wrongSvcs)
	assert.Error(t, err)
}

func TestApplyJobs(t *testing.T) {
	am := initApplyKubeAMI(t)
	ns := "baetyl-edge"
	lables := map[string]string{
		"baetyl": "baetyl",
	}
	js := map[string]*batchv1.Job{
		"d1": {
			ObjectMeta: metav1.ObjectMeta{Name: "j1", Namespace: ns, Labels: lables},
		},
		"d2": {
			ObjectMeta: metav1.ObjectMeta{Name: "j2", Namespace: ns, Labels: lables},
		},
	}
	err := am.applyJobs(ns, js)
	assert.NoError(t, err)

	wrongJs := map[string]*batchv1.Job{
		"d1": {
			ObjectMeta: metav1.ObjectMeta{Name: "j1", Namespace: "default", Labels: lables},
		},
		"d3": {
			ObjectMeta: metav1.ObjectMeta{Name: "j3", Namespace: "default", Labels: lables},
		},
	}
	err = am.applyJobs(ns, wrongJs)
	assert.Error(t, err)
}

func genApplyRuntime() []runtime.Object {
	ns := "baetyl-edge"
	rs := []runtime.Object{
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		},
		&v1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "sec1", Namespace: ns},
		},
		&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: "cfg1", Namespace: ns},
		},
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: ns},
		},
		&appv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "d1", Namespace: ns},
		},
		&appv1.DaemonSet{
			ObjectMeta: metav1.ObjectMeta{Name: "d1", Namespace: ns},
		},
		&batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{Name: "j1", Namespace: ns},
		},
	}
	return rs
}

func initApplyKubeAMI(t *testing.T) *kubeImpl {
	fc := fake.NewSimpleClientset(genApplyRuntime()...)
	cli := client{
		core:  fc.CoreV1(),
		app:   fc.AppsV1(),
		batch: fc.BatchV1(),
	}
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())
	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)
	ami.Hooks[BaetylSetPodSpec] = SetPodSpecFunc(SetPodSpec)
	return &kubeImpl{cli: &cli, store: sto, knn: "node1", log: log.With()}
}

func TestApplyApplication(t *testing.T) {
	ami := initApplyKubeAMI(t)
	ns := "baetyl-edge"
	app1 := specv1.Application{
		Name:      "svc1",
		Namespace: ns,
		Version:   "a1",
		Services: []specv1.Service{
			{
				Name:  "s1",
				Image: "image1",
				Ports: []specv1.ContainerPort{{
					HostPort:      80,
					ContainerPort: 80,
				}},
			},
			{
				Name:  "s2",
				Image: "image1",
				Ports: []specv1.ContainerPort{{
					HostPort:      80,
					ContainerPort: 80,
				}},
			},
		},
		Volumes: []specv1.Volume{{
			Name:         "cfg1",
			VolumeSource: specv1.VolumeSource{Config: &specv1.ObjectReference{Name: "cfg1", Version: "c1"}},
		}, {
			Name:         "sec1",
			VolumeSource: specv1.VolumeSource{Secret: &specv1.ObjectReference{Name: "sec1", Version: "s1"}},
		}},
	}
	secs := []string{"sec1"}
	err := ami.applyApplication(ns, app1, secs)
	assert.NoError(t, err)

	app2 := specv1.Application{
		Name:      "svc1",
		Namespace: ns,
		Version:   "a2",
		Services: []specv1.Service{
			{
				Name:  "s1",
				Image: "image1",
				Type:  specv1.WorkloadDaemonSet,
				Ports: []specv1.ContainerPort{{
					HostPort:      80,
					ContainerPort: 80,
				}},
			},
		},
		Volumes: []specv1.Volume{{
			Name:         "cfg1",
			VolumeSource: specv1.VolumeSource{Config: &specv1.ObjectReference{Name: "cfg1", Version: "c1"}},
		}, {
			Name:         "sec1",
			VolumeSource: specv1.VolumeSource{Secret: &specv1.ObjectReference{Name: "sec1", Version: "s1"}},
		}},
	}
	err = ami.applyApplication(ns, app2, secs)
	assert.NoError(t, err)

	app3 := specv1.Application{
		Name:      "svc1",
		Namespace: ns,
		Version:   "a2",
		Workload:  specv1.WorkloadJob,
		Services: []specv1.Service{
			{
				Name:  "s1",
				Image: "image1",
			},
		},
		Volumes: []specv1.Volume{{
			Name:         "sec1",
			VolumeSource: specv1.VolumeSource{Secret: &specv1.ObjectReference{Name: "sec1", Version: "s1"}},
		}},
	}
	err = ami.applyApplication(ns, app3, secs)
	assert.NoError(t, err)
}

func TestPrepareService(t *testing.T) {
	ami := initApplyKubeAMI(t)
	ns := "baetyl-edge"
	app := specv1.Application{
		Name:      "svc",
		Namespace: ns,
		Version:   "a1",
		Services: []specv1.Service{
			{
				Name:  "s1",
				Image: "image1",
				Ports: []specv1.ContainerPort{{
					HostPort:      80,
					ContainerPort: 80,
				}},
			},
			{
				Name:  "s2",
				Image: "image1",
				Ports: []specv1.ContainerPort{{
					HostPort:      80,
					ContainerPort: 8080,
				}},
			},
		},
	}
	service := ami.prepareService(ns, app)
	expected := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: app.Name, Namespace: ns, Labels: map[string]string{AppName: app.Name}},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Name:       "s1-0",
				Port:       80,
				TargetPort: intstr.IntOrString{IntVal: 80},
			}, {
				Name:       "s2-0",
				Port:       8080,
				TargetPort: intstr.IntOrString{IntVal: 8080},
			}},
			Selector: map[string]string{
				AppName: app.Name,
			},
		},
	}
	assert.Equal(t, service, expected)

	// bad case
	app2 := specv1.Application{
		Name:      "app2",
		Namespace: ns,
		Version:   "a1",
		Services: []specv1.Service{
			{Name: "svc1"},
		},
	}
	service = ami.prepareService(ns, app2)
	assert.Nil(t, service)
}

func TestPrepareDeploy(t *testing.T) {
	_ = initApplyKubeAMI(t)
	ns := "baetyl-edge"
	app := specv1.Application{
		Name:      "svc",
		Namespace: ns,
		Version:   "a1",
		Replica:   1,
		Services: []specv1.Service{
			{
				Name: "s0",
				VolumeMounts: []specv1.VolumeMount{{
					Name: "cfg",
				}, {
					Name: "sec",
				}, {
					Name: "hostPath",
				}},
				Resources: &specv1.Resources{
					Limits: map[string]string{
						"cpu":    "1",
						"memory": "456456",
					},
				},
			},
		},
	}
	cpuQuan, _ := resource.ParseQuantity("1")
	memoryQuan, _ := resource.ParseQuantity("456456")
	emptyDirQuan, _ := resource.ParseQuantity("29218")
	app.Volumes = []specv1.Volume{
		{
			Name: "cfg",
			VolumeSource: specv1.VolumeSource{
				Config: &specv1.ObjectReference{
					Name: "cfg",
				},
			},
		},
		{
			Name: "sec",
			VolumeSource: specv1.VolumeSource{
				Secret: &specv1.ObjectReference{
					Name: "sec",
				},
			},
		},
		{
			Name: "hostPath",
			VolumeSource: specv1.VolumeSource{
				HostPath: &specv1.HostPathVolumeSource{Path: "/var/lib/baetyl"},
			},
		},
		{
			Name: "emptydir",
			VolumeSource: specv1.VolumeSource{
				EmptyDir: &specv1.EmptyDirVolumeSource{
					Medium:    "Memory",
					SizeLimit: "29218",
				},
			},
		},
	}
	deploy, err := prepareDeploy(ns, &app, nil)
	assert.NoError(t, err)

	replica := new(int32)
	*replica = 1
	expected := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: ns,
			Labels:    map[string]string{},
		},
		Spec: appv1.DeploymentSpec{
			Replicas: replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					AppName: app.Name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						AppName: app.Name,
					},
				},
				Spec: v1.PodSpec{
					Affinity: &v1.Affinity{
						NodeAffinity: &v1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
								NodeSelectorTerms: []v1.NodeSelectorTerm{{MatchExpressions: []v1.NodeSelectorRequirement{
									{
										Key:      MasterRole,
										Operator: v1.NodeSelectorOpExists,
									},
								}}},
							},
						},
					},
					Volumes: []v1.Volume{
						{
							Name: "cfg",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{Name: "cfg"},
								},
							},
						},
						{
							Name: "sec",
							VolumeSource: v1.VolumeSource{
								Secret: &v1.SecretVolumeSource{
									SecretName: "sec",
								},
							},
						},
						{
							Name: "hostPath",
							VolumeSource: v1.VolumeSource{
								HostPath: &v1.HostPathVolumeSource{
									Path: "/var/lib/baetyl",
								},
							},
						},
						{
							Name: "emptydir",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{
									Medium:    "Memory",
									SizeLimit: &emptyDirQuan,
								},
							},
						},
					},
					Containers: []v1.Container{{
						Env: []v1.EnvVar{
							{
								Name:      KubeNodeName,
								ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"}},
							},
						},
						Name: "s0",
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								"cpu":    cpuQuan,
								"memory": memoryQuan,
							},
							Requests: v1.ResourceList{},
						},
						VolumeMounts: []v1.VolumeMount{{
							Name: "cfg",
						}, {
							Name: "sec",
						}, {
							Name: "hostPath",
						}},
					}},
				},
			},
			Strategy: appv1.DeploymentStrategy{
				Type: appv1.RecreateDeploymentStrategyType,
			},
		},
	}
	assert.Equal(t, expected, deploy)
}

func Test_compatibleDeprecatedFiled(t *testing.T) {
	k := initApplyKubeAMI(t)
	// case 0
	app0 := &specv1.Application{
		Name: "a0",
		Services: []specv1.Service{{
			Name:        "s1",
			Labels:      map[string]string{"a": "b"},
			HostNetwork: true,
			Replica:     3,
			JobConfig: &specv1.ServiceJobConfig{
				Completions:   1,
				Parallelism:   2,
				BackoffLimit:  3,
				RestartPolicy: "Never",
			},
			Type: specv1.WorkloadJob,
		}},
	}

	expectApp0 := &specv1.Application{
		Name:        "a0",
		Labels:      map[string]string{"a": "b"},
		HostNetwork: true,
		Replica:     3,
		JobConfig: &specv1.AppJobConfig{
			Completions:   1,
			Parallelism:   2,
			BackoffLimit:  3,
			RestartPolicy: "Never",
		},
		Workload: specv1.WorkloadJob,
		Services: []specv1.Service{{
			Name:        "s1",
			Labels:      map[string]string{"a": "b"},
			HostNetwork: true,
			Replica:     3,
			JobConfig: &specv1.ServiceJobConfig{
				Completions:   1,
				Parallelism:   2,
				BackoffLimit:  3,
				RestartPolicy: "Never",
			},
			Type: specv1.WorkloadJob,
		}},
	}

	k.compatibleDeprecatedFiled(app0)
	assert.EqualValues(t, expectApp0, app0)

	// case 1
	app1 := &specv1.Application{
		Name:     "a1",
		Services: []specv1.Service{},
	}

	expectApp1 := &specv1.Application{
		Name:        "a1",
		Labels:      map[string]string{},
		HostNetwork: false,
		Replica:     1,
		JobConfig: &specv1.AppJobConfig{
			Completions:   0,
			Parallelism:   0,
			BackoffLimit:  0,
			RestartPolicy: "Never",
		},
		Workload: specv1.WorkloadDeployment,
		Services: []specv1.Service{},
	}

	k.compatibleDeprecatedFiled(app1)
	assert.EqualValues(t, expectApp1, app1)
}

func Test_cutSysServiceRandSuffix(t *testing.T){
	assert.Equal(t, "baetyl-",cutSysServiceRandSuffix("baetyl-"))
	assert.Equal(t, "baetyl-broker",cutSysServiceRandSuffix("baetyl-broker"))
	assert.Equal(t, "baetyl-broker",cutSysServiceRandSuffix("baetyl-broker-"))
	assert.Equal(t, "baetyl-rule",cutSysServiceRandSuffix("baetyl-rule"))
	assert.Equal(t, "baetyl-broker",cutSysServiceRandSuffix("baetyl-broker-123"))
	assert.Equal(t, "baetyl-init-baetyl",cutSysServiceRandSuffix("baetyl-init-baetyl-broker"))
	assert.Equal(t, "baetyl-device-ops",cutSysServiceRandSuffix("baetyl-device-ops-123"))
	assert.Equal(t, "baetyl-device-ops-123-456",cutSysServiceRandSuffix("baetyl-device-ops-123-456-789"))
	assert.Equal(t, "app",cutSysServiceRandSuffix("app"))
	assert.Equal(t, "app-123",cutSysServiceRandSuffix("app-123"))
	assert.Equal(t, "",cutSysServiceRandSuffix(""))
	assert.Equal(t, "baetyl-init_123_21312-32323-baetyl",cutSysServiceRandSuffix("baetyl-init_123_21312-32323-baetyl-21312"))
	assert.Equal(t, "rule-baetyl",cutSysServiceRandSuffix("rule-baetyl"))
}
