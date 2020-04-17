package engine

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/mock"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/spec/crd"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/golang/mock/gomock"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/stretchr/testify/assert"
	bh "github.com/timshannon/bolthold"
	"github.com/valyala/fasthttp"
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

func TestGetServiceLog(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockAmi := mock.NewMockAMI(mockCtl)
	e := Engine{
		Ami: mockAmi,
		sto: nil,
		nod: nil,
		cfg: config.EngineConfig{},
		ns:  "baetyl-edge",
		log: log.With(log.Any("engine", "any")),
	}
	assert.NotNil(t, e)

	router := routing.New()
	router.Get("/services/<service>/log", e.GetServiceLog)
	go fasthttp.ListenAndServe(":50030", router.HandleRequest)
	time.Sleep(100 * time.Millisecond)

	client := &fasthttp.Client{}

	mockAmi.EXPECT().FetchLog("baetyl-edge", "service1", int64(10), int64(60)).Return(ioutil.NopCloser(strings.NewReader("hello world")), nil).Times(1)
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	url := fmt.Sprintf("%s%s", "http://127.0.0.1:50030", "/services/service1/log?tailLines=10&sinceSeconds=60")
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	err := client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 200)
	assert.Equal(t, "hello world", string(resp.Body()))

	mockAmi.EXPECT().FetchLog("baetyl-edge", "unknown", int64(10), int64(60)).Return(nil, errors.New("error")).Times(1)
	req2 := fasthttp.AcquireRequest()
	resp2 := fasthttp.AcquireResponse()
	url2 := fmt.Sprintf("%s%s", "http://127.0.0.1:50030", "/services/unknown/log?tailLines=10&sinceSeconds=60")
	req2.SetRequestURI(url2)
	req2.Header.SetMethod("GET")
	err2 := client.Do(req2, resp2)
	assert.NoError(t, err2)
	assert.Equal(t, resp2.StatusCode(), 500)

	req3 := fasthttp.AcquireRequest()
	resp3 := fasthttp.AcquireResponse()
	url3 := fmt.Sprintf("%s%s", "http://127.0.0.1:50030", "/services/unknown/log?tailLines=a&sinceSeconds=12")
	req3.SetRequestURI(url3)
	req3.Header.SetMethod("GET")
	err3 := client.Do(req3, resp3)
	assert.NoError(t, err3)
	assert.Equal(t, resp3.StatusCode(), 400)

	req4 := fasthttp.AcquireRequest()
	resp4 := fasthttp.AcquireResponse()
	url4 := fmt.Sprintf("%s%s", "http://127.0.0.1:50030", "/services/unknown/log?tailLines=12&sinceSeconds=a")
	req4.SetRequestURI(url4)
	req4.Header.SetMethod("GET")
	err4 := client.Do(req4, resp4)
	assert.NoError(t, err4)
	assert.Equal(t, resp4.StatusCode(), 400)
}
