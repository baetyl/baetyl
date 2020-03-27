package ami

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-go/spec/crd"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestToConfigMap(t *testing.T) {
	config := &crd.Configuration{
		Name:      "cfg",
		Namespace: "baetyl-edge",
		Data: map[string]string{
			"test-key": "test-val",
		},
	}
	expected := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "cfg", Namespace: "baetyl-edge"},
		Data: map[string]string{
			"test-key": "test-val",
		},
	}
	configMap, err := toConfigMap(config)
	assert.NoError(t, err)
	assert.Equal(t, configMap, expected)
}

func TestToSecret(t *testing.T) {
	sec := &crd.Secret{
		Name:      "sec",
		Namespace: "baetyl-edge",
	}
	secKey := "sec-key"
	secVal := "sec-val"
	sec.Data = map[string][]byte{
		secKey: []byte(secVal),
	}
	secret, err := toSecret(sec)
	assert.NoError(t, err)
	expected := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "sec", Namespace: "baetyl-edge"},
	}
	expected.Data = map[string][]byte{
		secKey: []byte(secVal),
	}
	assert.Equal(t, secret, expected)

	// registry
	reg := &crd.Secret{Name: "registry"}
	reg.Labels = map[string]string{
		crd.SecretLabel: crd.SecretRegistry,
	}
	reg.Data = map[string][]byte{
		RegistryAddress:  []byte("server"),
		RegistryUsername: []byte("test"),
		RegistryPassword: []byte("1234"),
	}
	registry, err := toSecret(reg)
	assert.NoError(t, err)
	expected = &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "registry"},
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
	assert.Equal(t, registry, expected)
}

func TestToService(t *testing.T) {
	svcName := "svc"
	namespace := "baetyl-edge"
	svc := &crd.Service{
		Name:  svcName,
		Ports: []crd.ContainerPort{{ContainerPort: 80}, {ContainerPort: 8080}},
	}
	service, err := toService(namespace, svc)
	assert.NoError(t, err)
	expected := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: svcName, Namespace: namespace},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{{
				Port:       80,
				TargetPort: intstr.IntOrString{IntVal: 80},
			}, {
				Port:       8080,
				TargetPort: intstr.IntOrString{IntVal: 8080},
			}},
			Selector: map[string]string{
				"baetyl": svcName,
			},
			ClusterIP: "None",
		},
	}
	assert.Equal(t, service, expected)
	svc = &crd.Service{
		Name: svcName,
	}
	service, err = toService(namespace, svc)
	assert.NoError(t, err)
	assert.Nil(t, service)
}

func TestToDeploy(t *testing.T) {
	namespace := "baetyl-edge"
	svcName := "svc"
	app := &crd.Application{
		Name:      "app",
		Namespace: namespace,
		Version:   "a1",
	}
	svc := &crd.Service{
		Name:    svcName,
		Replica: 1,
		VolumeMounts: []crd.VolumeMount{{
			Name: "cfg",
		}, {
			Name: "sec",
		}, {
			Name: "hostPath",
		}},
		Restart: &crd.RestartPolicyInfo{Policy: "Never"},
		Resources: &crd.Resources{
			Limits: map[string]string{
				"cpu":    "1",
				"memory": "456456",
			},
		},
	}
	cpuQuan, _ := resource.ParseQuantity("1")
	memoryQuan, _ := resource.ParseQuantity("456456")
	volumes := []crd.Volume{{
		Name: "cfg",
		VolumeSource: crd.VolumeSource{
			Config: &crd.ObjectReference{
				Name: "cfg",
			},
		},
	}, {
		Name: "sec",
		VolumeSource: crd.VolumeSource{
			Secret: &crd.ObjectReference{
				Name: "sec",
			},
		},
	}, {
		Name: "hostPath",
		VolumeSource: crd.VolumeSource{
			HostPath: &crd.HostPathVolumeSource{Path: "/var/lib/baetyl"},
		},
	}}
	deploy, err := toDeploy(app, svc, volumes, nil)
	assert.NoError(t, err)
	replica := new(int32)
	*replica = 1
	expected := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: namespace,
			Labels: map[string]string{
				"baetyl": "baetyl",
			}},
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
						"baetyl":    "baetyl",
						AppName:     app.Name,
						AppVersion:  app.Version,
						ServiceName: svcName,
					},
				},
				Spec: v1.PodSpec{
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
					RestartPolicy: "Never",
				},
			},
			Strategy: appv1.DeploymentStrategy{
				Type: appv1.RecreateDeploymentStrategyType,
			},
		},
	}
	assert.Equal(t, deploy, expected)
}

func TestApplyDeploy(t *testing.T) {
	ami := initKubeAMI(t)
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
	err := ami.applyDeploys(ds)
	assert.NoError(t, err)

	wrongDs := map[string]*appv1.Deployment{
		"d1": {
			ObjectMeta: metav1.ObjectMeta{Name: "d1", Namespace: "default", Labels: lables},
		},
		"d3": {
			ObjectMeta: metav1.ObjectMeta{Name: "d3", Namespace: "default", Labels: lables},
		},
	}
	err = ami.applyDeploys(wrongDs)
	assert.Error(t, err)

	deleteDs := map[string]*appv1.Deployment{
		"d3": {
			ObjectMeta: metav1.ObjectMeta{Name: "d3", Namespace: ns, Labels: lables},
		},
	}
	err = ami.applyDeploys(deleteDs)
	assert.NoError(t, err)
	_, err = ami.cli.App.Deployments(ns).Get("d1", metav1.GetOptions{})
	assert.Error(t, err)
}

func TestApplySecret(t *testing.T) {
	ami := initKubeAMI(t)
	ns := "baetyl-edge"
	secs := map[string]*v1.Secret{
		"sec1": {
			ObjectMeta: metav1.ObjectMeta{Name: "sec1", Namespace: ns},
		},
		"sec2": {
			ObjectMeta: metav1.ObjectMeta{Name: "sec2", Namespace: ns},
		},
	}
	err := ami.applySecrets(secs)
	assert.NoError(t, err)
	wrongSecs := map[string]*v1.Secret{
		"sec1": {
			ObjectMeta: metav1.ObjectMeta{Name: "sec1", Namespace: "default"},
		},
		"sec3": {
			ObjectMeta: metav1.ObjectMeta{Name: "sec3", Namespace: "default"},
		},
	}
	err = ami.applySecrets(wrongSecs)
	assert.Error(t, err)
}

func TestApplyConfigMap(t *testing.T) {
	ami := initKubeAMI(t)
	ns := "baetyl-edge"
	cfgs := map[string]*v1.ConfigMap{
		"cfg1": {
			ObjectMeta: metav1.ObjectMeta{Name: "cfg1", Namespace: ns},
		},
		"cfg2": {
			ObjectMeta: metav1.ObjectMeta{Name: "cfg2", Namespace: ns},
		},
	}
	err := ami.applyConfigMaps(cfgs)
	assert.NoError(t, err)
	wrongCfgs := map[string]*v1.ConfigMap{
		"cfg1": {
			ObjectMeta: metav1.ObjectMeta{Name: "cfg1", Namespace: "default"},
		},
		"cfg3": {
			ObjectMeta: metav1.ObjectMeta{Name: "cfg3", Namespace: "default"},
		},
	}
	err = ami.applyConfigMaps(wrongCfgs)
	assert.Error(t, err)
}

func TestApplyService(t *testing.T) {
	ami := initKubeAMI(t)
	ns := "baetyl-edge"
	svcs := map[string]*v1.Service{
		"svc1": {
			ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: ns},
		},
		"svc2": {
			ObjectMeta: metav1.ObjectMeta{Name: "svc2", Namespace: ns},
		},
	}
	err := ami.applyServices(svcs)
	assert.NoError(t, err)
	wrongSvcs := map[string]*v1.Service{
		"svc1": {
			ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: "default"},
		},
		"svc3": {
			ObjectMeta: metav1.ObjectMeta{Name: "svc3", Namespace: "default"},
		},
	}
	err = ami.applyServices(wrongSvcs)
	assert.Error(t, err)
}

func genRuntime() []runtime.Object {
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
			ObjectMeta: metav1.ObjectMeta{Name: "d1", Namespace: ns, Labels: map[string]string{"baetyl": "baetyl"}},
		},
	}
	return rs
}

func initKubeAMI(t *testing.T) *kubeImpl {
	fc := fake.NewSimpleClientset(genRuntime()...)
	cli := Client{
		Namespace: "baetyl-edge",
		Core:      fc.CoreV1(),
		App:       fc.AppsV1(),
	}
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())
	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)
	return &kubeImpl{cli: &cli, store: sto}
}
