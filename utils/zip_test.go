package utils

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZip(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "example")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	tmpfile, err := ioutil.TempFile(tmpdir, "test")
	assert.NoError(t, err)
	zipPath := tmpfile.Name() + ".zip"
	err = Zip([]string{tmpfile.Name()}, zipPath)
	assert.NoError(t, err)
	err = Unzip(zipPath, tmpdir)
	assert.NoError(t, err)
	err = Unzip(zipPath, tmpdir)
	assert.NoError(t, err)
}
