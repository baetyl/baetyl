package main

import (
	"context"
	"encoding/json"
	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl/baetyl-agent/common"
	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	mock "github.com/baetyl/baetyl/sdk/baetyl-go/mock"
	"github.com/baetyl/baetyl/utils"
	"github.com/golang/mock/gomock"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"testing"
)

type TestApp struct {
	AppName    string `yaml:"appName,omitempty"`
	AppVersion string `yaml:"appVersion,omitempty"`
}

func TestTransform(t *testing.T) {
	d := map[string]interface{}{
		"appName":    "hub",
		"appVersion": "32423",
	}
	var app TestApp
	err := mapstructure.Decode(d, &app)
	assert.NoError(t, err)
	expected := TestApp{
		AppName:    "hub",
		AppVersion: "32423",
	}
	assert.Equal(t, app, expected)
}

func TestCheckVolumeExists(t *testing.T) {
	dir, err := ioutil.TempDir("", "template")
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	assert.NoError(t, err)
	vName := "hub-conf"
	containerDir := path.Join(baetyl.DefaultDBDir, "volumes", vName, "v1")
	os.MkdirAll(containerDir, 0755)
	volume := baetyl.VolumeInfo{Name: vName}
	volume.Path = "var/db/baetyl/hub-conf/v1"
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockCtx := mock.NewMockContext(mockCtl)
	mockCtx.EXPECT().Log().Return(logger.Global).AnyTimes()
	a := agent{
		ctx: mockCtx,
	}
	exists := a.checkVolumeExists(volume)
	assert.True(t, exists)
	os.Chdir(cwd)
}

func TestGetCurrentDeploy(t *testing.T) {
	inspect := &baetyl.Inspect{}
	inspect.Software.ConfVersion = "v1"
	dir, err := ioutil.TempDir("", "template")
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
	res, err := a.getCurrentDeploy(inspect)
	assert.NoError(t, err)
	expected := map[string]string{
		"app": "v1",
	}
	assert.Equal(t, res, expected)
	os.Chdir(cwd)
}

func TestSendData(t *testing.T) {
	serv := initMockLinkServer()
	defer serv.Close()
	data := map[string]string{
		"test": "test",
	}
	content, _ := json.Marshal(data)
	msg := &link.Message{}
	msg.Content = content
	ret := &link.Message{}
	ret.Content = []byte{2, 1}
	serv.AddMsg(ret)
	cfg := link.ClientConfig{
		Address: "127.0.0.1:50006",
	}
	utils.SetDefaults(&cfg)
	cli, _ := link.NewClient(cfg, nil)
	a := agent{link: cli}
	res, err := a.sendData(data)
	assert.NoError(t, err)
	assert.Equal(t, res, ret.Content)
}

func TestProcessVolume(t *testing.T) {
	serv := initMockLinkServer()
	defer serv.Close()
	msg := &link.Message{}
	content := "test"
	configName := "hub-conf"
	filename := "service.yml"
	info := BackwardInfo{
		Metadata: map[string]interface{}{
			string(common.Config): ModuleConfig{
				Name: configName,
				Data: map[string]string{
					filename: content,
				},
				Version: "v1",
			},
		},
	}
	msg.Content, _ = json.Marshal(info)
	serv.AddMsg(msg)
	dir, err := ioutil.TempDir("", "template")
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

	cfg := link.ClientConfig{
		Address: "127.0.0.1:50006",
	}
	utils.SetDefaults(&cfg)
	cli, _ := link.NewClient(cfg, nil)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockCtx := mock.NewMockContext(mockCtl)
	mockCtx.EXPECT().Log().Return(logger.Global).AnyTimes()
	a := agent{link: cli, ctx: mockCtx, node: &node{Name: "test", Namespace: "default"}}
	err = a.processVolumes(volumes)
	assert.NoError(t, err)
	rp, _ := filepath.Rel(baetyl.DefaultDBDir, volumePath)
	containerDir := path.Join(baetyl.DefaultDBDir, "volumes", rp)
	volumeFile := path.Join(containerDir, filename)
	res, _ := ioutil.ReadFile(volumeFile)
	assert.Equal(t, string(res), content)
	assert.NoError(t, err)
	os.Chdir(cwd)
}

func TestProcessApplication(t *testing.T) {
	serv := initMockLinkServer()
	defer serv.Close()
	msg := &link.Message{}
	appName := "app"
	appVersion := "v1"
	expected := baetyl.ComposeAppConfig{Name: appName, AppVersion: appVersion}
	info := BackwardInfo{
		Metadata: map[string]interface{}{
			string(common.Application): DeployConfig{
				AppConfig: expected,
				Metadata:  nil,
			},
		},
	}
	msg.Content, _ = json.Marshal(info)
	serv.AddMsg(msg)
	dir, err := ioutil.TempDir("", "template")
	assert.NoError(t, err)
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	cfg := link.ClientConfig{
		Address: "127.0.0.1:50006",
	}
	utils.SetDefaults(&cfg)
	cli, _ := link.NewClient(cfg, nil)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockCtx := mock.NewMockContext(mockCtl)
	mockCtx.EXPECT().Log().Return(logger.Global).AnyTimes()
	a := agent{link: cli, ctx: mockCtx, node: &node{Name: "test", Namespace: "default"}}
	appInfo := map[string]string{
		appName: appVersion,
	}
	_, hostDir := a.processApplication(appInfo)
	rp, _ := filepath.Rel(baetyl.DefaultDBDir, hostDir)
	containerDir := path.Join(baetyl.DefaultDBDir, "volumes", rp)
	volumeFile := path.Join(containerDir, baetyl.AppConfFileName)
	res, _ := ioutil.ReadFile(volumeFile)
	var appCfg baetyl.ComposeAppConfig
	err = yaml.Unmarshal(res, &appCfg)
	assert.NoError(t, err)
	assert.Equal(t, appCfg, expected)
	os.Chdir(cwd)
}

func initMockLinkServer() *mockLinkServer {
	testAddr := "0.0.0.0:50006"
	cfg := link.ServerConfig{
		Address: testAddr,
	}
	utils.SetDefaults(&cfg)
	s, _ := link.NewServer(cfg, nil)
	ms := &mockLinkServer{s: s, msgs: make(chan *link.Message, 10)}
	link.RegisterLinkServer(s, ms)
	lis, _ := net.Listen("tcp", testAddr)
	go s.Serve(lis)
	return ms
}

type mockLinkServer struct {
	msgs chan *link.Message
	s    *grpc.Server
}

func (l *mockLinkServer) AddMsg(msg *link.Message) {
	l.msgs <- msg
}

func (l *mockLinkServer) Call(ctx context.Context, msg *link.Message) (*link.Message, error) {
	m := <-l.msgs
	return m, nil
}

func (l *mockLinkServer) Talk(stream link.Link_TalkServer) error {
	return nil
}

func (l *mockLinkServer) Close() {
	l.s.Stop()
}
