// +build !windows

package master

import (
	"os"
	"testing"
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	_ "github.com/baidu/openedge/master/engine/native"
	"github.com/baidu/openedge/utils"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/assert"
)

var (
	v1 = []byte(`
version: V1
services:
  - name: wait_exit_1
    image: cmd
    mounts:
      - name: cmd-bin
        path: lib/openedge/cmd
  - name: wait_exit_2
    image: cmd
    mounts:
      - name: cmd-bin
        path: lib/openedge/cmd
  - name: hi
    image: cmd
    mounts:
      - name: cmd-bin
        path: lib/openedge/cmd
volumes:
  - name: cmd-bin
    path: var/db/openedge/cmd
`)
	v2 = []byte(`
version: V2
services:
  - name: wait_exit_2
    image: hub.baidubce.com/openedge/cmd:0.1.2
    mounts:
      - name: cmd-bin
        path: lib/openedge/hub.baidubce.com/openedge/cmd:0.1.2
  - name: hi
    image: cmd
    mounts:
      - name: cmd-bin
        path: lib/openedge/cmd
  - name: wait_exit_4
    image: cmd
    mounts:
      - name: cmd-bin
        path: lib/openedge/cmd
volumes:
  - name: cmd-bin
    path: var/db/openedge/cmd
`)
	v3 = []byte(`
version: V3
services: []
volumes: []
`)
	v4 []byte
	v5 = []byte(`
version: V5
services:
  - name: wait_exit_5
    image: cmd
    mounts:
      - name: cmd-bin
        path: lib/openedge/cmd
`)
	v6 = []byte(`
version: V6
services:
  - name: wait_exit_5
    image: cmd-nonexist
    mounts:
      - name: cmd-bin
        path: lib/openedge/cmd
volumes:
  - name: cmd-bin
    path: var/db/openedge/cmd
`)
)

func TestUpdateSystem(t *testing.T) {
	err := os.Chdir("testdata")
	assert.NoError(t, err)
	defer os.RemoveAll(appConfigFile)
	defer os.RemoveAll(appBackupFile)
	defer os.RemoveAll("var/run")

	pwd, err := os.Getwd()
	assert.NoError(t, err)

	m := &Master{
		pwd:      pwd,
		accounts: cmap.New(),
		services: cmap.New(),
		context:  cmap.New(),
		log:      logger.WithField("openedge", "master"),
	}
	m.engine, err = engine.New("native", time.Second, m.pwd)
	assert.NoError(t, err)
	defer m.Close()

	err = m.UpdateSystem(v4)
	assert.EqualError(t, err, "failed to update system: application config is null")
	assert.Equal(t, "", m.appcfg.Version)
	assert.False(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	msg, ok := m.context.Get("error")
	assert.True(t, ok)
	assert.Equal(t, "failed to update system: application config is null", msg)

	err = m.UpdateSystem(v5)
	assert.EqualError(t, err, "failed to update system: volume 'cmd-bin' not found")
	assert.Equal(t, "", m.appcfg.Version)
	assert.False(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	msg, ok = m.context.Get("error")
	assert.True(t, ok)
	assert.Equal(t, "failed to update system: volume 'cmd-bin' not found", msg)

	err = m.UpdateSystem(v6)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "wait_exit_5/lib/openedge/cmd-nonexist/package.yml: no such file or directory")
	assert.Equal(t, "", m.appcfg.Version)
	assert.False(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	msg, ok = m.context.Get("error")
	assert.True(t, ok)
	assert.Contains(t, msg, "wait_exit_5/lib/openedge/cmd-nonexist/package.yml: no such file or directory")

	err = m.UpdateSystem(v1)
	assert.NoError(t, err)
	assert.Equal(t, "V1", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	_, ok = m.context.Get("error")
	assert.False(t, ok)

	err = m.UpdateSystem(v4)
	assert.EqualError(t, err, "failed to update system: application config is null")
	assert.Equal(t, "V1", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	msg, ok = m.context.Get("error")
	assert.True(t, ok)
	assert.Equal(t, "failed to update system: application config is null", msg)

	err = m.UpdateSystem(v2)
	assert.Equal(t, "V2", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	_, ok = m.context.Get("error")
	assert.False(t, ok)

	err = m.UpdateSystem(v5)
	assert.EqualError(t, err, "failed to update system: volume 'cmd-bin' not found")
	assert.Equal(t, "V2", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	msg, ok = m.context.Get("error")
	assert.True(t, ok)
	assert.Equal(t, "failed to update system: volume 'cmd-bin' not found", msg)

	err = m.UpdateSystem(v3)
	assert.Equal(t, "V3", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	_, ok = m.context.Get("error")
	assert.False(t, ok)

	err = m.UpdateSystem(v6)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "wait_exit_5/lib/openedge/cmd-nonexist/package.yml: no such file or directory")
	assert.Equal(t, "V3", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	msg, ok = m.context.Get("error")
	assert.True(t, ok)
	assert.Contains(t, msg, "wait_exit_5/lib/openedge/cmd-nonexist/package.yml: no such file or directory")

	err = m.UpdateSystem(v2)
	assert.Equal(t, "V2", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	_, ok = m.context.Get("error")
	assert.False(t, ok)

	err = m.UpdateSystem(v1)
	assert.NoError(t, err)
	assert.Equal(t, "V1", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	_, ok = m.context.Get("error")
	assert.False(t, ok)
}
