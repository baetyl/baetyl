package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	"github.com/baetyl/baetyl/baetyl-agent/config"
	baetylHttp "github.com/baetyl/baetyl/protocol/http"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/stretchr/testify/assert"
)

func TestReport(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	cwd, _ := os.Getwd()
	os.Chdir(dir)
	confString := `name: app
app_version: v1`
	containerDir := path.Join(baetyl.DefaultDBDir, "volumes")
	os.MkdirAll(containerDir, 0755)
	filePath := path.Join(containerDir, baetyl.AppConfFileName)
	ioutil.WriteFile(filePath, []byte(confString), 0755)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := config.BackwardInfo{
			Delta: map[string]interface{}{
				"deploy": "v1",
			},
			Metadata: map[string]interface{}{
				"trace": "TRACE",
				"type":  "APP",
			},
		}
		data, _ := json.Marshal(info)
		w.Write(data)
	}))
	defer ts.Close()
	a := initAgent(t)
	a.cfg.Remote.HTTP.Address = ts.URL
	a.http, _ = baetylHttp.NewClient(*a.cfg.Remote.HTTP)
	a.report()
	e := <-a.events
	le := &EventLink{
		Trace: "TRACE",
		Type:  "APP",
		Info: map[string]interface{}{
			"deploy": "v1",
		},
	}
	assert.Equal(t, le, e.Content)
	os.Chdir(cwd)
}
