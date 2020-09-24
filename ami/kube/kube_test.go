package kube

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/v2/config"
)

func Test_newKubeImpl(t *testing.T) {
	c := config.AmiConfig{}
	c.Kube.OutCluster = false

	_, err := newKubeImpl(c)
	assert.Error(t, err)
}
