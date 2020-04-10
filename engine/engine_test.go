package engine

import (
	"fmt"
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

func prepare(t *testing.T) (*node.Node, config.EngineConfig) {
	log.Init(log.Config{Level: "debug"})

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

	cfg := config.EngineConfig{}
	cfg.Kind = "kubernetes"
	cfg.Report.Interval = time.Second
	return sha, cfg
}

func TestEngine(t *testing.T) {
	eng, err := NewEngine(config.EngineConfig{}, nil, nil)
	assert.Error(t, err, os.ErrInvalid.Error())
	assert.Nil(t, eng)
}

func TestEngineReport(t *testing.T) {
	sha, cfg := prepare(t)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockAmi := mock.NewMockAMI(mockCtl)
	r0 := v1.Report{
		"apps": []interface{}{},
	}
	r1 := v1.Report{
		"apps": []interface{}{
			map[string]string{
				"name":    "app",
				"version": "v1",
			},
		},
	}
	r2 := v1.Report{
		"apps": []interface{}{
			map[string]string{
				"name":    "app",
				"version": "v2",
			},
		},
	}
	var wg sync.WaitGroup
	wg.Add(1)
	gomock.InOrder(
		mockAmi.EXPECT().Collect().Return(nil, nil).Times(1),
		mockAmi.EXPECT().Collect().DoAndReturn(func() (v1.Report, error) {
			fmt.Println("1")
			_, err := sha.Desire(v1.Desire(r0))
			assert.NoError(t, err)
			return r1, nil
		}).Times(1),
		mockAmi.EXPECT().Apply([]v1.AppInfo{}).DoAndReturn(func(apps []v1.AppInfo) error {
			fmt.Println("2", apps)
			sd, err := sha.Get()
			assert.NoError(t, err)
			assert.Len(t, sd.Desire.AppInfos(), 0)
			return nil
		}).Times(1),
		mockAmi.EXPECT().Collect().DoAndReturn(func() (v1.Report, error) {
			fmt.Println("3")
			_, err := sha.Desire(v1.Desire(r1))
			assert.NoError(t, err)
			return r1, nil
		}).Times(1),
		mockAmi.EXPECT().Collect().DoAndReturn(func() (v1.Report, error) {
			fmt.Println("4")
			_, err := sha.Desire(v1.Desire(r2))
			assert.NoError(t, err)
			return r1, nil
		}).Times(1),
		mockAmi.EXPECT().Apply([]v1.AppInfo{v1.AppInfo{Name: "app", Version: "v2"}}).DoAndReturn(func(apps []v1.AppInfo) error {
			fmt.Println("5", apps)
			defer wg.Done()
			sd, err := sha.Get()
			assert.NoError(t, err)
			assert.Equal(t, "v2", sd.Desire.AppInfos()[0].Version)
			return nil
		}).Times(1),
	)

	engine, err := NewEngine(cfg, mockAmi, sha)
	assert.NoError(t, err)
	defer engine.Close()
	wg.Wait()
}
