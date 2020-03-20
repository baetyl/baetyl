package initialize

import (
	"encoding/json"
	"testing"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-go/mock"
	"github.com/baetyl/baetyl-go/spec/api"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/stretchr/testify/assert"
)

func TestInitialize_Activate(t *testing.T) {
	resp := &api.ActiveResponse{
		NodeName:  "node.test",
		Namespace: "default",
		Certificate: utils.Certificate{
			CA:                 "ca info",
			Key:                "key info",
			Cert:               "cert info",
			Name:               "name info",
			InsecureSkipVerify: false,
		},
	}
	data, err := json.Marshal(resp)
	assert.NoError(t, err)

	ms := mock.NewServer(nil, mock.NewResponse(200, data))
	assert.NotNil(t, ms)
	defer ms.Close()

	ic := &config.InitConfig{}
	err = utils.UnmarshalYAML(nil, ic)
	assert.NoError(t, err)
	ic.Batch.Name = "batch.test"
	ic.Batch.Namespace = "default"
	ic.Batch.SecurityType = "Token"
	ic.Batch.SecurityKey = "123456"
	ic.Cloud.HTTP.Address = ms.URL
	ic.ActivateConfig.Attributes = []config.Attribute{
		{
			Name:  "abc",
			Value: "abc",
		},
	}
	ic.ActivateConfig.Fingerprints = []config.Fingerprint{
		{
			Proof: config.ProofInput,
			Value: "abc",
		},
	}
	c := &config.Config{}
	c.Init = *ic

	// good case
	init, err := NewInit(c)
	assert.Nil(t, err)
	init.WaitAndClose()
	assert.Equal(t, resp.NodeName, c.Sync.Node.Name)
	assert.Equal(t, resp.Namespace, c.Sync.Node.Namespace)
	assert.Equal(t, resp.Certificate.Cert, c.Sync.Cloud.HTTP.Cert)
	assert.Equal(t, resp.Certificate.Name, c.Sync.Cloud.HTTP.Name)
	assert.Equal(t, resp.Certificate.CA, c.Sync.Cloud.HTTP.CA)
	assert.Equal(t, resp.Certificate.Key, c.Sync.Cloud.HTTP.Key)
	assert.Equal(t, resp.Certificate.InsecureSkipVerify, c.Sync.Cloud.HTTP.InsecureSkipVerify)
}
