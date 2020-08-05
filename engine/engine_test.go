package engine

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/baetyl/baetyl-go/v2/context"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/pki"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
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

const (
	caCrt = `
-----BEGIN CERTIFICATE-----
MIICjTCCAjKgAwIBAgIIFiYYXpptZ7AwCgYIKoZIzj0EAwIwgawxCzAJBgNVBAYT
AkNOMRAwDgYDVQQIEwdCZWlqaW5nMRkwFwYDVQQHExBIYWlkaWFuIERpc3RyaWN0
MRUwEwYDVQQJEwxCYWlkdSBDYW1wdXMxDzANBgNVBBETBjEwMDA5MzEeMBwGA1UE
ChMVTGludXggRm91bmRhdGlvbiBFZGdlMQ8wDQYDVQQLEwZCQUVUWUwxFzAVBgNV
BAMTDmRlZmF1bHQuMDcyOTAxMCAXDTIwMDcyOTAyMzE1MloYDzIwNzAwNzE3MDIz
MTUyWjCBrDELMAkGA1UEBhMCQ04xEDAOBgNVBAgTB0JlaWppbmcxGTAXBgNVBAcT
EEhhaWRpYW4gRGlzdHJpY3QxFTATBgNVBAkTDEJhaWR1IENhbXB1czEPMA0GA1UE
ERMGMTAwMDkzMR4wHAYDVQQKExVMaW51eCBGb3VuZGF0aW9uIEVkZ2UxDzANBgNV
BAsTBkJBRVRZTDEXMBUGA1UEAxMOZGVmYXVsdC4wNzI5MDEwWTATBgcqhkjOPQIB
BggqhkjOPQMBBwNCAASIpuCgm+W8OIb6njb4K8XCBnuGCNNkGwmFX1S45Y0A0Nm1
Fbi/bmTL4zeyxfzDYkMSzzb3rVra9r07OMd4zTeLozowODAOBgNVHQ8BAf8EBAMC
AYYwDwYDVR0TAQH/BAUwAwEB/zAVBgNVHREEDjAMhwQAAAAAhwR/AAABMAoGCCqG
SM49BAMCA0kAMEYCIQCDw7CMJ8V2ZvKwPx4uRpb0WFOfDn28mah0FqiCenbGqQIh
AM2I0IByWzc+9vOcoHB1sl8DY2128sSWwTBlMPU8A6yt
-----END CERTIFICATE-----
`
	crt = `
-----BEGIN CERTIFICATE-----
MIICmDCCAj+gAwIBAgIIFiYYYP2g1WgwCgYIKoZIzj0EAwIwgawxCzAJBgNVBAYT
AkNOMRAwDgYDVQQIEwdCZWlqaW5nMRkwFwYDVQQHExBIYWlkaWFuIERpc3RyaWN0
MRUwEwYDVQQJEwxCYWlkdSBDYW1wdXMxDzANBgNVBBETBjEwMDA5MzEeMBwGA1UE
ChMVTGludXggRm91bmRhdGlvbiBFZGdlMQ8wDQYDVQQLEwZCQUVUWUwxFzAVBgNV
BAMTDmRlZmF1bHQuMDcyOTAxMB4XDTIwMDcyOTAyMzIwMloXDTQwMDcyNDAyMzIw
Mlowga0xCzAJBgNVBAYTAkNOMRAwDgYDVQQIEwdCZWlqaW5nMRkwFwYDVQQHExBI
YWlkaWFuIERpc3RyaWN0MRUwEwYDVQQJEwxCYWlkdSBDYW1wdXMxDzANBgNVBBET
BjEwMDA5MzEeMBwGA1UEChMVTGludXggRm91bmRhdGlvbiBFZGdlMQ8wDQYDVQQL
EwZCQUVUWUwxGDAWBgNVBAMTD2JhZXR5bC1mdW5jdGlvbjBZMBMGByqGSM49AgEG
CCqGSM49AwEHA0IABH0y7lZWNCo512UgbZFzbZodPk+aO0fX14TXzITqnmYoK7Rk
9rTSprk8lx7JwVxTz6Opv7XKh7yDknpPSSLq7QKjSDBGMA4GA1UdDwEB/wQEAwIF
oDAPBgNVHSUECDAGBgRVHSUAMAwGA1UdEwEB/wQCMAAwFQYDVR0RBA4wDIcEAAAA
AIcEfwAAATAKBggqhkjOPQQDAgNHADBEAiAC3PluuUxcoVnvz8JtaHrQumEJNeo/
Ja9CCrkp24b8rQIgT/+ZbszAFlVO76iI7AtgoJ0cg7hUFjZHVgxh3diCuhY=
-----END CERTIFICATE-----
`
	key = `
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEILraKdvNbV2kwWHbCNecVCvaWJezGthwxTZfMtDCAV4aoAoGCCqGSM49
AwEHoUQDQgAEfTLuVlY0KjnXZSBtkXNtmh0+T5o7R9fXhNfMhOqeZigrtGT2tNKm
uTyXHsnBXFPPo6m/tcqHvIOSek9JIurtAg==
-----END EC PRIVATE KEY-----
`
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
		cfg: config.Config{},
		ns:  ns,
		nod: nod,
		log: log.With(log.Any("engine", "test")),
	}
	assert.NotNil(t, e)
	nodeInfo := &specv1.NodeInfo{}
	nodeStats := &specv1.NodeStats{}
	info := specv1.AppInfo{
		Name:    "app1",
		Version: "v1",
	}
	apps := []specv1.AppInfo{info}
	appStats := []specv1.AppStats{{AppInfo: info}}
	mockAmi.EXPECT().CollectNodeInfo().Return(nodeInfo, nil)
	mockAmi.EXPECT().CollectNodeStats().Return(nodeStats, nil)
	mockAmi.EXPECT().CollectAppStats(gomock.Any()).Return(appStats, nil)
	res := e.Collect(ns, false, nil)
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
	mockAmi.EXPECT().CollectAppStats(gomock.Any()).Return(appStats, nil)
	res = e.Collect(ns, false, nil)
	resNode = res["node"]
	assert.Nil(t, resNode)

	mockAmi.EXPECT().CollectNodeInfo().Return(nodeInfo, nil)
	mockAmi.EXPECT().CollectNodeStats().Return(nil, errors.New("failed to get node stats"))
	mockAmi.EXPECT().CollectAppStats(gomock.Any()).Return(appStats, nil)
	res = e.Collect(ns, false, nil)
	resNodeStats = res["nodestats"]
	assert.Nil(t, resNodeStats)

	mockAmi.EXPECT().CollectNodeInfo().Return(nodeInfo, nil)
	mockAmi.EXPECT().CollectNodeStats().Return(nodeStats, nil)
	mockAmi.EXPECT().CollectAppStats(gomock.Any()).Return(nil, errors.New("failed to get app stats"))
	res = e.Collect(ns, false, nil)
	resApps = res["apps"]
	resAppStats = res["appstats"]
	assert.Equal(t, resApps, []specv1.AppInfo{})
	assert.Nil(t, resAppStats)
}

func TestEngine(t *testing.T) {
	eng, err := NewEngine(config.Config{}, nil, nil, nil)
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
		cfg: config.Config{},
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
		cfg: config.Config{},
		ns:  ns,
		sto: sto,
		syn: mockSync,
		nod: nod,
		log: log.With(log.Any("engine", "test")),
	}
	assert.NotNil(t, eng)
	mockAmi.EXPECT().CollectNodeInfo().Return(nil, nil)
	mockAmi.EXPECT().CollectNodeStats().Return(nil, nil)
	appStats := []specv1.AppStats{{AppInfo: specv1.AppInfo{Name: "app1", Version: "v1"}}, {AppInfo: specv1.AppInfo{Name: "app2", Version: "v2"}}}
	mockAmi.EXPECT().CollectAppStats(gomock.Any()).Return(appStats, nil)

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
	mockSync.EXPECT().SyncApps(gomock.Any()).Return(nil, nil)
	mockAmi.EXPECT().ApplyConfigurations(gomock.Any(), gomock.Any()).Return(nil)
	mockAmi.EXPECT().ApplySecrets(gomock.Any(), gomock.Any()).Return(nil)
	mockAmi.EXPECT().ApplyApplication(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockAmi.EXPECT().DeleteApplication(gomock.Any(), gomock.Any()).Return(nil)
	err = eng.reportAndApply(false, true, nil)
	assert.NoError(t, err)

	// desire app is nil
	mockAmi.EXPECT().CollectNodeInfo().Return(nil, nil)
	mockAmi.EXPECT().CollectNodeStats().Return(nil, nil)
	appStats = []specv1.AppStats{{AppInfo: specv1.AppInfo{Name: "app1", Version: "v1"}}}
	mockAmi.EXPECT().CollectAppStats(gomock.Any()).Return(appStats, nil)
	reApp = specv1.Report{
		"apps": []specv1.AppInfo{{Name: "app1", Version: "v1"}},
	}
	deApp = specv1.Desire{"apps": nil}
	_, err = nod.Report(reApp)
	assert.NoError(t, err)
	_, err = nod.Desire(deApp)
	assert.NoError(t, err)
	err = eng.reportAndApply(false, true, nil)
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
		cfg: config.Config{},
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

func TestInjectEnv(t *testing.T) {
	nod, _, sto := prepare(t)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	ns := "baetyl-edge"
	eng := Engine{
		cfg: config.Config{},
		ns:  ns,
		sto: sto,
		nod: nod,
		log: log.With(log.Any("engine", "test")),
	}

	info := specv1.AppInfo{
		Name:    "app1",
		Version: "v1",
	}

	app := specv1.Application{
		Name:    "app1",
		Version: "v1",
		Services: []specv1.Service{
			{Name: "s0"},
		},
		Volumes: []specv1.Volume{},
	}
	key := makeKey(specv1.KindApplication, "app1", "v1")
	err := sto.Upsert(key, app)
	assert.NoError(t, err)

	err = os.Setenv(context.EnvKeyNodeName, "node01")
	assert.NoError(t, err)

	expApp := &specv1.Application{
		Name:    "app1",
		Version: "v1",
		Services: []specv1.Service{
			{
				Name: "s0",
				Env: []specv1.Environment{
					{
						Name:  context.EnvKeyAppName,
						Value: app.Name,
					},
					{
						Name:  context.EnvKeyServiceName,
						Value: "s0",
					},
					{
						Name:  EnvKeyAppVersion,
						Value: app.Version,
					},
					{
						Name:  context.EnvKeyNodeName,
						Value: "node01",
					},
					{
						Name:  context.EnvKeyCertPath,
						Value: SystemCertPath,
					},
				},
			},
		},
	}

	res, err := eng.injectEnv(info)
	assert.NoError(t, err)
	assert.EqualValues(t, expApp, res)
}

func TestInjectCert(t *testing.T) {
	nod, _, sto := prepare(t)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockSecurity := mock.NewMockSecurity(mockCtl)
	ns := "baetyl-edge"
	eng := Engine{
		cfg: config.Config{},
		ns:  ns,
		sto: sto,
		nod: nod,
		sec: mockSecurity,
		log: log.With(log.Any("engine", "test")),
	}

	app := &specv1.Application{
		Name:      "app1",
		Namespace: "default",
		Version:   "v1",
		Services: []specv1.Service{
			{
				Name:         "s0",
				VolumeMounts: []specv1.VolumeMount{},
			},
			{
				Name:         "s1234567890",
				VolumeMounts: []specv1.VolumeMount{},
			},
		},
		Volumes: []specv1.Volume{},
	}

	cn0 := app.Name + ".s0"
	cn1 := app.Name + ".s1234567890"
	suffix0 := fmt.Sprintf("%x", md5.Sum([]byte(cn0))) + "-s0"
	suffix1 := fmt.Sprintf("%x", md5.Sum([]byte(cn1))) + "-s123456789"

	mockSecurity.EXPECT().GetCA().Return([]byte(caCrt), nil).Times(1)
	mockSecurity.EXPECT().IssueCertificate(gomock.Any(), gomock.Any()).Return(&pki.CertPem{
		Crt: []byte(crt),
		Key: []byte(key),
	}, nil).Times(2)

	secs := map[string]specv1.Secret{}
	err := eng.injectCert(app, secs)
	assert.NoError(t, err)

	expSec := map[string]specv1.Secret{
		SystemCertSecretPrefix + suffix0: {
			Name:      SystemCertSecretPrefix + suffix0,
			Namespace: app.Namespace,
			Labels: map[string]string{
				"baetyl-app-name": app.Name,
				"security-type":   "certificate",
			},
			Data: map[string][]byte{
				"crt.pem": []byte(crt),
				"key.pem": []byte(key),
				"ca.pem":  []byte(caCrt),
			},
			System: app.Namespace == eng.sysns,
		},
		SystemCertSecretPrefix + suffix1: {
			Name:      SystemCertSecretPrefix + suffix1,
			Namespace: app.Namespace,
			Labels: map[string]string{
				"baetyl-app-name": app.Name,
				"security-type":   "certificate",
			},
			Data: map[string][]byte{
				"crt.pem": []byte(crt),
				"key.pem": []byte(key),
				"ca.pem":  []byte(caCrt),
			},
			System: app.Namespace == eng.sysns,
		},
	}

	expApp := &specv1.Application{
		Name:      "app1",
		Namespace: "default",
		Version:   "v1",
		Services: []specv1.Service{
			{
				Name: "s0",
				VolumeMounts: []specv1.VolumeMount{
					{
						Name:      SystemCertVolumePrefix + suffix0,
						MountPath: SystemCertPath,
						ReadOnly:  true,
					},
				},
			},
			{
				Name: "s1234567890",
				VolumeMounts: []specv1.VolumeMount{
					{
						Name:      SystemCertVolumePrefix + suffix1,
						MountPath: SystemCertPath,
						ReadOnly:  true,
					},
				},
			},
		},
		Volumes: []specv1.Volume{
			{
				Name:         SystemCertVolumePrefix + suffix0,
				VolumeSource: specv1.VolumeSource{Secret: &specv1.ObjectReference{Name: SystemCertSecretPrefix + suffix0}},
			},
			{
				Name:         SystemCertVolumePrefix + suffix1,
				VolumeSource: specv1.VolumeSource{Secret: &specv1.ObjectReference{Name: SystemCertSecretPrefix + suffix1}},
			},
		},
	}

	assert.EqualValues(t, expApp, app)
	assert.EqualValues(t, expSec, secs)
	eng.Close()
}
