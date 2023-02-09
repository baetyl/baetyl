package sync

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mock"
	"github.com/golang/mock/gomock"

	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/mock/plugin"
	"github.com/baetyl/baetyl/v2/node"

	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/v2/store"
)

func TestSyncMakeKey(t *testing.T) {
	key := makeKey(specv1.KindApplication, "app", "a1")
	expected := "application-app-a1"
	assert.Equal(t, key, expected)
	key = makeKey(specv1.KindApplication, "", "a1")
	assert.Equal(t, key, "")
}

func TestSyncStore(t *testing.T) {
	f, err := os.CreateTemp("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())
	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)
	syn := sync{store: sto}
	app := &specv1.Application{
		Name:    "app",
		Version: "a1",
	}
	err = syn.storeApplication(app)
	assert.NoError(t, err)

	var expectedApp specv1.Application
	err = sto.Get(makeKey(specv1.KindApplication, app.Name, app.Version), &expectedApp)
	assert.NoError(t, err)
	assert.Equal(t, app, &expectedApp)
	app.Name = ""
	err = syn.storeApplication(app)
	assert.Error(t, err)

	sec := &specv1.Secret{
		Name:    "sec",
		Version: "s1",
	}
	err = syn.storeSecret(sec)
	var expectedSec specv1.Secret
	err = sto.Get(makeKey(specv1.KindSecret, sec.Name, sec.Version), &expectedSec)
	assert.NoError(t, err)
	assert.Equal(t, sec, &expectedSec)
	sec.Version = ""
	err = syn.storeSecret(sec)
	assert.Error(t, err)
}

func TestSyncProcessConfiguration(t *testing.T) {
	f, err := os.CreateTemp("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)
	content := []byte("test")

	objMs := mock.NewServer(nil, mock.NewResponse(200, content))
	assert.NotNil(t, objMs)
	sc := config.Config{}
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)
	sc.Sync.Download.Address = objMs.URL
	sc.Sync.Download.CA = "./testcert/ca.pem"
	sc.Sync.Download.Key = "./testcert/client.key"
	sc.Sync.Download.Cert = "./testcert/client.pem"
	sc.Sync.Download.InsecureSkipVerify = true
	ops, err := sc.Sync.Download.ToClientOptions()
	assert.NoError(t, err)
	syn := &sync{
		cfg:      sc,
		store:    sto,
		download: http.NewClient(ops),
		log:      log.With(log.Any("test", "sync")),
	}
	cfg := &specv1.Configuration{Name: "cfg", Version: "c1"}
	err = syn.processConfiguration(cfg)
	assert.NoError(t, err)
	var expectedCfg specv1.Configuration
	err = sto.Get(makeKey(specv1.KindConfiguration, "cfg", "c1"), &expectedCfg)
	assert.NoError(t, err)
	cfg.Name = ""
	err = syn.processConfiguration(cfg)
	assert.Error(t, err)
	cfg.Name = "cfg"

	// object process
	dir := t.TempDir()
	syn.cfg.Sync.Download.Path = dir
	file1 := filepath.Join(dir, "file1")
	os.WriteFile(file1, content, 0644)
	md5, err := utils.CalculateFileMD5(file1)
	obj := specv1.ConfigurationObject{
		URL: objMs.URL,
		MD5: md5,
	}
	objData, _ := json.Marshal(obj)
	cfg.Data = map[string]string{
		"_object_file2": string(objData),
	}
	err = syn.processConfiguration(cfg)
	assert.NoError(t, err)
	hostPath := filepath.Join(dir, "cfg", "c1")
	data, err := os.ReadFile(filepath.Join(hostPath, "file2"))
	assert.NoError(t, err)
	assert.Equal(t, data, content)

	cfg.Data = map[string]string{
		"_object_file3": "wrong",
	}
	err = syn.processConfiguration(cfg)
	assert.Error(t, err)
}

func TestSyncResources(t *testing.T) {
	f, err := os.CreateTemp("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	nod, err := node.NewNode(sto)
	assert.NoError(t, err)
	assert.NotNil(t, nod)

	appName := "test-app"
	appVer := "v1"
	namespace := "baetyl-edge"
	app := specv1.Application{
		Name:      appName,
		Namespace: namespace,
		Version:   appVer,
		Volumes: []specv1.Volume{
			{
				Name: "test-cfg",
				VolumeSource: specv1.VolumeSource{
					Config: &specv1.ObjectReference{
						Name:    "test-cfg",
						Version: "c1",
					},
				},
			},
			{
				Name: "test-sec",
				VolumeSource: specv1.VolumeSource{
					Secret: &specv1.ObjectReference{
						Name:    "test-sec",
						Version: "s1",
					},
				},
			},
		},
	}
	cfgName := "test-cfg"
	cfgVer := "c1"
	cfg := specv1.Configuration{
		Name:      cfgName,
		Namespace: namespace,
		Version:   cfgVer,
	}
	secName := "test-sec"
	secVer := "s1"
	sec := specv1.Secret{
		Name:      secName,
		Namespace: namespace,
		Version:   secVer,
	}
	appCrd := specv1.DesireResponse{
		Values: []specv1.ResourceValue{{
			ResourceInfo: specv1.ResourceInfo{Kind: specv1.KindApplication, Name: appName, Version: appVer},
			Value:        specv1.LazyValue{Value: &app},
		}},
	}
	cfgCrd := specv1.DesireResponse{
		Values: []specv1.ResourceValue{{
			ResourceInfo: specv1.ResourceInfo{Kind: specv1.KindConfiguration, Name: cfgName, Version: cfgVer},
			Value:        specv1.LazyValue{Value: &cfg},
		}},
	}
	secCrd := specv1.DesireResponse{
		Values: []specv1.ResourceValue{{
			ResourceInfo: specv1.ResourceInfo{Kind: specv1.KindSecret, Name: secName, Version: secVer},
			Value:        specv1.LazyValue{Value: &sec},
		}},
	}
	msg1 := &specv1.Message{
		Kind: specv1.MessageDesire,
		Content: specv1.LazyValue{
			Value: appCrd,
		},
	}
	dt, err := json.Marshal(msg1)
	assert.NoError(t, err)
	m1 := &specv1.Message{}
	err = json.Unmarshal(dt, m1)
	assert.NoError(t, err)

	msg2 := &specv1.Message{
		Kind: specv1.MessageDesire,
		Content: specv1.LazyValue{
			Value: cfgCrd,
		},
	}
	dt, err = json.Marshal(msg2)
	assert.NoError(t, err)
	m2 := &specv1.Message{}
	err = json.Unmarshal(dt, m2)
	assert.NoError(t, err)

	msg3 := &specv1.Message{
		Kind: specv1.MessageDesire,
		Content: specv1.LazyValue{
			Value: secCrd,
		},
	}
	dt, err = json.Marshal(msg3)
	assert.NoError(t, err)
	m3 := &specv1.Message{}
	err = json.Unmarshal(dt, m3)
	assert.NoError(t, err)

	sc := config.Config{}
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)
	assert.NoError(t, err)
	mockCtl := gomock.NewController(t)
	link := plugin.NewMockLink(mockCtl)
	link.EXPECT().Request(gomock.Any()).Return(m1, nil)
	link.EXPECT().Request(gomock.Any()).Return(m2, nil)
	link.EXPECT().Request(gomock.Any()).Return(m3, nil)
	syn := &sync{
		link:  link,
		cfg:   sc,
		store: sto,
		log:   log.With(log.Any("test", "sync")),
		nod:   nod,
	}

	err = syn.SyncResource(specv1.AppInfo{Name: "desire-app", Version: "v1"})
	var appRes specv1.Application
	err = sto.Get(makeKey(specv1.KindApplication, appName, appVer), &appRes)
	assert.NoError(t, err)
	assert.Equal(t, appRes, app)
	var cfgRes specv1.Configuration
	err = sto.Get(makeKey(specv1.KindConfiguration, cfgName, cfgVer), &cfgRes)
	assert.NoError(t, err)
	assert.Equal(t, cfgRes, cfg)
	var secRes specv1.Secret
	err = sto.Get(makeKey(specv1.KindSecret, secName, secVer), &secRes)
	assert.NoError(t, err)
	assert.Equal(t, secRes, sec)

	link.EXPECT().Request(gomock.Any()).Return(nil, errors.New("failed to sync resource"))
	err = syn.SyncResource(specv1.AppInfo{})
	assert.Error(t, err)
}
