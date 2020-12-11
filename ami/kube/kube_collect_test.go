package kube

import (
	"fmt"
	"io/ioutil"
	"testing"

	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/stretchr/testify/assert"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/baetyl/baetyl/v2/store"
)

func TestCollectNodeInfo(t *testing.T) {
	ami := initCollectKubeAMI(t)
	res, err := ami.CollectNodeInfo()
	assert.NoError(t, err)
	expected := map[string]interface{}{
		"node1": &specv1.NodeInfo{
			Hostname:         "hostname",
			Arch:             "arch",
			KernelVersion:    "kernel",
			OS:               "os",
			ContainerRuntime: "runtime",
			MachineID:        "machine",
			BootID:           "boot",
			SystemUUID:       "system",
			OSImage:          "image",
		},
	}
	assert.EqualValues(t, expected["node1"], res["node1"])
}

func initCollectKubeAMI(t *testing.T) *kubeImpl {
	fc := fake.NewSimpleClientset(genCollectRuntime()...)
	cli := client{
		core: fc.CoreV1(),
		app:  fc.AppsV1(),
	}
	node, err := fc.CoreV1().Nodes().Get("node1", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, node)

	se, err := fc.CoreV1().Secrets("baetyl-edge").Get("sec1", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, se)

	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())
	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)
	return &kubeImpl{cli: &cli, store: sto, knn: "node1"}
}

func genCollectRuntime() []runtime.Object {
	ns := "baetyl-edge"
	rs := []runtime.Object{
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node1"},
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
