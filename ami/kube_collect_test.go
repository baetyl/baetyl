package ami

import (
	"fmt"
	"github.com/baetyl/baetyl-core/store"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	appv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestCollectNodeInfo(t *testing.T) {
	ami := initCollectKubeAMI(t)
	node, err := ami.cli.Core.Nodes().Get("node1", metav1.GetOptions{})
	assert.NoError(t, err)
	res := ami.collectNodeInfo(node)
	expected := specv1.NodeInfo{
		Hostname:         "hostname",
		Address:          "nodeip",
		Arch:             "arch",
		KernelVersion:    "kernel",
		OS:               "os",
		ContainerRuntime: "runtime",
		MachineID:        "machine",
		BootID:           "boot",
		SystemUUID:       "system",
		OSImage:          "image",
	}
	assert.Equal(t, res, expected)
}

func initCollectKubeAMI(t *testing.T) *kubeImpl {
	fc := fake.NewSimpleClientset(genCollectRuntime()...)
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
	return &kubeImpl{cli: &cli, store: sto, knn: "node1"}
}

func genCollectRuntime() []runtime.Object {
	ns := "baetyl-edge"
	rs := []runtime.Object{
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node1", Namespace: ns},
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
					{Type: v1.NodeInternalIP, Address: "nodeip"},
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
