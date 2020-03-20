package initialize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/spec/api"
)

func (init *Initialize) activating() error {
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
func (init *Initialize) Activate() {
	info := api.ActiveRequest{
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
	if fv == "" {
		init.log.Error("fingerprint value is null", log.Error(err))
		return
	}
	info.FingerprintValue = fv
	data, err := json.Marshal(info)
	if err != nil {
		init.log.Error("failed to marshal activate info", log.Error(err))
		return
	}

	url := fmt.Sprintf("%s%s", init.cfg.Init.Cloud.HTTP.Address, init.cfg.Init.Cloud.Active.URL)
	resp, err := init.http.Post(url, "application/json", bytes.NewReader(data))

	if err != nil {
		init.log.Error("failed to send activate data", log.Error(err))
		return
	}
	data, err = http.HandleResponse(resp)
	if err != nil {
		init.log.Error("failed to send activate data", log.Error(err))
		return
	}
	var res api.ActiveResponse
	err = json.Unmarshal(data, &res)
	if err != nil {
		init.log.Error("error to unmarshal activate response data returned", log.Error(err))
		return
	}

	init.cfg.Sync.Node.Name = res.NodeName
	init.cfg.Sync.Node.Namespace = res.Namespace
	init.cfg.Sync.Cloud.HTTP.CA = res.Certificate.CA
	init.cfg.Sync.Cloud.HTTP.Cert = res.Certificate.Cert
	init.cfg.Sync.Cloud.HTTP.Key = res.Certificate.Key
	init.cfg.Sync.Cloud.HTTP.Name = res.Certificate.Name

	init.sig <- true
}
