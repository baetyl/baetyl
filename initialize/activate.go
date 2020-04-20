package initialize

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-go/utils"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/spec/v1"
)

func (init *Initialize) activating() error {
	init.activate()
	t := time.NewTicker(init.cfg.Init.Cloud.Active.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			init.activate()
		case <-init.tomb.Dying():
			return nil
		}
	}
}

// Report reports info
func (init *Initialize) activate() {
	info := v1.ActiveRequest{
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
	init.log.Debug("init", log.Any("info data", string(data)))

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
	var res v1.ActiveResponse
	err = json.Unmarshal(data, &res)
	if err != nil {
		init.log.Error("error to unmarshal activate response data returned", log.Error(err))
		return
	}

	if err := init.genCert(res.Certificate); err != nil {
		init.log.Error("error to create cert file", log.Error(err))
		return
	}

	init.sig <- true
}

func (init *Initialize) genCert(c utils.Certificate) error {
	if err := ioutil.WriteFile(init.cfg.Sync.Cloud.HTTP.CA, []byte(c.CA), 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(init.cfg.Sync.Cloud.HTTP.Cert, []byte(c.Cert), 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(init.cfg.Sync.Cloud.HTTP.Key, []byte(c.Key), 0755); err != nil {
		return err
	}
	return nil
}
