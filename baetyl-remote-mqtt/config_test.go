package main

import (
	"testing"

	"github.com/baetyl/baetyl/utils"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	var c Config
	err := utils.LoadYAML("../example/native/var/db/baetyl/remote-iothub-conf/service.yml", &c)
	assert.NoError(t, err)

	assert.Len(t, c.Remotes, 1)
	assert.Equal(t, "iothub", c.Remotes[0].Name)
	assert.Equal(t, "11dd7422353c46fc8851ef8fb7114533", c.Remotes[0].ClientID)

	assert.Len(t, c.Rules, 1)
	assert.Equal(t, "", c.Rules[0].Hub.ClientID)
	assert.Len(t, c.Rules[0].Hub.Subscriptions, 1)
	assert.Equal(t, "t", c.Rules[0].Hub.Subscriptions[0].Topic)
	assert.Equal(t, uint32(0), c.Rules[0].Hub.Subscriptions[0].QOS)
	assert.Equal(t, "iothub", c.Rules[0].Remote.Name)
	assert.Equal(t, "", c.Rules[0].Remote.ClientID)
	assert.Len(t, c.Rules[0].Remote.Subscriptions, 1)
	assert.Equal(t, "t/remote", c.Rules[0].Remote.Subscriptions[0].Topic)
	assert.Equal(t, uint32(1), c.Rules[0].Remote.Subscriptions[0].QOS)
}
