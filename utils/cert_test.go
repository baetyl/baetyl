package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSerialNumber(t *testing.T) {
	file := "../baetyl-hub/server/testcert/testssl2.pem"
	sn, err := GetSerialNumber(file)
	assert.NoError(t, err)
	assert.Equal(t, "4447389398516293299", sn)
}
