package kube

import (
	"bytes"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/baetyl/baetyl-go/v2/log"
	specV1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/v2/config"
)

func TestRPCApp(t *testing.T) {
	ami := initRPCAMI(t)

	// req rpc fail
	req := &specV1.RPCRequest{
		App:    "app",
		Method: "unknown",
		System: true,
		Params: "",
		Header: map[string]string{},
		Body:   "",
	}
	_, err := ami.RPCApp("", req)
	assert.NotNil(t, err)

	// req0 rpc fail
	req0 := &specV1.RPCRequest{
		App:    "app",
		Method: "get",
		System: true,
		Params: "",
		Header: map[string]string{},
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	_, err = ami.RPCApp(s.URL, req0)
	assert.NotNil(t, err)
	s.Close()

	// req1 rpc post success
	req1 := &specV1.RPCRequest{
		App:    "app",
		Method: "post",
		System: true,
		Params: "",
		Header: map[string]string{},
		Body:   "",
	}
	res1 := &specV1.RPCResponse{
		StatusCode: http.StatusOK,
		Header:     map[string][]string{},
		Body:       []byte{},
	}

	s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusOK)
		var buf bytes.Buffer
		err = binary.Write(&buf, binary.BigEndian, res1)
		_, err = w.Write(buf.Bytes())
		assert.NoError(t, err)
	}))
	res, err := ami.RPCApp(s.URL, req1)
	assert.NoError(t, err)
	assert.Equal(t, res.StatusCode, http.StatusOK)
	s.Close()

	// req2 rpc delete success
	req2 := &specV1.RPCRequest{
		App:    "app",
		Method: "delete",
		System: true,
		Params: "",
		Header: map[string]string{},
		Body:   "",
	}

	s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	res, err = ami.RPCApp(s.URL, req2)
	assert.NoError(t, err)
	s.Close()

	// req3 rpc put success
	req3 := &specV1.RPCRequest{
		App:    "app",
		Method: "put",
		System: true,
		Params: "",
		Header: map[string]string{},
		Body:   "",
	}

	s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	res, err = ami.RPCApp(s.URL, req3)
	assert.NoError(t, err)
	s.Close()
}

func initRPCAMI(t *testing.T) *kubeImpl {
	return &kubeImpl{
		cli:   nil,
		store: nil,
		knn:   "node1",
		conf: &config.KubeConfig{
			LogConfig: config.KubernetesLogConfig{
				Follow:     false,
				Previous:   false,
				TimeStamps: false,
			},
		},
		log: log.With(),
	}
}
