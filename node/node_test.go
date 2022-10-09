package node

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"

	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	routing "github.com/qiangxue/fasthttp-routing"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/baetyl/baetyl/v2/store"
	"github.com/baetyl/baetyl/v2/utils"
)

func TestNodeShadow(t *testing.T) {
	f, err := os.CreateTemp("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	s, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	ss, err := NewNode(s)
	assert.NoError(t, err)
	assert.NotNil(t, ss)

	// ! test sequence is important
	tests := []struct {
		name         string
		desired      string
		reported     string
		desireDelta  string
		reportDelta  string
		desireStored string
		reportStored string
		desireErr    error
		reportErr    error
	}{
		{
			name:         "1",
			desired:      "{}",
			reported:     "{}",
			desireDelta:  `{"core": null}`,
			reportDelta:  `{"core": null}`,
			desireStored: "{}",
			reportStored: `{}`,
		},
		{
			name:         "2",
			desired:      `{"name": "module", "version": "45"}`,
			reported:     `{"name": "module", "version": "43"}`,
			desireDelta:  `{"name": "module", "version": "45", "core": null}`,
			reportDelta:  `{"version": "45", "core": null}`,
			desireStored: `{"name": "module", "version": "45"}`,
			reportStored: `{"name": "module", "version": "43"}`,
		},
		{
			name:         "3",
			desired:      `{"name": "module", "module": {"image": "test:v2"}}`,
			reported:     `{"name": "module", "module": {"image": "test:v1"}}`,
			desireDelta:  `{"version": "45", "module": {"image": "test:v2"}, "core": null}`,
			reportDelta:  `{"version": "45", "module": {"image": "test:v2"}, "core": null}`,
			desireStored: `{"name": "module", "version": "45", "module": {"image": "test:v2"}}`,
			reportStored: `{"name": "module", "version": "43", "module": {"image": "test:v1"}}`,
		},
		{
			name:         "4",
			desired:      `{"module": {"image": "test:v2", "array": []}}`,
			reported:     `{"module": {"image": "test:v1", "object": {"attr": "value"}}}`,
			desireDelta:  `{"version": "45", "module": {"image": "test:v2", "array": []}, "core": null}`,
			reportDelta:  `{"version": "45", "module": {"image": "test:v2", "array": [], "object": null}, "core": null}`,
			desireStored: `{"name": "module", "version": "45", "module": {"image": "test:v2", "array": []}}`,
			reportStored: `{"name": "module", "version": "43", "module": {"image": "test:v1", "object": {"attr": "value"}}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var desired, desireStored v1.Desire
			var reported, reportStored v1.Report
			var desireDelta, reportDelta v1.Delta
			assert.NoError(t, json.Unmarshal([]byte(tt.desired), &desired))
			assert.NoError(t, json.Unmarshal([]byte(tt.reported), &reported))
			assert.NoError(t, json.Unmarshal([]byte(tt.desireDelta), &desireDelta))
			assert.NoError(t, json.Unmarshal([]byte(tt.reportDelta), &reportDelta))
			assert.NoError(t, json.Unmarshal([]byte(tt.desireStored), &desireStored))
			assert.NoError(t, json.Unmarshal([]byte(tt.reportStored), &reportStored))

			gotDelta, err := ss.Desire(desired, false)
			assert.Equal(t, tt.desireErr, err)
			if !reflect.DeepEqual(gotDelta, desireDelta) {
				t.Errorf("Node.Desire() = %v, want %v", gotDelta, desireDelta)
			}
			gotDelta, err = ss.Report(reported, false)
			assert.Equal(t, tt.reportErr, err)
			if !reflect.DeepEqual(gotDelta, reportDelta) {
				t.Errorf("Node.Report() = %v, want %v", gotDelta, reportDelta)
			}

			actual, err := ss.Get()
			assert.NoError(t, err)
			if actual.Desire == nil {
				assert.Empty(t, desireStored)
			} else {
				if !reflect.DeepEqual(actual.Desire, desireStored) {
					t.Errorf("Node.Get().Desire = %v, want %v", actual.Desire, desireStored)
				}
			}
			if actual.Report == nil {
				assert.Empty(t, reportStored)
			} else {
				delete(actual.Report, "core")
				if !reflect.DeepEqual(actual.Report, reportStored) {
					t.Errorf("Node.Get().Report = %v, want %v", actual.Report, reportStored)
				}
			}
		})
	}
}

func TestShadowRenew(t *testing.T) {
	f, err := os.CreateTemp("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())
	defer os.RemoveAll(f.Name())

	s, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	ss, err := NewNode(s)
	assert.NoError(t, err)
	assert.NotNil(t, ss)

	desire := v1.Desire{"apps": map[string]interface{}{"app1": "123", "app2": "234", "app3": "345", "app4": "456", "app5": ""}}
	delta, err := ss.Desire(desire, false)
	assert.NoError(t, err)
	apps := delta["apps"].(map[string]interface{})
	assert.Len(t, apps, 5)
	assert.Equal(t, "123", apps["app1"])
	assert.Equal(t, "234", apps["app2"])
	assert.Equal(t, "345", apps["app3"])
	assert.Equal(t, "456", apps["app4"])
	assert.Equal(t, "", apps["app5"])

	report := v1.Report{"apps": map[string]interface{}{"app1": "123", "app2": "235", "app3": "", "app5": "567", "app6": "678"}}
	delta, err = ss.Report(report, false)
	assert.NoError(t, err)
	apps = delta["apps"].(map[string]interface{})
	assert.Len(t, apps, 5)
	assert.Equal(t, "234", apps["app2"])
	assert.Equal(t, "345", apps["app3"])
	assert.Equal(t, "456", apps["app4"])
	assert.Equal(t, "", apps["app5"])
	assert.Equal(t, nil, apps["app6"])

	ss, err = NewNode(s)
	assert.NoError(t, err)
	assert.NotNil(t, ss)

	delta, err = ss.Report(report, false)
	assert.NoError(t, err)
	apps = delta["apps"].(map[string]interface{})
	assert.Len(t, apps, 5)
	assert.Equal(t, "234", apps["app2"])
	assert.Equal(t, "345", apps["app3"])
	assert.Equal(t, "456", apps["app4"])
	assert.Equal(t, "", apps["app5"])
	assert.Equal(t, nil, apps["app6"])
}

func TestGetStats(t *testing.T) {
	f, err := os.CreateTemp("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())
	defer os.RemoveAll(f.Name())

	s, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	ss, err := NewNode(s)
	assert.NoError(t, err)
	assert.NotNil(t, ss)

	router := routing.New()
	router.Get("/node/stats", utils.Wrapper(ss.GetStats))
	go fasthttp.ListenAndServe(":50020", router.HandleRequest)
	time.Sleep(100 * time.Millisecond)

	client := &fasthttp.Client{}
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	url := fmt.Sprintf("%s%s", "http://127.0.0.1:50020", "/node/stats")
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	err = client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 200)

	appInfo := []v1.AppInfo{
		{
			Name:    "app1",
			Version: "version1",
		},
	}
	appStats := []v1.AppStats{
		{
			AppInfo: appInfo[0],
		},
	}
	sysappInfo := []v1.AppInfo{
		{
			Name:    "baetyl1",
			Version: "version1",
		},
	}
	sysappStats := []v1.AppStats{
		{
			AppInfo: sysappInfo[0],
		},
	}
	core := &v1.CoreInfo{
		GoVersion: runtime.Version(),
	}
	reportView := v1.ReportView{
		Apps:        appInfo,
		SysApps:     sysappInfo,
		Core:        core,
		AppStats:    appStats,
		SysAppStats: sysappStats,
		Node: map[string]*v1.NodeInfo{
			"hostname": {
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
				Role:             "master",
			},
		},
		NodeStats: map[string]*v1.NodeStats{
			"hostname": {},
		},
	}
	nodeView := v1.NodeView{
		Report: &reportView,
	}
	report := v1.Report{
		"core": core,
	}
	report.SetAppInfos(false, appInfo)
	report.SetAppInfos(true, sysappInfo)
	report.SetAppStats(false, appStats)
	report.SetAppStats(true, sysappStats)
	_, err = ss.Report(report, false)
	assert.NoError(t, err)

	req2 := fasthttp.AcquireRequest()
	resp2 := fasthttp.AcquireResponse()
	req2.SetRequestURI(url)
	req2.Header.SetMethod("GET")
	err = client.Do(req2, resp2)
	assert.NoError(t, err)
	assert.Equal(t, resp2.StatusCode(), 200)
	// time unequal
	var respNodeView v1.NodeView
	json.Unmarshal(resp2.Body(), &respNodeView)

	respNodeView.Report.Time = nodeView.Report.Time
	respNodeView.CreationTimestamp = nodeView.CreationTimestamp
	assert.Equal(t, http.StatusOK, resp2.StatusCode())
	assertEqualNodeView(t, nodeView.Report, respNodeView.Report)
}

func TestGetNodeProperties(t *testing.T) {
	f, err := os.CreateTemp("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())
	defer os.RemoveAll(f.Name())

	s, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	ss, err := NewNode(s)
	assert.NoError(t, err)
	assert.NotNil(t, ss)

	router := routing.New()
	router.Get("/node/properties", utils.Wrapper(ss.GetNodeProperties))
	go fasthttp.ListenAndServe(":50021", router.HandleRequest)
	time.Sleep(100 * time.Millisecond)

	client := &fasthttp.Client{}
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	url := fmt.Sprintf("%s%s", "http://127.0.0.1:50021", "/node/properties")
	req.SetRequestURI(url)
	req.Header.SetMethod("GET")
	err = client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 200)

	desire := v1.Desire{
		"nodeprops": map[string]interface{}{
			"a": "1",
			"b": "2",
		},
	}
	_, err = ss.Desire(desire, true)
	assert.NoError(t, err)
	report := v1.Report{
		"nodeprops": map[string]interface{}{
			"a": "1",
			"b": "3",
		},
	}
	_, err = ss.Report(report, true)
	assert.NoError(t, err)

	req2 := fasthttp.AcquireRequest()
	resp2 := fasthttp.AcquireResponse()
	req2.SetRequestURI(url)
	req2.Header.SetMethod("GET")
	err = client.Do(req2, resp2)
	assert.NoError(t, err)
	assert.Equal(t, resp2.StatusCode(), 200)
	// time unequal
	var respNodeProps map[string]interface{}
	json.Unmarshal(resp2.Body(), &respNodeProps)

	expect := map[string]interface{}{
		"report": report[KeyNodeProps],
		"desire": desire[KeyNodeProps],
	}
	assert.Equal(t, http.StatusOK, resp2.StatusCode())
	assert.Equal(t, expect, respNodeProps)
}

func TestUpdateNodeProperties(t *testing.T) {
	f, err := os.CreateTemp("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())
	defer os.RemoveAll(f.Name())

	s, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, s)

	ss, err := NewNode(s)
	assert.NoError(t, err)
	assert.NotNil(t, ss)

	router := routing.New()
	router.Put("/node/properties", utils.Wrapper(ss.UpdateNodeProperties))
	go fasthttp.ListenAndServe(":50022", router.HandleRequest)
	time.Sleep(100 * time.Millisecond)

	delta := map[string]interface{}{}
	data, _ := json.Marshal(delta)
	client := &fasthttp.Client{}
	req := fasthttp.AcquireRequest()
	req.SetBody(data)
	resp := fasthttp.AcquireResponse()
	url := fmt.Sprintf("%s%s", "http://127.0.0.1:50022", "/node/properties")
	req.SetRequestURI(url)
	req.Header.SetMethod("PUT")
	err = client.Do(req, resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode(), 200)

	report := v1.Report{
		KeyNodeProps: map[string]interface{}{
			"a": "1",
		},
	}
	delta = map[string]interface{}{
		"a": "2",
		"b": "3",
	}
	expect := map[string]interface{}{
		"a": "2",
		"b": "3",
	}
	_, err = ss.Report(report, true)
	assert.NoError(t, err)

	data, _ = json.Marshal(delta)
	req2 := fasthttp.AcquireRequest()
	resp2 := fasthttp.AcquireResponse()
	req2.SetRequestURI(url)
	req2.SetBody(data)
	req2.Header.SetMethod("PUT")
	err = client.Do(req2, resp2)
	assert.NoError(t, err)
	assert.Equal(t, resp2.StatusCode(), 200)
	// time unequal
	var respReport map[string]interface{}
	json.Unmarshal(resp2.Body(), &respReport)

	assert.Equal(t, http.StatusOK, resp2.StatusCode())
	assert.Equal(t, respReport, expect)
	shadow, err := ss.Get()
	assert.NoError(t, err)
	assert.Equal(t, shadow.Report[KeyNodeProps], expect)
}

func assertEqualNodeView(t *testing.T, view1 *v1.ReportView, view2 *v1.ReportView) {
	assert.Equal(t, view1.AppStats, view2.AppStats)
	assert.Equal(t, view1.Apps, view2.Apps)
	assert.Equal(t, view1.SysApps, view2.SysApps)
	assert.Equal(t, view1.SysAppStats, view2.SysAppStats)
	assert.Equal(t, view1.Core, view2.Core)
}
