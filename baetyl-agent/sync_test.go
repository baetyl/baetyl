package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"testing"

	"net/http/httptest"

	"github.com/baetyl/baetyl/baetyl-agent/common"
	"github.com/baetyl/baetyl/baetyl-agent/config"
	"github.com/baetyl/baetyl/logger"
	baetylHttp "github.com/baetyl/baetyl/protocol/http"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	mock "github.com/baetyl/baetyl/sdk/baetyl-go/mock"
	"github.com/baetyl/baetyl/utils"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func initAgent(t *testing.T) *agent {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockCtx := mock.NewMockContext(mockCtl)
	mockCtx.EXPECT().Log().Return(logger.Global).AnyTimes()
	inspect := baetyl.Inspect{}
	inspect.Software.ConfVersion = "v1"
	mockCtx.EXPECT().InspectSystem().Return(&inspect, nil).AnyTimes()
	cfg := config.Config{}
	utils.SetDefaults(&cfg)
	a := agent{
		ctx:    mockCtx,
		node:   &node{Name: "test", Namespace: "default"},
		cfg:    cfg,
		events: make(chan *Event, 1),
	}
	return &a
}

func TestSyncResource(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	os.MkdirAll("var/db/baetyl/volumes/agent-conf/v1", 0755)
	expectedConfig := config.DesireResponse{Resources: []*config.Resource{
		{BaseResource: config.BaseResource{
			Type:    common.Config,
			Name:    "hub-conf",
			Version: "v1",
		},
			Value: config.ModuleConfig{
				Name:      "hub-conf",
				Namespace: "default",
				Data: map[string]string{
					"service.yml": "hub-test",
				},
				Version: "v1",
			},
		},
	}}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(expectedConfig)
		w.Write(data)
	}))
	defer ts.Close()

	ver := map[string]string{
		"hub-conf":   "v1",
		"agent-conf": "v1",
	}
	req, err := generateRequest(common.Config, ver)
	expectedReq := []*config.BaseResource{{
		Type:    common.Config,
		Name:    "hub-conf",
		Version: "v1",
	}}
	assert.Equal(t, req, expectedReq)
	assert.NoError(t, err)

	a := initAgent(t)
	a.cfg.Remote.HTTP.Address = ts.URL
	a.http, _ = baetylHttp.NewClient(*a.cfg.Remote.HTTP)
	res, err := a.syncResource(req)
	resDeploy := res[0].GetConfig()
	assert.NoError(t, err)
	assert.Equal(t, *resDeploy, expectedConfig.Resources[0].Value)
	os.Chdir(cwd)
}

func TestProcessVolumes(t *testing.T) {
	content := "test"
	configName := "hub-conf"
	filename := "service.yml"
	configs := map[string]config.ModuleConfig{
		configName: {
			Name:    configName,
			Version: "v1",
			Data: map[string]string{
				filename: content,
			},
		},
	}
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	volumePath := "var/db/baetyl/hub-conf/v1"
	volumes := map[string]baetyl.VolumeInfo{
		configName: {
			Name: configName,
			Path: volumePath,
			Meta: baetyl.Meta{
				Version: "v1",
			},
		},
	}

	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockCtx := mock.NewMockContext(mockCtl)
	mockCtx.EXPECT().Log().Return(logger.Global).AnyTimes()
	a := agent{ctx: mockCtx, node: &node{Name: "test", Namespace: "default"}}
	err = a.processVolumes(volumes, configs)
	assert.NoError(t, err)
	rp, _ := filepath.Rel(baetyl.DefaultDBDir, volumePath)
	containerDir := path.Join(baetyl.DefaultDBDir, "volumes", rp)
	volumeFile := path.Join(containerDir, filename)
	res, _ := ioutil.ReadFile(volumeFile)
	assert.Equal(t, string(res), content)
	assert.NoError(t, err)

	os.Chmod(dir, 0444)
	err = a.processVolumes(volumes, configs)
	assert.Error(t, err, "illegal path of config (hub-conf): Rel: can't make /illegal relative to var/db/baetyl")
	os.Chmod(dir, 0755)

	os.Chmod("var/db/baetyl/volumes/hub-conf/v1", 0444)
	err = a.processVolumes(volumes, configs)
	assert.Error(t, err, "failed to create file (var/db/baetyl/volumes/hub-conf/v1/service.yml): open var/db/baetyl/volumes/hub-conf/v1/service.yml: permission denied")
	os.Chmod("var/db/baetyl/volumes/hub-conf/v1", 0755)

	volumePath = "/illegal"
	volumes = map[string]baetyl.VolumeInfo{
		configName: {
			Name: configName,
			Path: volumePath,
			Meta: baetyl.Meta{
				Version: "v1",
			},
		},
	}
	err = a.processVolumes(volumes, configs)
	assert.Error(t, err, "illegal path of config (hub-conf): Rel: can't make /illegal relative to var/db/baetyl")

	os.Chdir(cwd)
}

func TestApplication(t *testing.T) {
	appName := "app"
	appVersion := "v1"
	expected := baetyl.ComposeAppConfig{Name: appName, AppVersion: appVersion}
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockCtx := mock.NewMockContext(mockCtl)
	mockCtx.EXPECT().Log().Return(logger.Global).AnyTimes()
	a := agent{ctx: mockCtx, node: &node{Name: "test", Namespace: "default"}}
	app := config.Application{
		AppConfig: expected,
	}
	_, hostDir := a.processApplication(app)
	rp, _ := filepath.Rel(baetyl.DefaultDBDir, hostDir)
	containerDir := path.Join(baetyl.DefaultDBDir, "volumes", rp)
	volumeFile := path.Join(containerDir, baetyl.AppConfFileName)
	res, _ := ioutil.ReadFile(volumeFile)
	var appCfg baetyl.ComposeAppConfig
	err = yaml.Unmarshal(res, &appCfg)
	assert.NoError(t, err)
	assert.Equal(t, appCfg, expected)

	os.Chmod(dir, 0444)
	meta, _ := a.processApplication(app)
	assert.Nil(t, meta)
	os.Chmod(dir, 0755)

	os.Chmod(containerDir, 0444)
	meta, _ = a.processApplication(app)
	assert.Nil(t, meta)
	os.Chmod(containerDir, 0755)

	os.Chdir(cwd)
}

func TestGetCurrentDeploy(t *testing.T) {
	inspect := &baetyl.Inspect{}
	inspect.Software.ConfVersion = "v1"
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	confString := `name: app
app_version: v1`
	containerDir := path.Join(baetyl.DefaultDBDir, "volumes")
	os.MkdirAll(containerDir, 0755)
	filePath := path.Join(containerDir, baetyl.AppConfFileName)
	ioutil.WriteFile(filePath, []byte(confString), 0755)
	a := agent{}
	res, err := a.getCurrentApp(inspect)
	assert.NoError(t, err)
	expected := map[string]string{
		"app": "v1",
	}
	assert.Equal(t, res, expected)

	inspect.Software.ConfVersion = ""
	_, err = a.getCurrentApp(inspect)
	assert.Error(t, err, "app version is empty")

	os.Remove(filePath)
	_, err = a.getCurrentApp(inspect)
	assert.Error(t, err, "open var/db/baetyl/volumes/application.yml: no such file or directory")
	os.Chdir(cwd)
}
