package kube

import (
	"context"
	"fmt"
	"os"
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

func TestGetAppStatus(t *testing.T) {
	var status specv1.Status = ""
	var replicas int32 = 2
	infos := map[string]specv1.InstanceStats{
		"ins-1": {
			Status: specv1.Running,
		},
		"ins-2": {
			Status: specv1.Running,
		},
	}
	res := getAppStatus(status, replicas, infos)
	expected := specv1.Running
	assert.Equal(t, expected, res)

	infos["ins-2"] = specv1.InstanceStats{Status: specv1.Pending}
	res = getAppStatus(status, replicas, infos)
	expected = specv1.Pending
	assert.Equal(t, expected, res)

	infos["ins-2"] = specv1.InstanceStats{Status: specv1.Failed}
	res = getAppStatus(status, replicas, infos)
	expected = specv1.Pending
	assert.Equal(t, expected, res)

	infos["ins-1"] = specv1.InstanceStats{Status: specv1.Succeeded}
	infos["ins-2"] = specv1.InstanceStats{Status: specv1.Succeeded}
	res = getAppStatus(status, replicas, infos)
	expected = specv1.Running
	assert.Equal(t, expected, res)

	infos["ins-2"] = specv1.InstanceStats{Status: specv1.Failed}
	res = getAppStatus(status, replicas, infos)
	expected = specv1.Pending
	assert.Equal(t, expected, res)

	infos["ins-2"] = specv1.InstanceStats{Status: specv1.Unknown}
	res = getAppStatus(status, replicas, infos)
	expected = specv1.Unknown
	assert.Equal(t, expected, res)

	infos["ins-1"] = specv1.InstanceStats{Status: specv1.Succeeded}
	infos["ins-2"] = specv1.InstanceStats{Status: specv1.Succeeded}
	infos["ins-3"] = specv1.InstanceStats{Status: specv1.Pending}
	res = getAppStatus(status, replicas, infos)
	expected = specv1.Running
	assert.Equal(t, expected, res)

	status = specv1.Pending
	res = getAppStatus(status, replicas, infos)
	expected = specv1.Pending
	assert.Equal(t, expected, res)
}

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
			Role:             "worker",
		},
	}
	assert.EqualValues(t, expected["node1"], res["node1"])
}

func TestGetContainerStatus(t *testing.T) {
	info := &v1.ContainerStatus{}

	info.State.Waiting = &v1.ContainerStateWaiting{Reason: "ImagePullErr"}
	state, _ := getContainerStatus(info)
	assert.Equal(t, state, specv1.ContainerWaiting)
	info.State.Waiting = nil

	info.State.Running = &v1.ContainerStateRunning{}
	state, _ = getContainerStatus(info)
	assert.Equal(t, state, specv1.ContainerRunning)
	info.State.Running = nil

	info.State.Terminated = &v1.ContainerStateTerminated{Reason: "Completed"}
	state, _ = getContainerStatus(info)
	assert.Equal(t, state, specv1.ContainerTerminated)
	info.State.Terminated = nil

	_, reason := getContainerStatus(info)
	assert.Equal(t, reason, "status unknown")
}

func initCollectKubeAMI(t *testing.T) *kubeImpl {
	fc := fake.NewSimpleClientset(genCollectRuntime()...)
	cli := client{
		core: fc.CoreV1(),
		app:  fc.AppsV1(),
	}
	node, err := fc.CoreV1().Nodes().Get(context.TODO(), "node1", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, node)

	se, err := fc.CoreV1().Secrets("baetyl-edge").Get(context.TODO(), "sec1", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, se)

	f, err := os.CreateTemp("", t.Name())
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
