package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZip(t *testing.T) {
	os.MkdirAll("var", 0755)
	defer os.RemoveAll("var")
	err := Zip([]string{"tomb.go"}, "var/tmp.zip")
	assert.NoError(t, err)
	err = Unzip("var/tmp.zip", "var/code")
	assert.NoError(t, err)
	err = Unzip("var/tmp.zip", "var/code")
	assert.NoError(t, err)
}
