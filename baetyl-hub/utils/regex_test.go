package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexIsClientID(t *testing.T) {
	assert.True(t, IsClientID(""))
	assert.False(t, IsClientID(" "))
	assert.True(t, IsClientID("-"))
	assert.True(t, IsClientID("_"))
	assert.True(t, IsClientID(GenRandomStr(0)))
	assert.True(t, IsClientID(GenRandomStr(1)))
	assert.True(t, IsClientID(GenRandomStr(128)))
	assert.False(t, IsClientID(GenRandomStr(129)))
}
