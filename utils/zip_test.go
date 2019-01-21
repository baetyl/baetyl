package utils

import (
	"os"
	"testing"

	"github.com/mholt/archiver"
	"github.com/stretchr/testify/assert"
)

func TestZip(t *testing.T) {
	os.MkdirAll("var", 0777)
	defer os.RemoveAll("var")
	err := archiver.Zip.Make("var/tmp.zip", []string{"tomb.go"})
	assert.NoError(t, err)
	err = archiver.Zip.Open("var/tmp.zip", "var/code")
	assert.NoError(t, err)
}
