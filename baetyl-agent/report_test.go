package main

import (
	"encoding/json"
	"github.com/baetyl/baetyl/baetyl-agent/common"
	"github.com/baetyl/baetyl/baetyl-agent/config"
	baetylHttp "github.com/baetyl/baetyl/protocol/http"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"
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

func TestCollector(t *testing.T) {
	proofs := map[common.Proof]string{
		common.HostID: "host",
		common.MAC:    "mac",
		common.SN:     "error path",
		common.CPU:    attrs["fingerprintValue"],
		"error":       "",
	}

	tmpDir, err := ioutil.TempDir("", "")
	assert.Nil(t, err)
	snPath := path.Join(tmpDir, "sn")
	err = ioutil.WriteFile(snPath, []byte(proofs[common.SN]), 0755)
	assert.Nil(t, err)
	defer os.RemoveAll(tmpDir)

	i := &baetyl.Inspect{
		Software: baetyl.Software{},
		Hardware: baetyl.Hardware{
			HostInfo: &utils.HostInfo{
				HostID: proofs[common.HostID],
			},
			NetInfo: &utils.NetInfo{Interfaces: []utils.Interface{
				{Name: "eh0", MAC: proofs[common.MAC]},
			}},
		},
	}
	a := initAgent(t)
	a.attrs = attrs
	expectAct := &config.Activation{
		PenetrateData: a.attrs,
	}

	for k, v := range proofs {
		t.Log(k)
		a.cfg.Fingerprints = []config.Fingerprint{
			{Proof: k},
		}
		if k == common.MAC {
			a.cfg.Fingerprints[0].Value = "eh0"
		}
		if k == common.SN {
			a.cfg.Fingerprints[0].Value = snPath
		}
		expectAct.FingerprintValue = v
		act, err := a.collectActiveInfo(i)
		if k == common.SN {
			assert.Error(t, err)
		} else if k == "error" {
			assert.Error(t, err, "proof invalid")
		} else {
			assert.Nil(t, err)
			assert.EqualValues(t, expectAct, act)
		}
	}
}
