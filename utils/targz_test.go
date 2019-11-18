package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTar(t *testing.T) {
	os.MkdirAll("var", 0755)
	defer os.RemoveAll("var")
	err := TarGz([]string{"tomb.go"}, "var/tmp.tar.gz")
	assert.NoError(t, err)
	err = UntarGz("var/tmp.tar.gz", "var/code")
	assert.NoError(t, err)
	err = UntarGz("var/tmp.tar.gz", "var/code")
	assert.NoError(t, err)
}
