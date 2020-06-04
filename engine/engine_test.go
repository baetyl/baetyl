package engine

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/log"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/mock"
	"github.com/baetyl/baetyl/node"
	"github.com/baetyl/baetyl/store"
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

func TestCollect(t *testing.T) {
	nod, _, _ := prepare(t)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockAmi := mock.NewMockAMI(mockCtl)
	ns := "baetyl-edge"
	e := Engine{
		ami: mockAmi,
		cfg: config.EngineConfig{},
		ns:  ns,
		nod: nod,
		log: log.With(log.Any("engine", "test")),
	}
	assert.NotNil(t, e)
	nodeInfo := &specv1.NodeInfo{}
	nodeStats := &specv1.NodeStatus{}
	info := specv1.AppInfo{
		Name:    "app1",
		Version: "v1",
	}
	apps := []specv1.AppInfo{info}
	appStats := []specv1.AppStatus{{AppInfo: info}}
	mockAmi.EXPECT().CollectNodeInfo().Return(nodeInfo, nil)
	mockAmi.EXPECT().CollectNodeStats().Return(nodeStats, nil)
	mockAmi.EXPECT().CollectAppStatus(gomock.Any()).Return(appStats, nil)
	res, err := e.Collect(ns, false)
	assert.NoError(t, err)
	resNode := res["node"]
	resNodeStats := res["nodestats"]
	resApps := res["apps"]
	resAppStats := res["appstats"]
	assert.Equal(t, resNode, nodeInfo)
	assert.Equal(t, resNodeStats, nodeStats)
	assert.Equal(t, resApps, apps)
	assert.Equal(t, resAppStats, appStats)

	mockAmi.EXPECT().CollectNodeInfo().Return(nil, errors.New("failed to get node info"))
	mockAmi.EXPECT().CollectNodeStats().Return(nodeStats, nil)
	mockAmi.EXPECT().CollectAppStatus(gomock.Any()).Return(appStats, nil)
	res, err = e.Collect(ns, false)
	assert.NoError(t, err)
	resNode = res["node"]
	assert.Nil(t, resNode)

	mockAmi.EXPECT().CollectNodeInfo().Return(nodeInfo, nil)
	mockAmi.EXPECT().CollectNodeStats().Return(nil, errors.New("failed to get node stats"))
	mockAmi.EXPECT().CollectAppStatus(gomock.Any()).Return(appStats, nil)
	res, err = e.Collect(ns, false)
	assert.NoError(t, err)
	resNodeStats = res["nodestats"]
	assert.Nil(t, resNodeStats)

	mockAmi.EXPECT().CollectNodeInfo().Return(nodeInfo, nil)
	mockAmi.EXPECT().CollectNodeStats().Return(nodeStats, nil)
	mockAmi.EXPECT().CollectAppStatus(gomock.Any()).Return(nil, errors.New("failed to get app stats"))
	res, err = e.Collect(ns, false)
	assert.NoError(t, err)
	resApps = res["apps"]
	resAppStats = res["appstats"]
	assert.Equal(t, resApps, []specv1.AppInfo{})
	assert.Nil(t, resAppStats)
}

func TestEngine(t *testing.T) {
	eng, err := NewEngine(config.EngineConfig{}, nil, nil, nil)
	assert.Error(t, err, os.ErrInvalid.Error())
	assert.Nil(t, eng)
}

func TestApplyApp(t *testing.T) {
	nod, _, sto := prepare(t)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockAmi := mock.NewMockAMI(mockCtl)
	mockSync := mock.NewMockSync(mockCtl)
	ns := "baetyl-edge"
	eng := Engine{
		ami: mockAmi,
		cfg: config.EngineConfig{},
		ns:  ns,
		sto: sto,
		syn: mockSync,
		nod: nod,
		log: log.With(log.Any("engine", "test")),
	}
	assert.NotNil(t, eng)
	mockSync.EXPECT().SyncResource(gomock.Any()).Return(nil)
	app := specv1.Application{
		Name:     "app1",
		Version:  "v1",
		Services: []specv1.Service{{}},
		Volumes: []specv1.Volume{
			{
				VolumeSource: specv1.VolumeSource{
					Config: &specv1.ObjectReference{
						Name:    "cfg1",
						Version: "c1",
					}}},
			{
				VolumeSource: specv1.VolumeSource{
					Secret: &specv1.ObjectReference{
						Name:    "sec1",
						Version: "s1",
					},
				}},
		},
	}
	cfg := specv1.Configuration{
		Name:    "cfg1",
		Version: "c1",
	}
	sec := specv1.Secret{
		Name:    "sec1",
		Version: "s1",
	}
	key := makeKey(specv1.KindApplication, "app1", "v1")
	err := sto.Upsert(key, app)
	assert.NoError(t, err)
	key = makeKey(specv1.KindConfiguration, "cfg1", "c1")
	err = sto.Upsert(key, cfg)
	assert.NoError(t, err)
	key = makeKey(specv1.KindSecret, "sec1", "s1")
	err = sto.Upsert(key, sec)
	assert.NoError(t, err)
	mockAmi.EXPECT().ApplySecrets(gomock.Any(), gomock.Any()).Return(nil)
	mockAmi.EXPECT().ApplyConfigurations(gomock.Any(), gomock.Any()).Return(nil)
	mockAmi.EXPECT().ApplyApplication(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	info := specv1.AppInfo{Name: "app1", Version: "v1"}
	err = eng.applyApp(ns, info)
	assert.NoError(t, err)

	mockSync.EXPECT().SyncResource(gomock.Any()).Return(errors.New("failed to sync resource"))
	err = eng.applyApp(ns, info)
	assert.Error(t, err)

	mockSync.EXPECT().SyncResource(gomock.Any()).Return(nil)
	mockAmi.EXPECT().ApplyConfigurations(gomock.Any(), gomock.Any()).Return(errors.New("failed to apply configuration"))
	err = eng.applyApp(ns, info)
	assert.Error(t, err)

	mockSync.EXPECT().SyncResource(gomock.Any()).Return(nil)
	mockAmi.EXPECT().ApplyConfigurations(gomock.Any(), gomock.Any()).Return(nil)
	mockAmi.EXPECT().ApplySecrets(gomock.Any(), gomock.Any()).Return(errors.New("failed to apply secret"))
	err = eng.applyApp(ns, info)
	assert.Error(t, err)

	mockSync.EXPECT().SyncResource(gomock.Any()).Return(nil)
	mockAmi.EXPECT().ApplyConfigurations(gomock.Any(), gomock.Any()).Return(nil)
	mockAmi.EXPECT().ApplySecrets(gomock.Any(), gomock.Any()).Return(nil)
	mockAmi.EXPECT().ApplyApplication(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("failed to apply application"))
	err = eng.applyApp(ns, info)
	assert.Error(t, err)
	eng.Close()
}

func TestReportAndApply(t *testing.T) {
	nod, _, sto := prepare(t)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockAmi := mock.NewMockAMI(mockCtl)
	mockSync := mock.NewMockSync(mockCtl)
	ns := "baetyl-edge"
	eng := Engine{
		ami: mockAmi,
		cfg: config.EngineConfig{},
		ns:  ns,
		sto: sto,
		syn: mockSync,
		nod: nod,
		log: log.With(log.Any("engine", "test")),
	}
	assert.NotNil(t, eng)
	mockAmi.EXPECT().CollectNodeInfo().Return(nil, nil)
	mockAmi.EXPECT().CollectNodeStats().Return(nil, nil)
	appStats := []specv1.AppStatus{{AppInfo: specv1.AppInfo{Name: "app1", Version: "v1"}}, {AppInfo: specv1.AppInfo{Name: "app2", Version: "v2"}}}
	mockAmi.EXPECT().CollectAppStatus(gomock.Any()).Return(appStats, nil)

	reApp := specv1.Report{
		"apps": []specv1.AppInfo{{Name: "app1", Version: "v1"}, {Name: "app2", Version: "v2"}},
	}
	deApp := specv1.Desire{
		"apps": []specv1.AppInfo{{Name: "app2", Version: "v2"}, {Name: "app3", Version: "v3"}},
	}
	_, err := nod.Report(reApp)
	assert.NoError(t, err)
	_, err = nod.Desire(deApp)
	assert.NoError(t, err)

	app1 := specv1.Application{Name: "app1", Version: "v1"}
	err = sto.Upsert(makeKey(specv1.KindApplication, "app1", "v1"), app1)
	assert.NoError(t, err)
	app3 := specv1.Application{Name: "app3", Version: "v3"}
	err = sto.Upsert(makeKey(specv1.KindApplication, "app3", "v3"), app3)
	mockSync.EXPECT().SyncResource(gomock.Any()).Return(nil)
	mockAmi.EXPECT().ApplyConfigurations(gomock.Any(), gomock.Any()).Return(nil)
	mockAmi.EXPECT().ApplySecrets(gomock.Any(), gomock.Any()).Return(nil)
	mockAmi.EXPECT().ApplyApplication(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockAmi.EXPECT().DeleteApplication(gomock.Any(), gomock.Any()).Return(nil)
	err = eng.reportAndApply(false, true)
	assert.NoError(t, err)

	// desire app is nil
	mockAmi.EXPECT().CollectNodeInfo().Return(nil, nil)
	mockAmi.EXPECT().CollectNodeStats().Return(nil, nil)
	appStats = []specv1.AppStatus{{AppInfo: specv1.AppInfo{Name: "app1", Version: "v1"}}}
	mockAmi.EXPECT().CollectAppStatus(gomock.Any()).Return(appStats, nil)
	reApp = specv1.Report{
		"apps": []specv1.AppInfo{{Name: "app1", Version: "v1"}},
	}
	deApp = specv1.Desire{"apps": nil}
	_, err = nod.Report(reApp)
	assert.NoError(t, err)
	_, err = nod.Desire(deApp)
	assert.NoError(t, err)
	err = eng.reportAndApply(false, true)
	assert.NoError(t, err)
}

func TestSortApp(t *testing.T) {
	var reApps []specv1.AppInfo
	var deApps []specv1.AppInfo
	res := alignApps(reApps, deApps)
	assert.Equal(t, res, reApps)

	reApps = nil
	deApps = []specv1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, reApps)

	reApps = []specv1.AppInfo{}
	deApps = []specv1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, reApps)

	reApps = []specv1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	deApps = nil
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, reApps)

	reApps = []specv1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	deApps = []specv1.AppInfo{}
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, reApps)

	reApps = []specv1.AppInfo{{Name: "a", Version: "a1"}, {Name: "b", Version: "b1"}}
	deApps = []specv1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	expected := []specv1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, expected)

	reApps = []specv1.AppInfo{{Name: "a", Version: "a1"}, {Name: "b", Version: "b1"}}
	deApps = []specv1.AppInfo{{Name: "b", Version: "b1"}, {Name: "c", Version: "c1"}, {Name: "a", Version: "a1"}}
	expected = []specv1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, expected)

	reApps = []specv1.AppInfo{{Name: "d", Version: "d1"}, {Name: "a", Version: "a1"}, {Name: "b", Version: "b1"}}
	deApps = []specv1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	expected = []specv1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}, {Name: "d", Version: "d1"}}
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, expected)

	reApps = []specv1.AppInfo{{Name: "a", Version: "a1"}, {Name: "d", Version: "d1"}, {Name: "b", Version: "b1"}}
	deApps = []specv1.AppInfo{{Name: "c", Version: "c1"}, {Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}}
	expected = []specv1.AppInfo{{Name: "b", Version: "b1"}, {Name: "a", Version: "a1"}, {Name: "d", Version: "d1"}}
	res = alignApps(reApps, deApps)
	assert.Equal(t, res, expected)
}

func TestGetServiceLog(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockAmi := mock.NewMockAMI(mockCtl)
	e := Engine{
		ami: mockAmi,
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
