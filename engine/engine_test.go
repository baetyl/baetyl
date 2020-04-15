package engine

import (
	"fmt"
	"github.com/baetyl/baetyl-go/spec/crd"
	bh "github.com/timshannon/bolthold"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/mock"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-go/log"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func prepare(t *testing.T) (*node.Node, config.EngineConfig, *bh.Store) {
	log.Init(log.Config{Level: "debug"})

	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	no, err := node.NewNode(sto)
	assert.NoError(t, err)
	assert.NotNil(t, no)

	cfg := config.EngineConfig{}
	cfg.Kind = "kubernetes"
	cfg.Report.Interval = time.Second
	return no, cfg, sto
}

func TestEngine(t *testing.T) {
	eng, err := NewEngine(config.EngineConfig{}, nil, nil)
	assert.Error(t, err, os.ErrInvalid.Error())
	assert.Nil(t, eng)
}

func TestEngineReport(t *testing.T) {
	nod, cfg, sto := prepare(t)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockAmi := mock.NewMockAMI(mockCtl)
	r0 := v1.Report{
		"apps": []v1.AppInfo{},
	}
	r1 := v1.Report{
		"apps": []v1.AppInfo{{Name: "app", Version: "v1"}},
	}
	r2 := v1.Report{
		"apps": []v1.AppInfo{{Name: "app", Version: "v2"}},
	}
	ns := "baetyl-edge"
	app := crd.Application{}
	err := sto.Upsert(makeKey(crd.KindApplication, "app", "v2"), app)
	assert.NoError(t, err)
	var wg sync.WaitGroup
	wg.Add(1)
	gomock.InOrder(
		mockAmi.EXPECT().Collect(gomock.Any()).Return(nil, nil).Times(1),
		mockAmi.EXPECT().Collect(gomock.Any()).DoAndReturn(func(ns string) (v1.Report, error) {
			fmt.Println("1")
			_, err := nod.Desire(v1.Desire(r0))
			assert.NoError(t, err)
			return r1, nil
		}).Times(1),
		mockAmi.EXPECT().Apply(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ns string, apps []v1.AppInfo, cond string) error {
			fmt.Println("2", apps)
			sd, err := nod.Get()
			assert.NoError(t, err)
			assert.Len(t, sd.Desire.AppInfos(), 0)
			return nil
		}).Times(1),
		mockAmi.EXPECT().Collect(gomock.Any()).DoAndReturn(func(ns string) (v1.Report, error) {
			fmt.Println("3")
			_, err := nod.Desire(v1.Desire(r1))
			assert.NoError(t, err)
			return r1, nil
		}).Times(1),
		mockAmi.EXPECT().Collect(gomock.Any()).DoAndReturn(func(ns string) (v1.Report, error) {
			fmt.Println("4")
			_, err := nod.Desire(v1.Desire(r2))
			assert.NoError(t, err)
			return r1, nil
		}).Times(1),
		mockAmi.EXPECT().Apply(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ns string, apps []v1.AppInfo, cond string) error {
			fmt.Println("5", apps)
			defer wg.Done()
			sd, err := nod.Get()
			assert.NoError(t, err)
			assert.Equal(t, "v2", sd.Desire.AppInfos()[0].Version)
			return nil
		}).Times(1),
	)

	e := &Engine{
		nod: nod,
		sto: sto,
		Ami: mockAmi,
		cfg: cfg,
		ns:  ns,
		log: log.With(log.Any("engine", cfg.Kind)),
	}
	e.tomb.Go(e.reporting)
	defer e.Close()
	wg.Wait()
}

func TestSortApp(t *testing.T) {
	var reApps []v1.AppInfo
	var deApps []v1.AppInfo
	res := alignApps(reApps, deApps)
	assert.Equal(t, res, reApps)

	reApps = nil
	deApps = []v1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, reApps)

	reApps = []v1.AppInfo{}
	deApps = []v1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, reApps)

	reApps = []v1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	deApps = nil
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, reApps)

	reApps = []v1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	deApps = []v1.AppInfo{}
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, reApps)

	reApps = []v1.AppInfo{{Name: "a", Version: "a1"}, {Name: "b", Version: "b1"}}
	deApps = []v1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	expected := []v1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, expected)

	reApps = []v1.AppInfo{{Name: "a", Version: "a1"}, {Name: "b", Version: "b1"}}
	deApps = []v1.AppInfo{{Name: "b", Version: "b1"}, {Name: "c", Version: "c1"}, {Name: "a", Version: "a1"}}
	expected = []v1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, expected)

	reApps = []v1.AppInfo{{Name: "d", Version: "d1"}, {Name: "a", Version: "a1"}, {Name: "b", Version: "b1"}}
	deApps = []v1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	expected = []v1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}, {Name: "d", Version: "d1"}}
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, expected)

	reApps = []v1.AppInfo{{Name: "a", Version: "a1"}, {Name: "d", Version: "d1"}, {Name: "b", Version: "b1"}}
	deApps = []v1.AppInfo{{Name: "c", Version: "c1"}, {Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	expected = []v1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}, {Name: "d", Version: "d1"}}
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, expected)
}
