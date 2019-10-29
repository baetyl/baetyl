package master

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path"
	"testing"
	"time"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/master/engine"
	_ "github.com/baetyl/baetyl/master/engine/native"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/assert"
)

func TestUpdateAPP(t *testing.T) {
	oldpwd, err := os.Getwd()
	assert.NoError(t, err)
	err = os.Chdir("testdata")
	assert.NoError(t, err)
	defer os.Chdir(oldpwd)

	os.RemoveAll(appConfigFile)
	os.RemoveAll(appBackupFile)
	os.RemoveAll(appBackupFile)
	os.RemoveAll("var/run")
	defer os.RemoveAll(appConfigFile)
	defer os.RemoveAll(appBackupFile)
	defer os.RemoveAll("var/run")

	pwd, err := os.Getwd()
	assert.NoError(t, err)
	badapp := path.Join("var", "db", "baetyl", "app", "v5", "application.yml")
	goodapp := path.Join("var", "db", "baetyl", "app", "v2", "application.yml")

	// round 1: failed to reload
	utils.CopyFile(badapp, appConfigFile)
	utils.CopyFile(goodapp, appBackupFile)

	m := &Master{
		sig:       make(chan os.Signal, 1),
		pwd:       pwd,
		accounts:  cmap.New(),
		services:  cmap.New(),
		infostats: newInfoStats(pwd, "native", "", "", "var/run/baetyl.stats"),
		log:       logger.WithField("baetyl", "master"),
	}
	opts := engine.Options{
		Grace: time.Second,
		Pwd:   pwd,
	}
	m.engine, err = engine.New("native", m.infostats, opts)
	assert.NoError(t, err)

	err = m.UpdateAPP("", "")
	assert.Equal(t, "v2", m.infostats.getVersion())
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	assert.EqualError(t, err, "failed to start app: host path is empty")
	m.Close()
	m.Wait()

	os.RemoveAll(appConfigFile)
	os.RemoveAll(appBackupFile)
	m.infostats.setVersion("")

	// round 2: failed to reload
	utils.CopyFile(badapp, appConfigFile)
	utils.CopyFile(badapp, appBackupFile)
	m.engine, err = engine.New("native", m.infostats, opts)
	assert.NoError(t, err)

	err = m.UpdateAPP("", "")
	assert.Equal(t, "", m.infostats.getVersion())
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	assert.EqualError(t, err, "failed to restart old app: host path is empty; failed to roll back: host path is empty")
	m.Close()
	m.Wait()

	os.RemoveAll(appConfigFile)
	os.RemoveAll(appBackupFile)
	m.infostats.setVersion("")

	// round 2: success to reload
	utils.CopyFile(goodapp, appConfigFile)
	utils.CopyFile(badapp, appBackupFile)
	m.engine, err = engine.New("native", m.infostats, opts)

	err = m.UpdateAPP("", "")
	assert.NoError(t, err)
	assert.Equal(t, "v2", m.infostats.getVersion())
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	m.Close()
	m.Wait()
}

func TestUpdateSystemAPP(t *testing.T) {
	oldpwd, err := os.Getwd()
	assert.NoError(t, err)
	err = os.Chdir("testdata")
	assert.NoError(t, err)
	defer os.Chdir(oldpwd)

	os.RemoveAll(appConfigFile)
	os.RemoveAll(appBackupFile)
	os.RemoveAll("var/run")
	defer os.RemoveAll(appConfigFile)
	defer os.RemoveAll(appBackupFile)
	defer os.RemoveAll("var/run")

	pwd, err := os.Getwd()
	assert.NoError(t, err)

	m := &Master{
		pwd:       pwd,
		sig:       make(chan os.Signal, 1),
		accounts:  cmap.New(),
		services:  cmap.New(),
		infostats: newInfoStats(pwd, "native", "", "", "var/run/baetyl.stats"),
		log:       logger.WithField("baetyl", "master"),
	}
	opts := engine.Options{
		Grace: time.Second,
		Pwd:   pwd,
	}
	err = utils.UnmarshalYAML(nil, &m.cfg)
	assert.NoError(t, err)
	m.engine, err = engine.New("native", m.infostats, opts)
	assert.NoError(t, err)
	defer m.Wait()
	defer m.Close()

	target := path.Join("var", "db", "baetyl", "app")
	err = m.UpdateSystem("", "APP", path.Join(target, "v4"))
	assert.EqualError(t, err, "failed to update system: failed to reload config: open var/db/baetyl/app/v4/application.yml: no such file or directory")
	assert.Equal(t, "", m.infostats.getVersion())
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	assert.Equal(t, "failed to update system: failed to reload config: open var/db/baetyl/app/v4/application.yml: no such file or directory", m.infostats.getError())
	checkOTALog(t, baetyl.OTAUpdating, baetyl.OTARollingBack, baetyl.OTARolledBack)

	err = m.UpdateSystem("", "APP", path.Join(target, "v5"))
	assert.EqualError(t, err, "failed to update system: failed to start app: host path is empty")
	assert.Equal(t, "", m.infostats.getVersion())
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	assert.Equal(t, "failed to update system: failed to start app: host path is empty", m.infostats.getError())
	checkOTALog(t, baetyl.OTAUpdating, baetyl.OTARollingBack, baetyl.OTARolledBack)

	err = m.UpdateSystem("", "APP", path.Join(target, "v6"))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "wait_exit_5/lib/baetyl/cmd-nonexist/package.yml: no such file or directory")
	assert.Equal(t, "", m.infostats.getVersion())
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	assert.Contains(t, m.infostats.getError(), "wait_exit_5/lib/baetyl/cmd-nonexist/package.yml: no such file or directory")
	checkOTALog(t, baetyl.OTAUpdating, baetyl.OTARollingBack, baetyl.OTARolledBack)

	err = m.UpdateSystem("", "APP", path.Join(target, "v1"))
	assert.NoError(t, err)
	assert.Equal(t, "v1", m.infostats.getVersion())
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	assert.Equal(t, "", m.infostats.getError())
	checkOTALog(t, baetyl.OTAUpdating, baetyl.OTAUpdated)

	err = m.UpdateSystem("", "APP", path.Join(target, "v4"))
	assert.EqualError(t, err, "failed to update system: failed to reload config: open var/db/baetyl/app/v4/application.yml: no such file or directory")
	assert.Equal(t, "v1", m.infostats.getVersion())
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	assert.Equal(t, "failed to update system: failed to reload config: open var/db/baetyl/app/v4/application.yml: no such file or directory", m.infostats.getError())
	checkOTALog(t, baetyl.OTAUpdating, baetyl.OTARollingBack, baetyl.OTARolledBack)

	err = m.UpdateSystem("", "APP", path.Join(target, "v2"))
	assert.Equal(t, "v2", m.infostats.getVersion())
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	assert.Equal(t, "", m.infostats.getError())
	checkOTALog(t, baetyl.OTAUpdating, baetyl.OTAUpdated)

	err = m.UpdateSystem("", "APP", path.Join(target, "v5", "application.yml"))
	assert.EqualError(t, err, "failed to update system: failed to start app: host path is empty")
	assert.Equal(t, "v2", m.infostats.getVersion())
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	assert.Equal(t, "failed to update system: failed to start app: host path is empty", m.infostats.getError())
	checkOTALog(t, baetyl.OTAUpdating, baetyl.OTARollingBack, baetyl.OTARolledBack)

	err = m.UpdateSystem("", "APP", path.Join(target, "v3", "application.yml"))
	assert.Equal(t, "v3", m.infostats.getVersion())
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	assert.Equal(t, "", m.infostats.getError())
	checkOTALog(t, baetyl.OTAUpdating, baetyl.OTAUpdated)

	err = m.UpdateSystem("", "APP", path.Join(target, "v6", "application.yml"))
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "wait_exit_5/lib/baetyl/cmd-nonexist/package.yml: no such file or directory")
	assert.Equal(t, "v3", m.infostats.getVersion())
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	assert.Contains(t, m.infostats.getError(), "wait_exit_5/lib/baetyl/cmd-nonexist/package.yml: no such file or directory")
	checkOTALog(t, baetyl.OTAUpdating, baetyl.OTARollingBack, baetyl.OTARolledBack)

	err = m.UpdateSystem("", "APP", path.Join(target, "v2", "application.yml"))
	assert.Equal(t, "v2", m.infostats.getVersion())
	assert.True(t, utils.FileExists(appConfigFile))
	assert.False(t, utils.FileExists(appBackupFile))
	assert.Equal(t, "", m.infostats.getError())
	checkOTALog(t, baetyl.OTAUpdating, baetyl.OTAUpdated)
}

func TestUpdateSystemAPP2(t *testing.T) {
	oldpwd, err := os.Getwd()
	assert.NoError(t, err)
	err = os.Chdir("testdata")
	assert.NoError(t, err)
	defer os.Chdir(oldpwd)

	os.RemoveAll(appConfigFile)
	os.RemoveAll(appBackupFile)
	os.RemoveAll("var/run")
	defer os.RemoveAll(appConfigFile)
	defer os.RemoveAll(appBackupFile)
	defer os.RemoveAll("var/run")

	pwd, err := os.Getwd()
	assert.NoError(t, err)
	badapp := path.Join("var", "db", "baetyl", "app", "v5", "application.yml")
	goodapp := path.Join("var", "db", "baetyl", "app", "v2", "application.yml")
	otalog := "var/db/baetyl/ota.log"

	m := &Master{
		pwd:       pwd,
		sig:       make(chan os.Signal, 1),
		accounts:  cmap.New(),
		services:  cmap.New(),
		infostats: newInfoStats(pwd, "native", "", "", "var/run/baetyl.stats"),
		log:       logger.WithField("baetyl", "master"),
	}
	opts := engine.Options{
		Grace: time.Second,
		Pwd:   pwd,
	}
	err = utils.UnmarshalYAML(nil, &m.cfg)
	assert.NoError(t, err)
	defer m.Wait()
	defer m.Close()
	m.engine, err = engine.New("native", m.infostats, opts)
	assert.NoError(t, err)

	wantErr := "failed to update system: failed to start app: host path is empty"
	wantErrRB := "failed to update system: failed to restart old app: host path is empty; failed to roll back: host path is empty"
	tests := []struct {
		name        string
		target      string
		pcur        string // prepare applicatuib.yml if not empty
		pold        string // prepare applicatuib.yml.old if not empty
		ccur        bool   // check if applicatuib.yml exists
		cold        bool   // check if applicatuib.yml.old exists
		wantErr     string
		wantVersion string
		otalog      string
	}{
		{
			name:        "nil",
			target:      "",
			ccur:        false,
			cold:        false,
			wantErr:     "",
			wantVersion: "",
		},
		{
			name:        "bad app.yml",
			target:      "",
			pcur:        badapp,
			ccur:        true,
			cold:        false,
			wantErr:     wantErr,
			wantVersion: "",
		},
		{
			name:        "bad app.yml.old",
			target:      "",
			pold:        badapp,
			ccur:        false,
			cold:        false,
			wantErr:     "",
			wantVersion: "",
		},
		{
			name:        "good app.yml",
			target:      "",
			pcur:        goodapp,
			ccur:        true,
			cold:        false,
			wantErr:     "",
			wantVersion: "v2",
		},
		{
			name:        "good app.yml.old",
			target:      "",
			pold:        goodapp,
			ccur:        false,
			cold:        false,
			wantErr:     "",
			wantVersion: "",
		},
		{
			name:        "bad app.yml and app.yml.old",
			target:      "",
			pcur:        badapp,
			pold:        badapp,
			ccur:        true,
			cold:        false,
			wantErr:     wantErrRB,
			wantVersion: "",
		},
		{
			name:        "good app.yml and app.yml.old",
			target:      "",
			pcur:        goodapp,
			pold:        goodapp,
			ccur:        true,
			cold:        false,
			wantErr:     "",
			wantVersion: "v2",
		},
		{
			name:        "good app.yml and bad app.yml.old",
			target:      "",
			pcur:        goodapp,
			pold:        badapp,
			ccur:        true,
			cold:        false,
			wantErr:     "",
			wantVersion: "v2",
		},
		{
			name:        "bad app.yml and good app.yml.old",
			target:      "",
			pcur:        badapp,
			pold:        goodapp,
			ccur:        true,
			cold:        false,
			wantErr:     wantErr,
			wantVersion: "v2",
		},
	}
	for _, tt := range tests {
		os.RemoveAll(appConfigFile)
		os.RemoveAll(appBackupFile)
		os.RemoveAll(otalog)
		if tt.pcur != "" {
			utils.CopyFile(tt.pcur, appConfigFile)
		}
		if tt.pold != "" {
			utils.CopyFile(tt.pold, appBackupFile)
		}
		t.Run(tt.name, func(t *testing.T) {
			err := m.UpdateSystem("", "APP", tt.target)
			assert.Equal(t, (tt.wantErr == ""), (err == nil))
			assert.Equal(t, tt.wantErr, m.infostats.getError())
			assert.Equal(t, tt.wantVersion, m.infostats.getVersion())
			assert.Equal(t, tt.ccur, utils.FileExists(appConfigFile))
			assert.Equal(t, tt.cold, utils.FileExists(appBackupFile))
			assert.Equal(t, false, utils.FileExists(otalog))
		})
	}
}

func checkOTALog(t *testing.T, ss ...string) {
	defer os.RemoveAll("var/db/baetyl/ota.log")
	fi, err := os.Open("var/db/baetyl/ota.log")
	assert.NoError(t, err)
	defer fi.Close()

	m := map[string]string{}
	br := bufio.NewReader(fi)
	for _, s := range ss {
		l, _, err := br.ReadLine()
		assert.NoError(t, err)
		json.Unmarshal(l, &m)
		assert.Equal(t, s, m["step"])
	}
	_, _, err = br.ReadLine()
	assert.Equal(t, io.EOF, err)
}
