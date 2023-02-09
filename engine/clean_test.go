package engine

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/node"
	"github.com/baetyl/baetyl/v2/store"
)

func TestRecycle(t *testing.T) {
	f, err := os.CreateTemp("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	dir := t.TempDir()
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
	info := specv1.AppInfo{Name: app.Name, Version: app.Version}
	r.SetAppInfos(false, []specv1.AppInfo{info})
	_, err = nod.Report(r, false)
	assert.NoError(t, err)

	var cfg config.Config
	cfg.Sync.Download.Path = dir
	e := engineImpl{sto: sto, nod: nod, cfg: cfg, log: log.With()}
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

func TestGetFinishedJobs(t *testing.T) {
	apps := map[string]*specv1.Application{
		"app1": {
			Name:     "app1",
			Version:  "v1",
			Workload: specv1.WorkloadJob,
			Services: []specv1.Service{{Name: "svc1"}},
		},
		"app2": {
			Name:     "app2",
			Version:  "v1",
			Workload: specv1.WorkloadJob,
			Services: []specv1.Service{{Name: "svc2"}},
		},
		"app3": {
			Name:     "app3",
			Version:  "v1",
			Workload: specv1.WorkloadDeployment,
			Services: []specv1.Service{{Name: "svc3"}},
		},
		"app4": {
			Name:     "app4",
			Version:  "v1",
			Workload: specv1.WorkloadJob,
			Services: []specv1.Service{{Name: "svc4"}},
		},
	}
	nod := &specv1.Node{
		Name:    "node1",
		Version: "v1",
		Report:  map[string]interface{}{},
	}
	appstats := []specv1.AppStats{{
		InstanceStats: map[string]specv1.InstanceStats{
			"job1": {
				AppName: "app1",
				Status:  specv1.Succeeded,
			},
			"job2": {
				AppName: "app1",
				Status:  specv1.Failed,
			},
			"job3": {
				AppName: "app2",
				Status:  specv1.Succeeded,
			},
			"job4": {
				AppName: "app2",
				Status:  specv1.Succeeded,
			},
			"job5": {
				AppName: "app4",
				Status:  specv1.Failed,
			},
			"job6": {
				AppName: "app4",
				Status:  specv1.Succeeded,
			},
			"deploy": {
				AppName: "app3",
				Status:  specv1.Running,
			},
		},
	}}
	nod.Report.SetAppStats(false, appstats)
	res := getFinishedJobs(apps, nod)
	expected := map[string]struct{}{
		"app2": {},
	}
	assert.Equal(t, expected, res)
}

func TestGetUsedObjectCfgs(t *testing.T) {
	finishedJobs := map[string]struct{}{
		"app2": {},
	}
	apps := map[string]*specv1.Application{
		"app1": {
			Name:     "app1",
			Version:  "v1",
			Workload: specv1.WorkloadJob,
			Volumes: []specv1.Volume{{
				Name: "vm1",
				VolumeSource: specv1.VolumeSource{
					Config: &specv1.ObjectReference{
						Name:    "cfg1",
						Version: "v1",
					},
				},
			}},
			Services: []specv1.Service{{
				Name: "svc1",
				VolumeMounts: []specv1.VolumeMount{{
					Name: "vm1",
				}},
			}},
		},
		"app2": {
			Name:     "app2",
			Version:  "v2",
			Workload: specv1.WorkloadJob,
			Volumes: []specv1.Volume{{
				Name: "vm2",
				VolumeSource: specv1.VolumeSource{
					Config: &specv1.ObjectReference{
						Name:    "cfg2",
						Version: "v2",
					},
				},
			}},
			Services: []specv1.Service{{
				Name: "svc2",
				VolumeMounts: []specv1.VolumeMount{{
					Name: "vm2",
				}},
			}},
		},
		"app3": {
			Name:     "app3",
			Version:  "v3",
			Workload: specv1.WorkloadJob,
			Volumes: []specv1.Volume{{
				Name: "vm3",
				VolumeSource: specv1.VolumeSource{
					Config: &specv1.ObjectReference{
						Name:    "cfg3",
						Version: "v3",
					},
				},
			}},
			Services: []specv1.Service{{
				Name: "svc3",
				VolumeMounts: []specv1.VolumeMount{{
					Name: "vm3",
				}},
			}},
		},
	}
	res := getUsedObjectCfgs(apps, finishedJobs)
	expected := map[string]string{
		"cfg1": "v1",
		"cfg3": "v3",
	}
	assert.Equal(t, expected, res)
}

func TestGetDelObjectCfgs(t *testing.T) {
	finishedJobs := map[string]struct{}{
		"app2": {},
	}
	cfg1 := specv1.Configuration{Name: "cfg1", Version: "v1"}
	cfg2 := specv1.Configuration{Name: "cfg2", Version: "v2"}
	cfg3 := specv1.Configuration{Name: "cfg3", Version: "v3"}
	objectCfgs := map[string]*specv1.Configuration{
		makeKey(specv1.KindConfiguration, "cfg1", "v1"): &cfg1,
		makeKey(specv1.KindConfiguration, "cfg2", "v2"): &cfg2,
		makeKey(specv1.KindConfiguration, "cfg3", "v3"): &cfg3,
	}
	occupiedApps := map[string]*specv1.Application{
		"app1": {
			Name:    "app1",
			Version: "v1",
			Volumes: []specv1.Volume{{
				Name: "vm1",
				VolumeSource: specv1.VolumeSource{
					Config: &specv1.ObjectReference{
						Name:    "cfg1",
						Version: "v1",
					},
				},
			}},
			Services: []specv1.Service{{
				Name: "svc1",
				VolumeMounts: []specv1.VolumeMount{{
					Name:      "vm1",
					AutoClean: false,
				}},
			}},
		},
		"app2": {
			Name:    "app2",
			Version: "v1",
			Volumes: []specv1.Volume{{
				Name: "vm2",
				VolumeSource: specv1.VolumeSource{
					Config: &specv1.ObjectReference{
						Name:    "cfg2",
						Version: "v2",
					},
				},
			}},
			Services: []specv1.Service{{
				Name: "svc2",
				VolumeMounts: []specv1.VolumeMount{{
					Name:      "vm2",
					AutoClean: true,
				}},
			}},
		},
	}
	obsoleteApps := map[string]*specv1.Application{
		"app3": {
			Name:     "app3",
			Version:  "v2",
			Workload: specv1.WorkloadJob,
			Volumes: []specv1.Volume{{
				Name: "vm1",
				VolumeSource: specv1.VolumeSource{
					Config: &specv1.ObjectReference{
						Name:    "cfg3",
						Version: "v3",
					},
				},
			}, {
				Name: "vm2",
				VolumeSource: specv1.VolumeSource{
					Config: &specv1.ObjectReference{
						Name:    "cfg4",
						Version: "v4",
					},
				},
			}, {
				Name: "vm3",
				VolumeSource: specv1.VolumeSource{
					Config: &specv1.ObjectReference{
						Name:    "cfg5",
						Version: "v5",
					},
				},
			}},
			Services: []specv1.Service{{
				Name: "svc2",
				VolumeMounts: []specv1.VolumeMount{{
					Name:      "vm1",
					AutoClean: true,
				}, {
					Name:      "vm2",
					AutoClean: false,
				}, {
					Name:      "vm3",
					AutoClean: true,
				}},
			}},
		},
		"app4": {
			Name:    "app4",
			Version: "v2",
			Volumes: []specv1.Volume{{
				Name: "vm1",
				VolumeSource: specv1.VolumeSource{
					Config: &specv1.ObjectReference{
						Name:    "cfg3",
						Version: "v3",
					},
				},
			}, {
				Name: "vm2",
				VolumeSource: specv1.VolumeSource{
					Config: &specv1.ObjectReference{
						Name:    "cfg4",
						Version: "v4",
					},
				},
			}},
			Services: []specv1.Service{{
				Name: "svc3",
				VolumeMounts: []specv1.VolumeMount{{
					Name:      "vm1",
					AutoClean: false,
				}, {
					Name:      "vm2",
					AutoClean: true,
				}},
			}},
		},
	}
	del := getDelObjectCfgs(occupiedApps, obsoleteApps, objectCfgs, finishedJobs)
	expected := map[string]*specv1.Configuration{
		makeKey(specv1.KindConfiguration, cfg2.Name, cfg2.Version): &cfg2,
		makeKey(specv1.KindConfiguration, cfg3.Name, cfg3.Version): &cfg3,
	}
	assert.Equal(t, expected, del)
}

func TestCleanObjectStorage(t *testing.T) {
	objDir := t.TempDir()
	nod, engCfg, sto := prepare(t)
	node1 := &specv1.Node{
		Name:    "node1",
		Version: "v1",
		Report:  map[string]interface{}{},
		Desire:  map[string]interface{}{},
	}
	appstats := []specv1.AppStats{{
		InstanceStats: map[string]specv1.InstanceStats{
			"job1": {
				AppName: "svc1",
				Status:  specv1.Succeeded,
			},
			"job2": {
				AppName: "svc1",
				Status:  specv1.Failed,
			},
			"deploy": {
				AppName: "svc1",
				Status:  specv1.Running,
			},
		},
	}}
	node1.Report.SetAppStats(false, appstats)
	appinfo := []specv1.AppInfo{{
		Name:    "app2",
		Version: "v2",
	}}
	node1.Report.SetAppInfos(false, appinfo)
	_, err := nod.Report(node1.Report, false)
	assert.NoError(t, err)
	objCfg1 := specv1.Configuration{
		Name:    "cfg1",
		Version: "v1",
		Data: map[string]string{
			specv1.PrefixConfigObject: "a.zip",
		},
	}
	objCfg2 := specv1.Configuration{
		Name:    "cfg2",
		Version: "v2",
		Data: map[string]string{
			specv1.PrefixConfigObject: "b.zip",
		},
	}
	objCfg3 := specv1.Configuration{
		Name:    "cfg3",
		Version: "v3",
		Data: map[string]string{
			specv1.PrefixConfigObject: "c.zip",
		},
	}
	err = sto.Upsert(makeKey(specv1.KindConfiguration, objCfg1.Name, objCfg1.Version), objCfg1)
	assert.NoError(t, err)
	err = sto.Upsert(makeKey(specv1.KindConfiguration, objCfg2.Name, objCfg2.Version), objCfg2)
	assert.NoError(t, err)
	err = sto.Upsert(makeKey(specv1.KindConfiguration, objCfg3.Name, objCfg3.Version), objCfg3)
	assert.NoError(t, err)
	app1 := specv1.Application{
		Name:    "app1",
		Version: "v1",
		Volumes: []specv1.Volume{{
			Name: "vm1",
			VolumeSource: specv1.VolumeSource{
				Config: &specv1.ObjectReference{
					Name:    objCfg1.Name,
					Version: objCfg1.Version,
				},
			},
		}, {
			Name: "vm2",
			VolumeSource: specv1.VolumeSource{
				Config: &specv1.ObjectReference{
					Name:    objCfg2.Name,
					Version: objCfg2.Version,
				},
			},
		}, {
			Name: "vm3",
			VolumeSource: specv1.VolumeSource{
				Config: &specv1.ObjectReference{
					Name:    objCfg3.Name,
					Version: objCfg3.Version,
				},
			},
		}},
		Services: []specv1.Service{{
			Name: "svc1",
			VolumeMounts: []specv1.VolumeMount{{
				Name:      "vm1",
				AutoClean: true,
			}, {
				Name:      "vm2",
				AutoClean: false,
			}, {
				Name:      "vm3",
				AutoClean: true,
			}},
		}},
	}
	app2 := specv1.Application{
		Name:    "app2",
		Version: "v2",
		Volumes: []specv1.Volume{{
			Name: "vm1",
			VolumeSource: specv1.VolumeSource{
				Config: &specv1.ObjectReference{
					Name:    objCfg3.Name,
					Version: objCfg3.Version,
				},
			},
		}},
		Services: []specv1.Service{{
			Name: "svc2",
			VolumeMounts: []specv1.VolumeMount{{
				Name: "vm1",
			}},
		}},
	}
	err = sto.Upsert(makeKey(specv1.KindApplication, app1.Name, app1.Version), app1)
	assert.NoError(t, err)
	err = sto.Upsert(makeKey(specv1.KindApplication, app2.Name, app2.Version), app2)
	assert.NoError(t, err)

	cfgDir1 := path.Join(objDir, objCfg1.Name, objCfg1.Version)
	cfgDir2 := path.Join(objDir, objCfg2.Name, objCfg2.Version)
	cfgDir3 := path.Join(objDir, objCfg3.Name, objCfg3.Version)
	err = os.MkdirAll(cfgDir1, 0755)
	assert.NoError(t, err)
	assert.True(t, utils.DirExists(cfgDir1))
	err = os.MkdirAll(cfgDir2, 0755)
	assert.NoError(t, err)
	assert.True(t, utils.DirExists(cfgDir2))
	err = os.MkdirAll(cfgDir3, 0755)
	assert.NoError(t, err)
	assert.True(t, utils.DirExists(cfgDir3))
	eng := &engineImpl{
		nod: nod,
		sto: sto,
		cfg: config.Config{Engine: engCfg},
		log: log.With(log.Any("engine", "clean")),
	}
	eng.cfg.Sync.Download.Path = objDir
	res, err := eng.cleanObjectStorage()
	assert.NoError(t, err)
	assert.Equal(t, 1, res)
	assert.False(t, utils.DirExists(cfgDir1))
	assert.True(t, utils.DirExists(cfgDir2))
	assert.True(t, utils.DirExists(cfgDir3))
}
