package kube

import (
	"testing"

	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/stretchr/testify/assert"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestApplyApplicationV2(t *testing.T) {
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
	err := ami.applyApplicationV2(ns, app1, secs)
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
	err = ami.applyApplicationV2(ns, app2, secs)
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
	err = ami.applyApplicationV2(ns, app3, secs)
	assert.NoError(t, err)
}

func TestPrepareServiceV2(t *testing.T) {
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
	service := ami.prepareServiceV2(ns, app)
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
	service = ami.prepareServiceV2(ns, app2)
	assert.Nil(t, service)
}

func TestPrepareDeployV2(t *testing.T) {
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
	deploy, err := prepareDeployV2(ns, &app, nil)
	assert.NoError(t, err)

	replica := new(int32)
	*replica = 1
	expected := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: ns,
			Labels: map[string]string{
			},
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
