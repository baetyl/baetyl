// +build !windows

package master

import (
	"os"
	"testing"

	"github.com/mholt/archiver"
	"github.com/stretchr/testify/assert"
)

func TestPrepare(t *testing.T) {
	t.Skip("Prepare test data")
	err := os.Chdir("testrun")
	assert.NoError(t, err)
	defer os.RemoveAll("var")
	err = archiver.Zip.Make("V1init.zip", []string{"V1init/var"})
	assert.NoError(t, err)
	err = archiver.Zip.Make("V2load.zip", []string{"V2load/var"})
	assert.NoError(t, err)
	err = archiver.Zip.Make("V3clean.zip", []string{"V3clean/var"})
	assert.NoError(t, err)
	err = archiver.Zip.Make("V5invalid.zip", []string{"V5invalid/var"})
	assert.NoError(t, err)
}

func TestMasterStart(t *testing.T) {
	err := os.Chdir("testrun")
	assert.NoError(t, err)
	defer os.RemoveAll("var")

	pwd, err := os.Getwd()
	assert.NoError(t, err)
	a, err := New(pwd, "etc/openedge/openedge.yml")
	assert.NoError(t, err)
	defer a.Close()
	err = a.Start()
	assert.NoError(t, err)
	err = a.reload("V1init")
	assert.NoError(t, err)
	assert.True(t, dirExists(backupDir))
	assert.False(t, fileExists(backupFile))
	err = a.reload("V2load")
	assert.NoError(t, err)
	assert.True(t, dirExists(backupDir))
	assert.False(t, fileExists(backupFile))
	err = a.reload("V3clean")
	assert.NoError(t, err)
	assert.True(t, dirExists(backupDir))
	assert.False(t, fileExists(backupFile))
	err = a.reload("V4error")
	assert.EqualError(t, err, "failed to unpack new config: zip: not a valid zip file")
	assert.True(t, dirExists(backupDir))
	assert.False(t, fileExists(backupFile))
	err = a.reload("V1init")
	assert.NoError(t, err)
	assert.True(t, dirExists(backupDir))
	assert.False(t, fileExists(backupFile))
	err = a.reload("V4error")
	assert.EqualError(t, err, "failed to unpack new config: zip: not a valid zip file")
	assert.True(t, dirExists(backupDir))
	assert.False(t, fileExists(backupFile))
	err = a.reload("V3clean")
	assert.NoError(t, err)
	assert.True(t, dirExists(backupDir))
	assert.False(t, fileExists(backupFile))
	err = a.reload("V5invalid")
	assert.EqualError(t, err, "failed to load new config: Modules[0].Name: zero value")
	assert.True(t, dirExists(backupDir))
	assert.False(t, fileExists(backupFile))
	err = a.reload("V1init")
	assert.NoError(t, err)
	assert.True(t, dirExists(backupDir))
	assert.False(t, fileExists(backupFile))
	err = a.reload("V5invalid")
	assert.EqualError(t, err, "failed to load new config: Modules[0].Name: zero value")
	assert.True(t, dirExists(backupDir))
	assert.False(t, fileExists(backupFile))
	err = a.reload("V3clean")
	assert.NoError(t, err)
	assert.True(t, dirExists(backupDir))
	assert.False(t, fileExists(backupFile))
}
