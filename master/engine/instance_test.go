package engine

import (
	"testing"

	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	static := []string{"name=baetyl", "org=linux", baetyl.EnvKeyServiceToken + "=key", baetyl.EnvServiceTokenKey + "=key"}
	dyn := map[string]string{
		"repo":    "github",
		"project": "baetyl",
	}
	envs := GenerateInstanceEnv(t.Name(), static, dyn)
	assert.Contains(t, envs, "name=baetyl")
	assert.Contains(t, envs, "org=linux")
	assert.Contains(t, envs, "repo=github")
	assert.Contains(t, envs, "project=baetyl")
	assert.Contains(t, envs, baetyl.EnvKeyServiceInstanceName+"="+t.Name())
	assert.NotContains(t, envs, baetyl.EnvKeyServiceToken+"=key")
	assert.NotContains(t, envs, baetyl.EnvServiceTokenKey+"=key")
}
