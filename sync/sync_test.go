package sync

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/shadow"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-go/faas"
	"github.com/baetyl/baetyl-go/mock"
	"github.com/baetyl/baetyl-go/spec"
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

	sha, err := shadow.NewShadow(t.Name(), t.Name(), sto)
	assert.NoError(t, err)
	assert.NotNil(t, sha)

	bi := &spec.Delta{"apps": map[string]interface{}{"app1": "123"}}
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
	sc.Cloud.HTTP.Key = "./testcert/server.key"
	sc.Cloud.HTTP.Cert = "./testcert/server.pem"
	sc.Cloud.HTTP.InsecureSkipVerify = true

	syn, err := NewSync(sc, sto, sha, nil)
	assert.NoError(t, err)

	err = syn.Report(faas.Message{})
	assert.NoError(t, err)

	sp, err := sha.Get()
	assert.NoError(t, err)
	assert.Equal(t, spec.Desire{"apps": map[string]interface{}{"app1": "123"}}, sp.Desire)
}
