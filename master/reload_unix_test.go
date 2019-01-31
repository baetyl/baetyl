// +build !windows

package master

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	_ "github.com/baidu/openedge/master/engine/native"
	"github.com/baidu/openedge/utils"
	"github.com/mholt/archiver"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/assert"
)

func TestPrepare(t *testing.T) {
	t.Skip("Prepare test data once")
	err := os.Chdir("testdata")
	assert.NoError(t, err)
	defer os.RemoveAll("var")
	err = archiver.Zip.Make("V1init.zip", []string{"V1init/var"})
	assert.NoError(t, err)
	err = archiver.Zip.Make("V2load.zip", []string{"V2load/var"})
	assert.NoError(t, err)
	err = archiver.Zip.Make("V3clean.zip", []string{"V3clean/var"})
	assert.NoError(t, err)
	err = ioutil.WriteFile("V4error.zip", []byte{'t'}, os.ModePerm)
	assert.NoError(t, err)
	err = archiver.Zip.Make("V5invalid.zip", []string{"V5invalid/var"})
	assert.NoError(t, err)
}

func TestReload(t *testing.T) {
	err := os.Chdir("testdata")
	assert.NoError(t, err)
	defer os.RemoveAll("var")

	pwd, err := os.Getwd()
	assert.NoError(t, err)
	os.Chmod("lib/openedge/packages/cmd/cmd.py", os.ModePerm)

	m := &Master{
		pwd:     pwd,
		services: cmap.New(),
		log:      logger.WithField("openedge", "master"),
	}
	m.engine, err = engine.New("native", m.pwd)
	assert.NoError(t, err)
	defer m.Close()
	err = m.reload("V5invalid.zip")
	assert.EqualError(t, err, "failed to load new service config: Services[0].Image: zero value")
	assert.Equal(t, "", m.dyncfg.Version)
	assert.False(t, utils.DirExists(serviceOldDir))
	assert.False(t, utils.FileExists(configFile))
	err = m.reload("V1init.zip")
	assert.NoError(t, err)
	assert.Equal(t, "V1", m.dyncfg.Version)
	assert.False(t, utils.DirExists(serviceOldDir))
	assert.True(t, utils.FileExists(configFile))
	err = m.reload("V2load.zip")
	assert.NoError(t, err)
	assert.Equal(t, "V2", m.dyncfg.Version)
	assert.False(t, utils.DirExists(serviceOldDir))
	assert.True(t, utils.FileExists(configFile))
	err = m.reload("V3clean.zip")
	assert.NoError(t, err)
	assert.Equal(t, "V3", m.dyncfg.Version)
	assert.False(t, utils.DirExists(serviceOldDir))
	assert.True(t, utils.FileExists(configFile))
	err = m.reload("V4error.zip")
	assert.EqualError(t, err, "failed to unpack new service config: zip: not a valid zip file")
	assert.Equal(t, "V3", m.dyncfg.Version)
	assert.False(t, utils.DirExists(serviceOldDir))
	assert.True(t, utils.FileExists(configFile))
	err = m.reload("Nonexist.zip")
	assert.EqualError(t, err, "no file: Nonexist.zip")
	assert.Equal(t, "V3", m.dyncfg.Version)
	assert.False(t, utils.DirExists(serviceOldDir))
	assert.True(t, utils.FileExists(configFile))
	err = m.reload("V1init.zip")
	assert.NoError(t, err)
	assert.Equal(t, "V1", m.dyncfg.Version)
	assert.False(t, utils.DirExists(serviceOldDir))
	assert.True(t, utils.FileExists(configFile))
	err = m.reload("V4error.zip")
	assert.EqualError(t, err, "failed to unpack new service config: zip: not a valid zip file")
	assert.Equal(t, "V1", m.dyncfg.Version)
	assert.False(t, utils.DirExists(serviceOldDir))
	assert.True(t, utils.FileExists(configFile))
	err = m.reload("V3clean.zip")
	assert.NoError(t, err)
	assert.Equal(t, "V3", m.dyncfg.Version)
	assert.False(t, utils.DirExists(serviceOldDir))
	assert.True(t, utils.FileExists(configFile))
	err = m.reload("V5invalid.zip")
	assert.EqualError(t, err, "failed to load new service config: Services[0].Image: zero value")
	assert.Equal(t, "V3", m.dyncfg.Version)
	assert.False(t, utils.DirExists(serviceOldDir))
	assert.True(t, utils.FileExists(configFile))
	err = m.reload("V1init.zip")
	assert.NoError(t, err)
	assert.Equal(t, "V1", m.dyncfg.Version)
	assert.False(t, utils.DirExists(serviceOldDir))
	assert.True(t, utils.FileExists(configFile))
	err = m.reload("V5invalid.zip")
	assert.EqualError(t, err, "failed to load new service config: Services[0].Image: zero value")
	assert.Equal(t, "V1", m.dyncfg.Version)
	assert.False(t, utils.DirExists(serviceOldDir))
	assert.True(t, utils.FileExists(configFile))
	err = m.reload("V3clean.zip")
	assert.NoError(t, err)
	assert.Equal(t, "V3", m.dyncfg.Version)
	assert.False(t, utils.DirExists(serviceOldDir))
	assert.True(t, utils.FileExists(configFile))
}
