package baetyl

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEnvClient(t *testing.T) {
	got, err := NewEnvClient()
	assert.EqualError(t, err, "Env (BAETYL_MASTER_API_ADDRESS) not found")
	assert.Nil(t, got)

	// old
	os.Setenv(EnvMasterAPIKey, "0.0.0.0")
	os.Setenv(EnvMasterAPIVersionKey, "v0")
	got, err = NewEnvClient()
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "/v0", got.ver)

	// new
	os.Setenv(EnvKeyMasterAPIAddress, "0.0.0.1")
	os.Setenv(EnvKeyMasterAPIVersion, "v1")
	got, err = NewEnvClient()
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "/v1", got.ver)
}
