package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPartialStatsByStatus(t *testing.T) {
	status := "sd"
	p := NewPartialStatsByStatus(status)
	assert.Equal(t, p[KeyStatus], status)
}
