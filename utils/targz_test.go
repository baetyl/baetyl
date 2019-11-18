package utils

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTar(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "example")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpdir)
	tmpfile, err := ioutil.TempFile(tmpdir, "test")
	assert.NoError(t, err)
	tgzPath := tmpfile.Name() + ".tar.gz"
	err = TarGz([]string{tmpfile.Name()}, tgzPath)
	assert.NoError(t, err)
	err = UntarGz(tgzPath, tmpdir)
	assert.NoError(t, err)
	err = UntarGz(tgzPath, tmpdir)
	assert.NoError(t, err)
}
