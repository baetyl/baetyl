package ami

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/baetyl/baetyl-go/log"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl/store"
	"github.com/stretchr/testify/assert"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
)

func TestApplyApplication(t *testing.T) {
	ami := initApplyKubeAMI(t)
	ns := "baetyl-edge"
	aname := "app1"
	ver := "a1"
	sname := "svc1"
	app := specv1.Application{
		Name:      aname,
		Namespace: ns,
		Version:   ver,
		Services: []specv1.Service{{
			Name:  sname,
			Image: "image1",
			Ports: []specv1.ContainerPort{{
				HostPort:      80,
				ContainerPort: 80,
			}},
		}},
		Volumes: []specv1.Volume{{
			Name:         "cfg1",
			VolumeSource: specv1.VolumeSource{Config: &specv1.ObjectReference{Name: "cfg1", Version: "c1"}},
		}, {
			Name:         "sec1",
			VolumeSource: specv1.VolumeSource{Secret: &specv1.ObjectReference{Name: "sec1", Version: "s1"}},
		}},
	}
	secs := []string{"sec1"}
	err := ami.ApplyApplication(ns, app, secs)
	assert.NoError(t, err)
}

func TestPrepareService(t *testing.T) {
	ami := initApplyKubeAMI(t)
	svcName := "svc"
	ns := "baetyl-edge"
	svc := &specv1.Service{
		Name:  svcName,
		Ports: []specv1.ContainerPort{{ContainerPort: 80}, {ContainerPort: 8080}},
	}
	service := ami.prepareService(ns, "", svc)
	expected := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: svcName, Namespace: ns, Labels: map[string]string{AppName: ""}},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Port:       80,
				TargetPort: intstr.IntOrString{IntVal: 80},
			}, {
				Port:       8080,
				TargetPort: intstr.IntOrString{IntVal: 8080},
			}},
			Selector: map[string]string{
				ServiceName: svcName,
			},
			ClusterIP: "None",
		},
	}
	assert.Equal(t, service, expected)
	svc = &specv1.Service{
		Name: svcName,
	}
	service = ami.prepareService(ns, "", svc)
	assert.Nil(t, service)
}

func TestPrepareDeploy(t *testing.T) {
	ami := initApplyKubeAMI(t)
	ns := "baetyl-edge"
	svcName := "svc"
	app := specv1.Application{
		Name:      "app",
		Namespace: ns,
		Version:   "a1",
	}
	svc := specv1.Service{
		Name:    svcName,
		Replica: 1,
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
	}
	cpuQuan, _ := resource.ParseQuantity("1")
	memoryQuan, _ := resource.ParseQuantity("456456")
	app.Volumes = []specv1.Volume{{
		Name: "cfg",
		VolumeSource: specv1.VolumeSource{
			Config: &specv1.ObjectReference{
				Name: "cfg",
			},
		},
	}, {
		Name: "sec",
		VolumeSource: specv1.VolumeSource{
			Secret: &specv1.ObjectReference{
				Name: "sec",
			},
		},
	}, {
		Name: "hostPath",
		VolumeSource: specv1.VolumeSource{
			HostPath: &specv1.HostPathVolumeSource{Path: "/var/lib/baetyl"},
		},
	}}
	deploy, err := ami.prepareDeploy(ns, app, svc, nil)
	assert.NoError(t, err)
	replica := new(int32)
	*replica = 1
	expected := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: ns,
			Labels:    map[string]string{AppName: app.Name},
		},
		Spec: appv1.DeploymentSpec{
			Replicas: replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					ServiceName: svcName,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						AppName:     app.Name,
						AppVersion:  app.Version,
						ServiceName: svcName,
					},
				},
				Spec: v1.PodSpec{
					NodeName: "node1",
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
					},
					Containers: []v1.Container{{
						Env: []v1.EnvVar{
							{Name: KubeNodeName, Value: "node1"}},
						Name: "svc",
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								"cpu":    cpuQuan,
								"memory": memoryQuan,
							},
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
	assert.Equal(t, deploy, expected)
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

func TestDeleteApplication(t *testing.T) {
	ami := initApplyKubeAMI(t)
	ns := "baetyl-edge"
	sname := "svc1"
	app := specv1.Application{
		Name:      "app1",
		Namespace: ns,
		Version:   "a1",
		Services: []specv1.Service{{
			Name:  sname,
			Image: "image1",
			Ports: []specv1.ContainerPort{{
				HostPort:      80,
				ContainerPort: 80,
			}},
		}},
	}
	err := ami.ApplyApplication(ns, app, nil)
	assert.NoError(t, err)
	d, err := ami.cli.app.Deployments(ns).Get(sname, metav1.GetOptions{})
	assert.NotNil(t, d)
	assert.NoError(t, err)
	s, err := ami.cli.core.Services(ns).Get(sname, metav1.GetOptions{})
	assert.NotNil(t, s)
	assert.NoError(t, err)

	err = ami.DeleteApplication(ns, app.Name)
	assert.NoError(t, err)
	d, _ = ami.cli.app.Deployments(ns).Get(sname, metav1.GetOptions{})
	assert.Nil(t, d)
	s, _ = ami.cli.core.Services(ns).Get(sname, metav1.GetOptions{})
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
	err := ami.ApplySecrets(ns, secs)
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
	err = ami.ApplySecrets(ns, secs)
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
	err = ami.ApplySecrets(ns, regs)
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
	err := ami.ApplyConfigurations(ns, cfgs)
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

func genApplyRuntime() []runtime.Object {
	ns := "baetyl-edge"
	rs := []runtime.Object{
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
	}
	return rs
}

func initApplyKubeAMI(t *testing.T) *kubeImpl {
	fc := fake.NewSimpleClientset(genApplyRuntime()...)
	cli := client{
		core: fc.CoreV1(),
		app:  fc.AppsV1(),
	}
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())
	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)
	return &kubeImpl{cli: &cli, store: sto, knn: "node1", log: log.With()}
}
