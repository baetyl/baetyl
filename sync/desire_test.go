package sync

import (
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/baetyl/baetyl-go/mock"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/node"
	"github.com/baetyl/baetyl/store"
	"github.com/stretchr/testify/assert"
)

func TestSyncMakeKey(t *testing.T) {
	key := makeKey(specv1.KindApplication, "app", "a1")
	expected := "application-app-a1"
	assert.Equal(t, key, expected)
	key = makeKey(specv1.KindApplication, "", "a1")
	assert.Equal(t, key, "")
}

func TestSyncStore(t *testing.T) {
	f, err := ioutil.TempFile("", t.Name())
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

func TestSyncCleanDir(t *testing.T) {
	dir, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, dir)
	dir1 := "dir1"
	err = os.Mkdir(filepath.Join(dir, dir1), 0755)
	assert.NoError(t, err)
	dir2 := "dir2"
	err = os.Mkdir(filepath.Join(dir, dir2), 0755)
	file1 := filepath.Join(dir, "file1")
	err = ioutil.WriteFile(file1, []byte("test"), 0644)
	assert.NoError(t, err)
	err = cleanDir(dir, dir1)
	assert.NoError(t, err)
	dir1Exist := utils.DirExists(filepath.Join(dir, dir1))
	assert.True(t, dir1Exist)
	dir2Exist := utils.DirExists(dir2)
	assert.False(t, dir2Exist)
	file1Exist := utils.FileExists(file1)
	assert.False(t, file1Exist)
}

func TestSyncProcessConfiguration(t *testing.T) {
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)
	content := []byte("test")

	objMs := mock.NewServer(nil, mock.NewResponse(200, content))
	assert.NotNil(t, objMs)
	sc := config.SyncConfig{}
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)
	sc.Cloud.HTTP.Address = objMs.URL
	sc.Cloud.HTTP.CA = "./testcert/ca.pem"
	sc.Cloud.HTTP.Key = "./testcert/client.key"
	sc.Cloud.HTTP.Cert = "./testcert/client.pem"
	sc.Cloud.HTTP.InsecureSkipVerify = true
	ops, err := sc.Cloud.HTTP.ToClientOptions()
	assert.NoError(t, err)
	syn := &sync{
		cfg:   sc,
		store: sto,
		http:  http.NewClient(ops),
		log:   log.With(log.Any("test", "sync")),
	}
	volume := &specv1.Volume{
		Name:         "cfg",
		VolumeSource: specv1.VolumeSource{Config: &specv1.ObjectReference{Name: "cfg", Version: "c1"}},
	}
	cfg := &specv1.Configuration{Name: "cfg", Version: "c1"}
	err = syn.processConfiguration(volume, cfg)
	assert.NoError(t, err)
	var expectedCfg specv1.Configuration
	err = sto.Get(makeKey(specv1.KindConfiguration, "cfg", "c1"), &expectedCfg)
	assert.NoError(t, err)
	cfg.Name = ""
	err = syn.processConfiguration(volume, cfg)
	assert.Error(t, err)
	cfg.Name = "cfg"

	// object process
	dir, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, dir)
	syn.cfg.Edge.DownloadPath = dir
	file1 := filepath.Join(dir, "file1")
	ioutil.WriteFile(file1, content, 0644)
	md5, err := utils.CalculateFileMD5(file1)
	obj := specv1.ConfigurationObject{
		URL: objMs.URL,
		MD5: md5,
	}
	objData, _ := json.Marshal(obj)
	cfg.Data = map[string]string{
		configKeyObject + "file2": string(objData),
	}
	err = syn.processConfiguration(volume, cfg)
	assert.NoError(t, err)
	assert.Nil(t, volume.Config)
	hostPath := filepath.Join(dir, "cfg", "c1")
	assert.Equal(t, volume.HostPath.Path, hostPath)
	data, err := ioutil.ReadFile(filepath.Join(hostPath, "file2"))
	assert.NoError(t, err)
	assert.Equal(t, data, content)

	cfg.Data = map[string]string{
		configKeyObject + "file3": "wrong",
	}
	err = syn.processConfiguration(volume, cfg)
	assert.Error(t, err)
}

func TestSyncResources(t *testing.T) {
	f, err := ioutil.TempFile("", t.Name())
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
		CRDDatas: []specv1.ResourceValue{{
			ResourceInfo: specv1.ResourceInfo{Kind: specv1.KindApplication, Name: appName, Version: appVer},
			Value:        specv1.VariableValue{Value: app},
		}},
	}
	cfgCrd := specv1.DesireResponse{
		CRDDatas: []specv1.ResourceValue{{
			ResourceInfo: specv1.ResourceInfo{Kind: specv1.KindConfiguration, Name: cfgName, Version: cfgVer},
			Value:        specv1.VariableValue{Value: cfg},
		}},
	}
	secCrd := specv1.DesireResponse{
		CRDDatas: []specv1.ResourceValue{{
			ResourceInfo: specv1.ResourceInfo{Kind: specv1.KindSecret, Name: secName, Version: secVer},
			Value:        specv1.VariableValue{Value: sec},
		}},
	}
	appData, _ := json.Marshal(appCrd)
	cfgData, _ := json.Marshal(cfgCrd)
	secData, _ := json.Marshal(secCrd)

	tlssvr, err := utils.NewTLSConfigServer(utils.Certificate{CA: "./testcert/ca.pem", Key: "./testcert/server.key", Cert: "./testcert/server.pem"})
	assert.NoError(t, err)
	assert.NotNil(t, tlssvr)
	ms := mock.NewServer(tlssvr, mock.NewResponse(200, appData),
		mock.NewResponse(200, cfgData), mock.NewResponse(200, secData))
	assert.NotNil(t, ms)
	sc := config.SyncConfig{}
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)
	sc.Cloud.HTTP.Address = ms.URL
	sc.Cloud.HTTP.CA = "./testcert/ca.pem"
	sc.Cloud.HTTP.Key = "./testcert/client.key"
	sc.Cloud.HTTP.Cert = "./testcert/client.pem"
	sc.Cloud.HTTP.InsecureSkipVerify = true
	ops, err := sc.Cloud.HTTP.ToClientOptions()
	assert.NoError(t, err)
	syn := &sync{
		cfg:   sc,
		store: sto,
		nod:   nod,
		http:  http.NewClient(ops),
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
	ms.Close()

	ms = mock.NewServer(tlssvr)
	assert.NotNil(t, ms)
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)
	sc.Cloud.HTTP.Address = ms.URL
	sc.Cloud.HTTP.CA = "./testcert/ca.pem"
	sc.Cloud.HTTP.Key = "./testcert/client.key"
	sc.Cloud.HTTP.Cert = "./testcert/client.pem"
	sc.Cloud.HTTP.InsecureSkipVerify = true
	ops, err = sc.Cloud.HTTP.ToClientOptions()
	assert.NoError(t, err)
	syn = &sync{
		store: sto,
		nod:   nod,
		cfg:   sc,
		http:  http.NewClient(ops),
		log:   log.With(log.Any("test", "sync")),
	}
	err = syn.SyncResource(specv1.AppInfo{})
	assert.Error(t, err)
	ms.Close()
}
