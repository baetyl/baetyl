package kube

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/baetyl/baetyl/v2/ami"
	"github.com/baetyl/baetyl/v2/store"
)

func genExecRuntime() []runtime.Object {
	ns := "baetyl-edge"
	rs := []runtime.Object{
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: ns},
		},
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node1",
				Labels: map[string]string{
					"beta.kubernetes.io/arch":        "amd64",
					"beta.kubernetes.io/os":          "linux",
					"kubernetes.io/arch":             "amd64",
					"kubernetes.io/hostname":         "docker-desktop",
					"kubernetes.io/os":               "linux",
					"node-role.kubernetes.io/master": "",
				},
			},
			Status: v1.NodeStatus{
				NodeInfo: v1.NodeSystemInfo{
					Architecture:            "arch",
					KernelVersion:           "kernel",
					OperatingSystem:         "os",
					ContainerRuntimeVersion: "runtime",
					MachineID:               "machine",
					OSImage:                 "image",
					BootID:                  "boot",
					SystemUUID:              "system",
				},
				Addresses: []v1.NodeAddress{
					{Type: v1.NodeHostName, Address: "hostname"},
				},
				Capacity: v1.ResourceList{
					"cpu":    *resource.NewQuantity(2, resource.DecimalSI),
					"memory": *resource.NewQuantity(200, resource.DecimalSI),
				},
			},
		},
	}
	return rs
}

func initExecKubeAMI(t *testing.T) *kubeImpl {
	fc := fake.NewSimpleClientset(genExecRuntime()...)
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
	ami.Hooks[BaetylSetPodSpec] = SetPodSpecFunc(SetPodSpec)
	return &kubeImpl{cli: &cli, store: sto, knn: "node1", log: log.With()}
}

func TestUpdateNodeLabels(t *testing.T) {
	am := initExecKubeAMI(t)
	oldLabels := map[string]string{
		"beta.kubernetes.io/arch":        "amd64",
		"beta.kubernetes.io/os":          "linux",
		"kubernetes.io/arch":             "amd64",
		"kubernetes.io/hostname":         "docker-desktop",
		"kubernetes.io/os":               "linux",
		"node-role.kubernetes.io/master": "",
	}
	newLabels := map[string]string{
		"a": "b",
	}
	no, err := am.cli.core.Nodes().Get("node1", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.EqualValues(t, oldLabels, no.Labels)

	err = am.UpdateNodeLabels("node1", newLabels)
	assert.NoError(t, err)

	no, err = am.cli.core.Nodes().Get("node1", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.EqualValues(t, newLabels, no.Labels)
}
