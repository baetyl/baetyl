package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTar(t *testing.T) {
	os.MkdirAll("var", 0755)
	defer os.RemoveAll("var")
	err := Tar([]string{"tomb.go"}, "var/tmp.tar")
	assert.NoError(t, err)
	err = Untar("var/tmp.tar", "var/code")
	assert.NoError(t, err)
	err = Untar("var/tmp.tar", "var/code")
	assert.NoError(t, err)
}
