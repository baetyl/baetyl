package ami

import (
	"github.com/baetyl/baetyl-core/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_newKubeImpl(t *testing.T) {
	c := config.EngineConfig{}
	c.Kind = "kubernetes"
	c.Kubernetes.InCluster = true

	_, err := newKubeImpl(c)
	assert.Error(t, err)
}
