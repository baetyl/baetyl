package docker

import (
	"testing"

	"github.com/baetyl/baetyl/master/engine"
	"github.com/stretchr/testify/assert"
)

func Test_toPartialStats(t *testing.T) {
	var attr attribute
	id := "test"
	name := "baetyl-test"
	attr.Name = name
	attr.Container.ID = id
	attr.Container.Name = name
	stats := attr.toPartialStats()
	assert.Equal(t, name, stats[engine.KeyName])
	assert.Equal(t, attr.Container, stats["container"])
}
