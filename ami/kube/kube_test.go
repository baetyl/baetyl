package kube

import (
	"github.com/baetyl/baetyl/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_newKubeImpl(t *testing.T) {
	c := config.AmiConfig{}
	c.Kind = "kubernetes"
	c.Kubernetes.InCluster = true

	_, err := newKubeImpl(c)
	assert.Error(t, err)
}
