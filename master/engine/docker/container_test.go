package docker

import (
	"testing"

	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func TestInitVolumes(t *testing.T) {
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Skip("docker not installed")
	}
	defer cli.Close()
	e := &dockerEngine{
		cli:      cli,
		networks: map[string]string{},
		log:      newMockLogger(),
	}
	vs := make(map[string]baetyl.ComposeVolume)
	opts := make(map[string]string)
	labels := make(map[string]string)
	vs["baetyl1"] = baetyl.ComposeVolume{
		Driver:     "local",
		DriverOpts: opts,
		Labels:     labels,
	}
	vs["baetyl2"] = baetyl.ComposeVolume{
		Driver:     "local",
		DriverOpts: opts,
		Labels:     labels,
	}
	err = e.initVolumes(vs)
	assert.NoError(t, err)
}

func TestInitNetworks(t *testing.T) {
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Skip("docker not installed")
	}
	defer cli.Close()
	e := &dockerEngine{
		cli:      cli,
		networks: map[string]string{},
		log:      newMockLogger(),
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

func TestPullImage(t *testing.T) {
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Skip("docker not installed")
	}
	defer cli.Close()
	e := &dockerEngine{
		cli:      cli,
		networks: map[string]string{},
		log:      newMockLogger(),
	}

	image := "busybox"
	err = e.pullImage(image)
	assert.NoError(t, err)
}
