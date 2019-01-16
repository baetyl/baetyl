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

	/* FIXME ota is break
	a, err := New("etc/openedge/openedge.yml")
	assert.NoError(t, err)
	defer a.Close()
	err = a.reload("V5invalid.zip")
	assert.EqualError(t, err, "failed to load new config: Modules[0].Name: zero value")
	assert.Equal(t, "", a.conf.Version)
	assert.False(t, utils.DirExists(backupDir))
	assert.False(t, utils.FileExists(backupFile))
	err = a.reload("V1init.zip")
	assert.NoError(t, err)
	assert.Equal(t, "V1", a.conf.Version)
	assert.True(t, utils.DirExists(backupDir))
	assert.False(t, utils.FileExists(backupFile))
	err = a.reload("V2load.zip")
	assert.NoError(t, err)
	assert.Equal(t, "V2", a.conf.Version)
	assert.True(t, utils.DirExists(backupDir))
	assert.False(t, utils.FileExists(backupFile))
	err = a.reload("V3clean.zip")
	assert.NoError(t, err)
	assert.Equal(t, "V3", a.conf.Version)
	assert.True(t, utils.DirExists(backupDir))
	assert.False(t, utils.FileExists(backupFile))
	err = a.reload("V4error.zip")
	assert.EqualError(t, err, "failed to unpack new config: zip: not a valid zip file")
	assert.Equal(t, "V3", a.conf.Version)
	assert.True(t, utils.DirExists(backupDir))
	assert.False(t, utils.FileExists(backupFile))
	err = a.reload("V1init.zip")
	assert.NoError(t, err)
	assert.Equal(t, "V1", a.conf.Version)
	assert.True(t, utils.DirExists(backupDir))
	assert.False(t, utils.FileExists(backupFile))
	err = a.reload("V4error.zip")
	assert.EqualError(t, err, "failed to unpack new config: zip: not a valid zip file")
	assert.Equal(t, "V1", a.conf.Version)
	assert.True(t, utils.DirExists(backupDir))
	assert.False(t, utils.FileExists(backupFile))
	err = a.reload("V3clean.zip")
	assert.NoError(t, err)
	assert.Equal(t, "V3", a.conf.Version)
	assert.True(t, utils.DirExists(backupDir))
	assert.False(t, utils.FileExists(backupFile))
	err = a.reload("V5invalid.zip")
	assert.EqualError(t, err, "failed to load new config: Modules[0].Name: zero value")
	assert.Equal(t, "V3", a.conf.Version)
	assert.True(t, utils.DirExists(backupDir))
	assert.False(t, utils.FileExists(backupFile))
	err = a.reload("V1init.zip")
	assert.NoError(t, err)
	assert.Equal(t, "V1", a.conf.Version)
	assert.True(t, utils.DirExists(backupDir))
	assert.False(t, utils.FileExists(backupFile))
	err = a.reload("V5invalid.zip")
	assert.EqualError(t, err, "failed to load new config: Modules[0].Name: zero value")
	assert.Equal(t, "V1", a.conf.Version)
	assert.True(t, utils.DirExists(backupDir))
	assert.False(t, utils.FileExists(backupFile))
	err = a.reload("V3clean.zip")
	assert.NoError(t, err)
	assert.Equal(t, "V3", a.conf.Version)
	assert.True(t, utils.DirExists(backupDir))
	assert.False(t, utils.FileExists(backupFile))
	*/
}
