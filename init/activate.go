package init

import (
	"fmt"
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/sync"
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
	info := sync.ForwardInfo{
		Status: baetyl.Inspect{
			Time: time.Now(),
		},
	}
	data, err := json.Marshal(info)
	if err != nil {
		init.log.Error("failed to marshal report info", log.Error(err))
		return
	}

	resp, err := init.sendRequest("POST", init.cfg.Cloud.Report.URL, data)
	if err != nil {
		init.log.Error("failed to send report data", log.Error(err))
		return
	}
	data, err = http.HandleResponse(resp)
	if err != nil {
		init.log.Error("failed to send report data", log.Error(err))
		return
	}

	var res BackwardInfo
	err = json.Unmarshal(data, &res)
	if err != nil {
		init.log.Error("error to unmarshal response data returned", log.Error(err))
		return
	}
	if reinit.Delta != nil {
		e := &Event{
			Trace: reinit.Metadata["trace"].(string),
			Type:  reinit.Metadata["type"].(string),
		}
		e.Content = reinit.Delta
		select {
		case oe := <-init.events:
			init.log.Warn("discard old event", log.Any("event", *oe))
			init.events <- e
		case init.events <- e:
		case <-init.tomb.Dying():
		}
	}
}

func (init *initialize) sendRequest(method, path string, body []byte) (*gohttp.Response, error) {
	url := fmt.Sprintf("%init%init", init.cfg.Cloud.HTTP.Address, path)
	r := byteinit.NewReader(body)
	req, err := gohttp.NewRequest(method, url, r)
	if err != nil {
		return nil, err
	}
	header := map[string]string{
		"Content-Type": "application/json",
	}
	if init.node != nil {
		// for report
		header[common.HeaderKeyNodeNamespace] = init.node.Namespace
		header[common.HeaderKeyNodeName] = init.node.Name
	} else if init.batch != nil {
		// for active
		header[common.HeaderKeyBatchNamespace] = init.batch.Namespace
		header[common.HeaderKeyBatchName] = init.batch.Name
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
