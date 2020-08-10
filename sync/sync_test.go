package sync

import (
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mock"
	"github.com/baetyl/baetyl/mock/plugin"
	"github.com/golang/mock/gomock"
	"io/ioutil"
	"testing"
	"time"

	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/node"
	"github.com/baetyl/baetyl/store"
	"github.com/stretchr/testify/assert"
)

func TestSync_Report(t *testing.T) {
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	nod, err := node.NewNode(sto)
	assert.NoError(t, err)
	assert.NotNil(t, nod)

	bi := &specv1.Desire{"apps": map[string]interface{}{"app1": "123"}}
	data, err := json.Marshal(bi)
	assert.NoError(t, err)

	tlssvr, err := utils.NewTLSConfigServer(utils.Certificate{CA: "./testcert/ca.pem", Key: "./testcert/server.key", Cert: "./testcert/server.pem"})
	assert.NoError(t, err)
	assert.NotNil(t, tlssvr)
	ms := mock.NewServer(tlssvr, mock.NewResponse(200, data))
	assert.NotNil(t, ms)
	defer ms.Close()

	sc := config.SyncConfig{}
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)
	sc.HTTP.Address = ms.URL
	sc.HTTP.CA = "./testcert/ca.pem"
	sc.HTTP.Key = "./testcert/client.key"
	sc.HTTP.Cert = "./testcert/client.pem"
	sc.HTTP.InsecureSkipVerify = true
	sc.ReportInterval = time.Millisecond * 500

	mockCtl := gomock.NewController(t)
	link := plugin.NewMockLink(mockCtl)
	ops, err := sc.HTTP.ToClientOptions()
	assert.NoError(t, err)

	syn := &sync{
		link:  link,
		cfg:   sc,
		store: sto,
		nod:   nod,
		http:  http.NewClient(ops),
		log:   log.With(log.Any("test", "sync")),
	}

	err = syn.reportAndDesire()
	assert.NoError(t, err)
	no, _ := syn.nod.Get()
	assert.Equal(t, specv1.Desire{"apps": map[string]interface{}{"app1": "123"}}, no.Desire)

	sc = config.SyncConfig{}
	_, err = NewSync(config.Config{Sync: sc}, sto, nod)
	assert.Error(t, err, ErrSyncTLSConfigMissing.Error())

	sc.Cloud.HTTP.Cert = "./testcert/notexist.pem"
	_, err = NewSync(config.Config{Sync: sc}, sto, nod)
	assert.Error(t, err)

	ms = mock.NewServer(tlssvr, mock.NewResponse(200, []byte{}))
	sc = config.SyncConfig{}
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)
	sc.Cloud.HTTP.Address = ms.URL
	sc.Cloud.HTTP.CA = "./testcert/ca.pem"
	sc.Cloud.HTTP.Key = "./testcert/client.key"
	sc.Cloud.HTTP.Cert = "./testcert/client.pem"
	sc.Cloud.HTTP.InsecureSkipVerify = true
	ops, err = sc.Cloud.HTTP.ToClientOptions()
	assert.NoError(t, err)
	syn = &sync{cfg: sc, store: sto, nod: nod, http: http.NewClient(ops), log: log.With(log.Any("test", "sync"))}
	syn.Start()

	ms = mock.NewServer(tlssvr, mock.NewResponse(500, []byte{}))
	sc = config.SyncConfig{}
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)
	sc.Cloud.HTTP.Address = ms.URL
	sc.Cloud.HTTP.CA = "./testcert/ca.pem"
	sc.Cloud.HTTP.Key = "./testcert/client.key"
	sc.Cloud.HTTP.Cert = "./testcert/client.pem"
	sc.Cloud.HTTP.InsecureSkipVerify = true
	ops, err = sc.Cloud.HTTP.ToClientOptions()
	assert.NoError(t, err)
	syn = &sync{
		cfg:   sc,
		store: sto,
		nod:   nod,
		http:  http.NewClient(ops),
		log:   log.With(log.Any("test", "sync")),
	}
	syn.Start()
	syn.Close()
}

func TestSyncResource(t *testing.T) {
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	nod, err := node.NewNode(sto)
	assert.NoError(t, err)
	assert.NotNil(t, nod)

	appName := "baetyl-core-923jdsn"
	appVer := "32451"
	bi := &specv1.Desire{"sysapps": []specv1.AppInfo{
		{
			Name:    appName,
			Version: appVer,
		},
	}}
	data, err := json.Marshal(bi)
	assert.NoError(t, err)

	appRes := &specv1.Application{
		Name:      appName,
		Namespace: "baetyl-edge",
		Version:   appVer,
	}
	appData, err := json.Marshal(appRes)
	assert.NoError(t, err)

	tlssvr, err := utils.NewTLSConfigServer(utils.Certificate{CA: "./testcert/ca.pem", Key: "./testcert/server.key", Cert: "./testcert/server.pem"})
	assert.NoError(t, err)
	assert.NotNil(t, tlssvr)

	resp := []*mock.Response{
		mock.NewResponse(200, data),
		mock.NewResponse(200, appData),
		mock.NewResponse(200, data),
	}
	ms := mock.NewServer(tlssvr, resp...)
	assert.NotNil(t, ms)
	defer ms.Close()

	sc := config.SyncConfig{}
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)
	sc.Cloud.HTTP.Address = ms.URL
	sc.Cloud.HTTP.CA = "./testcert/ca.pem"
	sc.Cloud.HTTP.Key = "./testcert/client.key"
	sc.Cloud.HTTP.Cert = "./testcert/client.pem"
	sc.Cloud.HTTP.InsecureSkipVerify = true
	sc.Cloud.Report.Interval = time.Millisecond * 500

	syn, err := NewSync(config.Config{Sync: sc}, sto, nod)
	assert.NoError(t, err)
	ds, err := syn.Report(specv1.Report{})
	assert.NoError(t, err)
	expected := specv1.Desire{
		"sysapps": []interface{}{
			map[string]interface{}{
				"name":    appName,
				"version": appVer,
			},
		},
	}
	assert.Equal(t, ds, expected)
}
