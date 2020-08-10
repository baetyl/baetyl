package sync

import (
	"errors"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/mock/plugin"
	"github.com/baetyl/baetyl/node"
	"github.com/baetyl/baetyl/store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
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

	var sc config.SyncConfig
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)

	mockCtl := gomock.NewController(t)
	link := plugin.NewMockLink(mockCtl)
	assert.NoError(t, err)
	syn := &sync{
		link:  link,
		cfg:   sc,
		store: sto,
		nod:   nod,
		log:   log.With(log.Any("test", "sync")),
	}
	link.EXPECT().IsAsyncSupported().Return(false)
	desire := specv1.Desire{"apps": map[string]interface{}{"app1": "123"}}
	msg := &specv1.Message{Content: desire, Kind: specv1.MessageReport}
	link.EXPECT().Request(gomock.Any()).Return(msg, nil)
	err = syn.reportAndDesire()
	assert.NoError(t, err)
	no, _ := syn.nod.Get()
	assert.Equal(t, desire, no.Desire)

	link.EXPECT().IsAsyncSupported().Return(true)
	link.EXPECT().Send(gomock.Any()).Return(nil)
	err = syn.reportAndDesire()
	assert.NoError(t, err)

	link.EXPECT().IsAsyncSupported().Return(false)
	link.EXPECT().Request(gomock.Any()).Return(nil, errors.New("failed to report"))
	err = syn.reportAndDesire()
	assert.Error(t, err)

	link.EXPECT().IsAsyncSupported().Return(true)
	link.EXPECT().Send(gomock.Any()).Return(errors.New("failed to report"))
	assert.Error(t, err)
}

//func TestSyncResource(t *testing.T) {
//	f, err := ioutil.TempFile("", t.Name())
//	assert.NoError(t, err)
//	assert.NotNil(t, f)
//	fmt.Println("-->tempfile", f.Name())
//
//	sto, err := store.NewBoltHold(f.Name())
//	assert.NoError(t, err)
//	assert.NotNil(t, sto)
//
//	nod, err := node.NewNode(sto)
//	assert.NoError(t, err)
//	assert.NotNil(t, nod)
//
//	appName := "baetyl-core-923jdsn"
//	appVer := "32451"
//	bi := &specv1.Desire{"sysapps": []specv1.AppInfo{
//		{
//			Name:    appName,
//			Version: appVer,
//		},
//	}}
//	data, err := json.Marshal(bi)
//	assert.NoError(t, err)
//
//	appRes := &specv1.Application{
//		Name:      appName,
//		Namespace: "baetyl-edge",
//		Version:   appVer,
//	}
//	appData, err := json.Marshal(appRes)
//	assert.NoError(t, err)
//
//	tlssvr, err := utils.NewTLSConfigServer(utils.Certificate{CA: "./testcert/ca.pem", Key: "./testcert/server.key", Cert: "./testcert/server.pem"})
//	assert.NoError(t, err)
//	assert.NotNil(t, tlssvr)
//
//	resp := []*mock.Response{
//		mock.NewResponse(200, data),
//		mock.NewResponse(200, appData),
//		mock.NewResponse(200, data),
//	}
//	ms := mock.NewServer(tlssvr, resp...)
//	assert.NotNil(t, ms)
//	defer ms.Close()
//
//	sc := config.SyncConfig{}
//	err = utils.UnmarshalYAML(nil, &sc)
//	assert.NoError(t, err)
//	sc.Cloud.HTTP.Address = ms.URL
//	sc.Cloud.HTTP.CA = "./testcert/ca.pem"
//	sc.Cloud.HTTP.Key = "./testcert/client.key"
//	sc.Cloud.HTTP.Cert = "./testcert/client.pem"
//	sc.Cloud.HTTP.InsecureSkipVerify = true
//	sc.Cloud.Report.Interval = time.Millisecond * 500
//
//	syn, err := NewSync(config.Config{Sync: sc}, sto, nod)
//	assert.NoError(t, err)
//	ds, err := syn.Report(specv1.Report{})
//	assert.NoError(t, err)
//	expected := specv1.Desire{
//		"sysapps": []interface{}{
//			map[string]interface{}{
//				"name":    appName,
//				"version": appVer,
//			},
//		},
//	}
//	assert.Equal(t, ds, expected)
//}
