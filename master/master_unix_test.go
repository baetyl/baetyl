// +build !windows

package master

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMasterStart(t *testing.T) {
	err := os.Chdir("testrun")
	assert.NoError(t, err)
	defer os.RemoveAll(appDir)

	a, err := New("conf/conf.yml")
	assert.NoError(t, err)
	defer a.Close()
	err = a.Start()
	assert.NoError(t, err)
	err = a.reload("V1init")
	assert.NoError(t, err)
	assert.True(t, dirExists(appDir))
	assert.False(t, fileExists(appBackupFile))
	err = a.reload("V2load")
	assert.NoError(t, err)
	assert.True(t, dirExists(appDir))
	assert.False(t, fileExists(appBackupFile))
	err = a.reload("V3clean")
	assert.NoError(t, err)
	assert.True(t, dirExists(appDir))
	assert.False(t, fileExists(appBackupFile))
	err = a.reload("V4error")
	assert.EqualError(t, err, "failed to unpack new config: zip: not a valid zip file")
	assert.True(t, dirExists(appDir))
	assert.False(t, fileExists(appBackupFile))
	err = a.reload("V1init")
	assert.NoError(t, err)
	assert.True(t, dirExists(appDir))
	assert.False(t, fileExists(appBackupFile))
	err = a.reload("V4error")
	assert.EqualError(t, err, "failed to unpack new config: zip: not a valid zip file")
	assert.True(t, dirExists(appDir))
	assert.False(t, fileExists(appBackupFile))
	err = a.reload("V3clean")
	assert.NoError(t, err)
	assert.True(t, dirExists(appDir))
	assert.False(t, fileExists(appBackupFile))
	err = a.reload("V5invalid")
	assert.EqualError(t, err, "failed to load new config: Modules[0].Config.Name: zero value")
	assert.True(t, dirExists(appDir))
	assert.False(t, fileExists(appBackupFile))
	err = a.reload("V1init")
	assert.NoError(t, err)
	assert.True(t, dirExists(appDir))
	assert.False(t, fileExists(appBackupFile))
	err = a.reload("V5invalid")
	assert.EqualError(t, err, "failed to load new config: Modules[0].Config.Name: zero value")
	assert.True(t, dirExists(appDir))
	assert.False(t, fileExists(appBackupFile))
	err = a.reload("V3clean")
	assert.NoError(t, err)
	assert.True(t, dirExists(appDir))
	assert.False(t, fileExists(appBackupFile))
}
