package main

import (
	"testing"

	"github.com/baetyl/baetyl/utils"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	var c Config
	err := utils.LoadYAML("./example/etc/baetyl/service.yml", &c)
	assert.NoError(t, err)

	assert.Len(t, c.Clients, 1)
	assert.Equal(t, "baidu-bos", c.Clients[0].Name)
	assert.Equal(t, "bos.gz.qasandbox.bcetest.baidu.com", c.Clients[0].Address)

	assert.Len(t, c.Rules, 1)
	assert.Equal(t, "remote-write-bos", c.Rules[0].ClientID)
	assert.Equal(t, "t", c.Rules[0].Subscribe.Topic)
}
