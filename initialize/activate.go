package initialize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	gohttp "net/http"
	"time"
)

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
		BatchName:     init.batch.name,
		Namespace:     init.batch.namespace,
		SecurityType:  init.batch.securityType,
		SecurityValue: init.batch.securityKey,
		PenetrateData: init.attrs,
	}
	fv, err := init.collect()
	if err != nil {
		init.log.Error("failed to get fingerprint value", log.Error(err))
		return
	}
	info.FingerprintValue = fv
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

	init.cfg.Node.Name = res.NodeName
	init.cfg.Node.Namespace = res.Namespace
	init.cfg.Sync.Cloud.HTTP.CA = res.Cert.CA
	init.cfg.Sync.Cloud.HTTP.Cert = res.Cert.Cert
	init.cfg.Sync.Cloud.HTTP.Key = res.Cert.Key
	init.cfg.Sync.Cloud.HTTP.Name = res.Cert.Name

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
