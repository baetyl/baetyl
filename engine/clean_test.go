package engine

import (
	"fmt"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/node"
	"github.com/baetyl/baetyl/store"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestRecycle(t *testing.T) {
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	dir, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, dir)
	fmt.Println("-->tempdir", dir)

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	nod, err := node.NewNode(sto)
	assert.NoError(t, err)
	assert.NotNil(t, nod)

	cfg1 := specv1.Configuration{Name: "cfg-meta-1", Version: "meta-1", Data: map[string]string{"_object_cfg1": "cfg1"}}
	cfg2 := specv1.Configuration{Name: "cfg-meta-2", Version: "meta-2", Data: map[string]string{"_object_cfg2": "cfg2"}}
	cfg3 := specv1.Configuration{Name: "cfg-3", Version: "3", Data: map[string]string{"cfg3": "cfg3"}}

	path1 := filepath.Join(dir, cfg1.Name, cfg1.Version)
	path2 := filepath.Join(dir, cfg2.Name, cfg2.Version)
	os.MkdirAll(filepath.Join(dir, cfg1.Name, cfg1.Version), 0755)
	os.MkdirAll(filepath.Join(dir, cfg2.Name, cfg2.Version), 0755)
	assert.True(t, utils.DirExists(path1))
	assert.True(t, utils.DirExists(path2))

	key1 := makeKey(specv1.KindConfiguration, cfg1.Name, cfg1.Version)
	key2 := makeKey(specv1.KindConfiguration, cfg2.Name, cfg2.Version)
	key3 := makeKey(specv1.KindConfiguration, cfg3.Name, cfg3.Version)
	sto.Upsert(key1, cfg1)
	sto.Upsert(key2, cfg2)
	sto.Upsert(key3, cfg3)

	app := specv1.Application{Name: "app-1", Version: "1"}
	vol := specv1.Volume{
		Name: "cfg1",
		VolumeSource: specv1.VolumeSource{
			Config: &specv1.ObjectReference{Name: "cfg-meta-1", Version: "meta-1"},
		},
	}
	app.Volumes = append(app.Volumes, vol)
	appKey := makeKey(specv1.KindApplication, app.Name, app.Version)
	sto.Upsert(appKey, app)

	r := specv1.Report{}
	info := specv1.AppInfo{Name: app.Name, Version:app.Version}
	r.SetAppInfos(false, []specv1.AppInfo{info})
	_, err = nod.Report(r)
	assert.NoError(t, err)

	var cfg config.Config
	cfg.Sync.Download.Path = dir
	e := Engine{sto: sto, nod: nod, cfg: cfg, log: log.With()}
	err = e.recycle()
	assert.NoError(t, err)

	path1 = filepath.Join(dir, cfg1.Name)
	path2 = filepath.Join(dir, cfg2.Name)
	assert.True(t, utils.DirExists(path1))
	assert.False(t, utils.DirExists(path2))

	var res1 specv1.Configuration
	sto.Get(key1, &res1)
	assert.NotNil(t, res1)
	var res2 specv1.Configuration
	err = sto.Get(key2, &res2)
	assert.Error(t, err)
	var res3 specv1.Configuration
	sto.Get(key3, &res3)
	assert.NotNil(t, res3)
}
