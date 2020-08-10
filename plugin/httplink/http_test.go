package httplink

import (
	"encoding/json"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/baetyl/baetyl/plugin"
	"github.com/stretchr/testify/assert"
	gohttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestRequest(t *testing.T) {
	apps := map[string]interface{}{"app1": "123"}
	de := &specv1.Desire{"apps": apps}
	data1, err := json.Marshal(de)
	assert.NoError(t, err)
	appRes := specv1.DesireResponse{Values: []specv1.ResourceValue{
		{
			ResourceInfo: specv1.ResourceInfo{Kind: specv1.KindApplication},
			Value: specv1.VariableValue{Value: &specv1.Application{Name:    "app1", Version: "123"}}},
		},
	}
	data2, err := json.Marshal(appRes)
	assert.NoError(t, err)
	tlssvr, err := utils.NewTLSConfigServer(utils.Certificate{CA: "./testcert/ca.pem", Key: "./testcert/server.key", Cert: "./testcert/server.pem"})
	assert.NoError(t, err)
	assert.NotNil(t, tlssvr)
	ms := httptest.NewTLSServer(gohttp.HandlerFunc(func(w gohttp.ResponseWriter, r *gohttp.Request) {
		if r.URL.Path == "/v1/sync/report" {
			w.Write(data1)
		} else if r.URL.Path == "/v1/sync/desire" {
			w.Write(data2)
		}
	}))
	assert.NotNil(t, ms)
	defer ms.Close()

	var cfg Config
	err = utils.UnmarshalYAML(nil, &cfg)
	assert.NoError(t, err)
	cfg.HTTPLink.HTTP = http.ClientConfig{
		Address: ms.URL,
		Certificate: utils.Certificate{
			CA:                 "./testcert/ca.pem",
			Key:                "./testcert/client.key",
			Cert:               "./testcert/client.pem",
			InsecureSkipVerify: true,
		},
	}
	ops, err := cfg.HTTPLink.HTTP.ToClientOptions()
	assert.NoError(t, err)
	link := &httpLink{
		cfg:  cfg,
		http: http.NewClient(ops),
		log:  log.With(log.Any("plugin", "httplink")),
	}
	msg := &plugin.Message{
		Kind: plugin.ReportKind,
	}
	res, err := link.Request(msg)
	assert.NotNil(t, res)
	desire := res.Content.(specv1.Desire)
	assert.Equal(t, desire["apps"], apps)
	assert.Equal(t, res.Kind, plugin.ReportKind)

	msg = &plugin.Message{
		Kind: plugin.DesireKind,
	}
	res, err = link.Request(msg)
	assert.NotNil(t, res)
	assert.NoError(t, err)
	desireRes := res.Content.(specv1.DesireResponse)
	aRes := desireRes.Values[0].App()
	assert.Equal(t, aRes.Name, "app1")
	assert.Equal(t, aRes.Version, "123")
}
