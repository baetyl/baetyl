package sync

import (
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-go/mock"
	"github.com/baetyl/baetyl-go/spec/crd"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestSyncMakeKey(t *testing.T) {
	key := makeKey(crd.KindApplication, "app", "a1")
	expected := "application-app-a1"
	assert.Equal(t, key, expected)
	key = makeKey(crd.KindApplication, "", "a1")
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
	syn := Sync{store: sto}
	app := &crd.Application{
		Name:    "app",
		Version: "a1",
	}
	err = syn.storeApplication(app)
	assert.NoError(t, err)

	var expectedApp crd.Application
	err = sto.Get(makeKey(crd.KindApplication, app.Name, app.Version), &expectedApp)
	assert.NoError(t, err)
	assert.Equal(t, app, &expectedApp)
	app.Name = ""
	err = syn.storeApplication(app)
	assert.Error(t, err)

	sec := &crd.Secret{
		Name:    "sec",
		Version: "s1",
	}
	err = syn.storeSecret(sec)
	var expectedSec crd.Secret
	err = sto.Get(makeKey(crd.KindSecret, sec.Name, sec.Version), &expectedSec)
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
	syn, err := NewSync(sc, sto, nil)
	assert.NoError(t, err)

	volume := &crd.Volume{
		Name:         "cfg",
		VolumeSource: crd.VolumeSource{Config: &crd.ObjectReference{Name: "cfg", Version: "c1"}},
	}
	cfg := &crd.Configuration{Name: "cfg", Version: "c1"}
	err = syn.processConfiguration(volume, cfg)
	assert.NoError(t, err)
	var expectedCfg crd.Configuration
	err = sto.Get(makeKey(crd.KindConfiguration, "cfg", "c1"), &expectedCfg)
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
	obj := specv1.CRDConfigObject{
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

	sha, err := node.NewNode(t.Name(), t.Name(), sto)
	assert.NoError(t, err)
	assert.NotNil(t, sha)

	appName := "test-app"
	appVer := "v1"
	namespace := "baetyl-edge"
	app := crd.Application{
		Name:      appName,
		Namespace: namespace,
		Version:   appVer,
		Volumes: []crd.Volume{
			{
				Name: "test-cfg",
				VolumeSource: crd.VolumeSource{
					Config: &crd.ObjectReference{
						Name:    "test-cfg",
						Version: "c1",
					},
				},
			},
			{
				Name: "test-sec",
				VolumeSource: crd.VolumeSource{
					Secret: &crd.ObjectReference{
						Name:    "test-sec",
						Version: "s1",
					},
				},
			},
		},
	}
	cfgName := "test-cfg"
	cfgVer := "c1"
	cfg := crd.Configuration{
		Name:      cfgName,
		Namespace: namespace,
		Version:   cfgVer,
	}
	secName := "test-sec"
	secVer := "s1"
	sec := crd.Secret{
		Name:      secName,
		Namespace: namespace,
		Version:   secVer,
	}
	appCrd := specv1.CRDResponse{
		CRDDatas: []specv1.CRDData{{
			CRDInfo: specv1.CRDInfo{Kind: crd.KindApplication, Name: appName, Version: appVer},
			Value:   specv1.VariableValue{Value: app},
		}},
	}
	cfgCrd := specv1.CRDResponse{
		CRDDatas: []specv1.CRDData{{
			CRDInfo: specv1.CRDInfo{Kind: crd.KindConfiguration, Name: cfgName, Version: cfgVer},
			Value:   specv1.VariableValue{Value: cfg},
		}},
	}
	secCrd := specv1.CRDResponse{
		CRDDatas: []specv1.CRDData{{
			CRDInfo: specv1.CRDInfo{Kind: crd.KindSecret, Name: secName, Version: secVer},
			Value:   specv1.VariableValue{Value: sec},
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
	syn, err := NewSync(sc, sto, sha)
	assert.NoError(t, err)
	err = syn.syncResources([]specv1.AppInfo{specv1.AppInfo{Name: "desire-app", Version: "v1"}})
	var appRes crd.Application
	err = sto.Get(makeKey(crd.KindApplication, appName, appVer), &appRes)
	assert.NoError(t, err)
	assert.Equal(t, appRes, app)
	var cfgRes crd.Configuration
	err = sto.Get(makeKey(crd.KindConfiguration, cfgName, cfgVer), &cfgRes)
	assert.NoError(t, err)
	assert.Equal(t, cfgRes, cfg)
	var secRes crd.Secret
	err = sto.Get(makeKey(crd.KindSecret, secName, secVer), &secRes)
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
	syn, err = NewSync(sc, sto, sha)
	assert.NoError(t, err)
	err = syn.syncResources([]specv1.AppInfo{})
	assert.NoError(t, err)
	ms.Close()
}
