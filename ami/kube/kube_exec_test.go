package kube

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/gorilla/websocket"
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
	f, err := os.CreateTemp("", t.Name())
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
	no, err := am.cli.core.Nodes().Get(context.TODO(), "node1", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.EqualValues(t, oldLabels, no.Labels)

	err = am.UpdateNodeLabels("node1", newLabels)
	assert.NoError(t, err)

	no, err = am.cli.core.Nodes().Get(context.TODO(), "node1", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.EqualValues(t, newLabels, no.Labels)
}

func TestWebsocket(t *testing.T) {
	impl := initExecKubeAMI(t)
	str := "testEOF"
	port, err := GetFreePort()
	assert.NoError(t, err)
	option := &ami.DebugOptions{
		WebsocketOptions: ami.WebsocketOptions{
			Host: "127.0.0.1:" + strconv.Itoa(port),
			Path: "",
		},
	}
	pipe := ami.Pipe{}
	pipe.InReader, pipe.InWriter = io.Pipe()
	pipe.OutReader, pipe.OutWriter = io.Pipe()
	svc := newWsServer(port)

	go func() {
		dt := make([]byte, 10240)
		n, readErr := pipe.OutReader.Read(dt)
		assert.NoError(t, readErr)
		assert.Equal(t, string(dt[0:n]), str)
	}()
	go func() {
		time.Sleep(time.Millisecond)
		_, err = pipe.InWriter.Write([]byte(str))
		assert.NoError(t, err)
	}()
	time.Sleep(time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(time.Second * 1)
		cancel()
		svc.Shutdown(ctx)
	}()
	impl.RemoteWebsocket(ctx, option, pipe)
}

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// 升级为 WebSocket 连接
	conn, connErr := upgrader.Upgrade(w, r, nil)
	if connErr != nil {
		fmt.Println(connErr)
		return
	}
	defer conn.Close()
	for {
		// 读取消息
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// 发送消息
		err = conn.WriteMessage(msgType, msg)
		if err != nil {
			break
		}
	}
}

func newWsServer(port int) *http.Server {
	http.HandleFunc("/", handleConnections)
	server := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: nil}
	go func() {
		server.ListenAndServe()
	}()
	return server
}
