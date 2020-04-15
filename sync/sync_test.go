package sync

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-go/mock"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl-go/utils"
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

	bi := &v1.Desire{"apps": map[string]interface{}{"app1": "123"}}
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
	sc.Cloud.HTTP.Address = ms.URL
	sc.Cloud.HTTP.CA = "./testcert/ca.pem"
	sc.Cloud.HTTP.Key = "./testcert/client.key"
	sc.Cloud.HTTP.Cert = "./testcert/client.pem"
	sc.Cloud.HTTP.InsecureSkipVerify = true
	sc.Cloud.Report.Interval = time.Millisecond * 500

	syn, err := NewSync(sc, sto, nod)
	assert.NoError(t, err)
	syn.Start()

	desire := <-syn.fifo
	assert.Equal(t, v1.Desire{"apps": map[string]interface{}{"app1": "123"}}, desire)

	sc = config.SyncConfig{}
	_, err = NewSync(sc, sto, nod)
	assert.Error(t, err, ErrSyncTLSConfigMissing.Error())

	sc.Cloud.HTTP.Cert = "./testcert/notexist.pem"
	_, err = NewSync(sc, sto, nod)
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
	syn, err = NewSync(sc, sto, nod)
	assert.NoError(t, err)
	syn.Start()
	err = syn.reportAndDesireAsync()
	assert.Error(t, err)

	ms = mock.NewServer(tlssvr, mock.NewResponse(500, []byte{}))
	sc = config.SyncConfig{}
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)
	sc.Cloud.HTTP.Address = ms.URL
	sc.Cloud.HTTP.CA = "./testcert/ca.pem"
	sc.Cloud.HTTP.Key = "./testcert/client.key"
	sc.Cloud.HTTP.Cert = "./testcert/client.pem"
	sc.Cloud.HTTP.InsecureSkipVerify = true
	syn, err = NewSync(sc, sto, nod)
	assert.NoError(t, err)
	syn.Start()
	err = syn.reportAndDesireAsync()
	assert.Error(t, err)
	syn.Close()
}

func TestSync_ReportAndDesire(t *testing.T) {
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

	bi := &v1.Desire{"apps": map[string]interface{}{"app1": "123"}}
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
	sc.Cloud.HTTP.Address = ms.URL
	sc.Cloud.HTTP.CA = "./testcert/ca.pem"
	sc.Cloud.HTTP.Key = "./testcert/client.key"
	sc.Cloud.HTTP.Cert = "./testcert/client.pem"
	sc.Cloud.HTTP.InsecureSkipVerify = true
	sc.Cloud.Report.Interval = time.Millisecond * 500

	syn, err := NewSync(sc, sto, nod)
	assert.NoError(t, err)
	err = syn.ReportAndDesire()
	assert.NoError(t, err)
	err = syn.ReportAndDesire()
	assert.NotNil(t, err)
}
