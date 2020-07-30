package engine

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/mock"
	"github.com/golang/mock/gomock"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestEngine_CollectNodeReport(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockAmi := mock.NewMockAMI(mockCtl)
	e := Engine{
		ami:   mockAmi,
		sto:   nil,
		nod:   nil,
		cfg:   config.Config{},
		ns:    "baetyl-edge",
		sysns: "baetyl-edge-system",
		log:   log.With(log.Any("engine", "any")),
	}
	assert.NotNil(t, e)
	router := routing.New()
	router.Get("/node/stats", e.CollectReport)
	go fasthttp.ListenAndServe(":50031", router.HandleRequest)
	time.Sleep(100 * time.Millisecond)
	client := &fasthttp.Client{}
	appStats := []specv1.AppStats{
		{
			AppInfo: specv1.AppInfo{Name: "app1", Version: "v1"},
		},
		{
			AppInfo: specv1.AppInfo{Name: "app2", Version: "v1"},
		},
	}
	nodeStats := &specv1.NodeStats{}
	nodeInfo := &specv1.NodeInfo{
		Hostname:         "hostname",
		Address:          "nodeip",
		Arch:             "arch",
		KernelVersion:    "kernel",
		OS:               "os",
		ContainerRuntime: "runtime",
		MachineID:        "machine",
		BootID:           "boot",
		SystemUUID:       "system",
		OSImage:          "image",
	}
	apps := make([]specv1.AppInfo, 0)
	for _, info := range appStats {
		app := specv1.AppInfo{
			Name:    info.Name,
			Version: info.Version,
		}
		apps = append(apps, app)
	}
	r := specv1.Report{
		"time":      time.Now(),
		"node":      nodeInfo,
		"nodestats": nodeStats,
		"core": specv1.CoreInfo{
			GoVersion:   runtime.Version(),
			BinVersion:  utils.VERSION,
			GitRevision: utils.REVISION,
		},
	}
	r.SetAppInfos(false, apps)
	r.SetAppStats(false, appStats)
	r.SetAppInfos(true, apps)
	r.SetAppStats(true, appStats)

	mockAmi.EXPECT().CollectAppStats(e.ns).Return(appStats, nil).Times(1)
	mockAmi.EXPECT().CollectAppStats(e.sysns).Return(appStats, nil).Times(1)
	mockAmi.EXPECT().CollectNodeStats().Return(nodeStats, nil).Times(1)
	mockAmi.EXPECT().CollectNodeInfo().Return(nodeInfo, nil).Times(1)

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	url := fmt.Sprintf("%s%s", "http://127.0.0.1:50031", "/node/stats")
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	err := client.Do(req, resp)
	assert.NoError(t, err)
	// time unequal
	var rMap, respMap map[string]interface{}
	data, err := json.Marshal(r)
	assert.NoError(t, err)
	err = json.Unmarshal(data, &rMap)
	assert.NoError(t, err)
	err = json.Unmarshal(resp.Body(), &respMap)
	assert.NoError(t, err)
	delete(respMap, "time")
	delete(rMap, "time")
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, true, reflect.DeepEqual(rMap, respMap))
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
