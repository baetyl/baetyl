package utils

import (
	"os"
	"testing"

	"github.com/mholt/archiver"
	"github.com/stretchr/testify/assert"
)

func TestZip(t *testing.T) {
	os.MkdirAll("var", 0755)
	defer os.RemoveAll("var")
	err := archiver.DefaultZip.Archive([]string{"tomb.go"}, "var/tmp.zip")
	assert.NoError(t, err)
	err = archiver.DefaultZip.Unarchive("var/tmp.zip", "var/code")
	assert.NoError(t, err)
}
