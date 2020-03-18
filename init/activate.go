package init

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	gohttp "net/http"
	"time"
)

type Event struct {
	Content interface{}
	Trace   string
	Type    string
}

func (init *initialize) activating() error {
	init.Activate()
	t := time.NewTicker(init.cfg.Init.Cloud.Active.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			init.Activate()
		case <-init.tomb.Dying():
			return nil
		}
	}
}

// Report reports info
func (init *initialize) Activate() {
	info := ForwardInfo{
		Status: baetyl.Inspect{
			Time: time.Now(),
		},
		Data: map[string]string{
			common.KeyActivateDataBatchName:      init.batch.Name,
			common.KeyActivateDataBatchNamespace: init.batch.Namespace,
			common.KeyActivateDataSecurityType:   init.batch.SecurityType,
			common.KeyActivateDataSecurityKey:    init.batch.SecurityKey,
		},
	}
	fv, err := init.collect()
	if err != nil {
		init.log.Error("failed to get fingerprint value", log.Error(err))
		return
	}
	info.Data[common.KeyActivateDataFingerprintValue] = fv
	data, err := json.Marshal(info)
	if err != nil {
		init.log.Error("failed to marshal activate info", log.Error(err))
		return
	}

	resp, err := init.sendRequest("POST", init.cfg.Init.Cloud.Active.URL, data)
	if err != nil {
		init.log.Error("failed to send activate data", log.Error(err))
		return
	}
	data, err = http.HandleResponse(resp)
	if err != nil {
		init.log.Error("failed to send activate data", log.Error(err))
		return
	}

	var res BackwardInfo
	err = json.Unmarshal(data, &res)
	if err != nil {
		init.log.Error("error to unmarshal activate response data returned", log.Error(err))
		return
	}

	init.cfg.Node.Name = res.Data[common.KeyActivateResNodeName]
	init.cfg.Node.Namespace = res.Data[common.KeyActivateResNodeNamespace]
	init.cfg.Sync.Cloud.HTTP.CA = res.Data[common.KeyActivateResCA]
	init.cfg.Sync.Cloud.HTTP.Cert = res.Data[common.KeyActivateResCert]
	init.cfg.Sync.Cloud.HTTP.Key = res.Data[common.KeyActivateResKey]
	init.cfg.Sync.Cloud.HTTP.Name = res.Data[common.KeyActivateResName]

	init.sig <- true
}

func (init *initialize) sendRequest(method, path string, body []byte) (*gohttp.Response, error) {
	url := fmt.Sprintf("%s%s", init.cfg.Init.Cloud.HTTP.Address, path)
	r := bytes.NewReader(body)
	req, err := gohttp.NewRequest(method, url, r)
	if err != nil {
		return nil, err
	}
	header := map[string]string{
		"Content-Type": "application/json",
	}
	init.log.Debug("request", log.Any("method", method),
		log.Any("path", path), log.Any("body", body),
		log.Any("header", header))
	req.Header = gohttp.Header{}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	res, err := init.http.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
