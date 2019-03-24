// +build !windows

package master

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/master/engine"
	_ "github.com/baidu/openedge/master/engine/native"
	"github.com/baidu/openedge/utils"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/assert"
)

func TestUpdateSystem(t *testing.T) {
	err := os.Chdir("testdata")
	assert.NoError(t, err)
	dir := path.Join("var", "db", "openedge", "app")
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

	err = m.UpdateSystem(path.Join(dir, "v4"), false)
	assert.EqualError(t, err, "failed to update system: open var/db/openedge/app/v4/application.yml: no such file or directory")
	assert.Equal(t, "", m.appcfg.Version)
	assert.False(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	msg, ok := m.context.Get("error")
	assert.True(t, ok)
	assert.Equal(t, "failed to update system: open var/db/openedge/app/v4/application.yml: no such file or directory", msg)

	err = m.UpdateSystem(path.Join(dir, "v5"), false)
	assert.EqualError(t, err, "failed to update system: volume 'cmd-bin' not found")
	assert.Equal(t, "", m.appcfg.Version)
	assert.False(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	msg, ok = m.context.Get("error")
	assert.True(t, ok)
	assert.Equal(t, "failed to update system: volume 'cmd-bin' not found", msg)

	err = m.UpdateSystem(path.Join(dir, "v6"), false)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "wait_exit_5/lib/openedge/cmd-nonexist/package.yml: no such file or directory")
	assert.Equal(t, "", m.appcfg.Version)
	assert.False(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	msg, ok = m.context.Get("error")
	assert.True(t, ok)
	assert.Contains(t, msg, "wait_exit_5/lib/openedge/cmd-nonexist/package.yml: no such file or directory")

	err = m.UpdateSystem(path.Join(dir, "v1"), false)
	assert.NoError(t, err)
	assert.Equal(t, "v1", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	_, ok = m.context.Get("error")
	assert.False(t, ok)

	err = m.UpdateSystem(path.Join(dir, "v4"), false)
	assert.EqualError(t, err, "failed to update system: open var/db/openedge/app/v4/application.yml: no such file or directory")
	assert.Equal(t, "v1", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	msg, ok = m.context.Get("error")
	assert.True(t, ok)
	assert.Equal(t, "failed to update system: open var/db/openedge/app/v4/application.yml: no such file or directory", msg)

	err = m.UpdateSystem(path.Join(dir, "v2"), false)
	assert.Equal(t, "v2", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	_, ok = m.context.Get("error")
	assert.False(t, ok)

	err = m.UpdateSystem(path.Join(dir, "v5"), false)
	assert.EqualError(t, err, "failed to update system: volume 'cmd-bin' not found")
	assert.Equal(t, "v2", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	msg, ok = m.context.Get("error")
	assert.True(t, ok)
	assert.Equal(t, "failed to update system: volume 'cmd-bin' not found", msg)

	err = m.UpdateSystem(path.Join(dir, "v3"), false)
	assert.Equal(t, "v3", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	_, ok = m.context.Get("error")
	assert.False(t, ok)

	err = m.UpdateSystem(path.Join(dir, "v6"), false)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "wait_exit_5/lib/openedge/cmd-nonexist/package.yml: no such file or directory")
	assert.Equal(t, "v3", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	msg, ok = m.context.Get("error")
	assert.True(t, ok)
	assert.Contains(t, msg, "wait_exit_5/lib/openedge/cmd-nonexist/package.yml: no such file or directory")

	err = m.UpdateSystem(path.Join(dir, "v2"), false)
	assert.Equal(t, "v2", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	_, ok = m.context.Get("error")
	assert.False(t, ok)

	dv := path.Join("var", "db", "openedge", "dummy")
	err = os.MkdirAll(dv, 0755)
	assert.NoError(t, err)
	f7 := path.Join("var", "db", "openedge", "app", "v7", "application.yml")
	err = os.MkdirAll(path.Dir(f7), 0755)
	assert.NoError(t, err)
	err = ioutil.WriteFile(f7, []byte(`
version: v7
volumes:
  - name: cmd-bin
    path: var/db/openedge/cmd
  - name: cmd-bin
    path: var/db/openedge/cmd
  - name: dummy
    path: var/db/openedge/dummy
`), 0755)
	assert.NoError(t, err)

	err = m.UpdateSystem(path.Join(dir, "v7"), true)
	assert.Equal(t, "v7", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	_, ok = m.context.Get("error")
	assert.False(t, ok)
	assert.True(t, utils.DirExists(dv))
	assert.False(t, utils.FileExists(f7))

	f8 := path.Join("var", "db", "openedge", "app", "v8", "application.yml")
	err = os.MkdirAll(path.Dir(f8), 0755)
	assert.NoError(t, err)
	err = ioutil.WriteFile(f8, []byte(`
version: v8
volumes:
  - name: cmd-bin
    path: var/db/openedge/cmd
`), 0755)
	assert.NoError(t, err)

	err = m.UpdateSystem(path.Join(dir, "v8"), true)
	assert.NoError(t, err)
	assert.Equal(t, "v8", m.appcfg.Version)
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	_, ok = m.context.Get("error")
	assert.False(t, ok)
	assert.False(t, utils.DirExists(dv))
	assert.False(t, utils.FileExists(f8))
}
