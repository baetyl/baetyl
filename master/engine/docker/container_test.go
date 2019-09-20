package docker

import (
	"testing"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func Test_dockerEngine_initNetworks(t *testing.T) {
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Skip("docker not installed")
	}
	e := &dockerEngine{
		cli:      cli,
		networks: map[string]string{},
		log:      logger.WithField("test", "test"),
	}

	err = e.initNetworks(nil)
	assert.NoError(t, err)
	err = e.initNetworks(ComposeNetworks{})
	assert.NoError(t, err)
	assert.Len(t, e.networks, 1)
	nw := baetyl.ComposeNetwork{}
	err = e.initNetworks(ComposeNetworks{"baetyl": nw})
	assert.NoError(t, err)
	assert.Len(t, e.networks, 1)
	utils.SetDefaults(&nw)
	err = e.initNetworks(ComposeNetworks{"baetyl": nw})
	assert.NoError(t, err)
	assert.Len(t, e.networks, 1)
	nw.Labels["dummy"] = "dummy"
	err = e.initNetworks(ComposeNetworks{"dummy": nw})
	assert.NoError(t, err)
	assert.Len(t, e.networks, 2)
	err = e.initNetworks(ComposeNetworks{"dummy": nw})
	assert.NoError(t, err)
	assert.Len(t, e.networks, 2)
}
