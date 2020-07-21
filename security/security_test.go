package security

import (
	"testing"

	"github.com/baetyl/baetyl/config"
	"github.com/stretchr/testify/assert"
)

func TestNewSecurity(t *testing.T) {
	// good case
	cfg := config.SecurityConfig{
		Kind: "pki",
	}
	cfg.PKIConfig = genPKIConf(t)
	bhSto := genBolthold(t)
	sec, err := NewSecurity(cfg, bhSto)
	assert.NoError(t, err)
	assert.NotNil(t, sec)
}
