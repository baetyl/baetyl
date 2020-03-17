package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	gohttp "net/http"
	"time"

	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
)

type Event struct {
	Content interface{}
	Trace   string
	Type    string
}

func (s *sync) reporting() error {
	s.Report()
	t := time.NewTicker(s.cfg.Cloud.Report.Interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			s.Report()
		case <-s.tomb.Dying():
			return nil
		}
	}
}

// Report reports info
func (s *sync) Report() {
	var apps AppsVersionResource
	err := s.store.Get(common.DefaultAppsKey, &apps)
	if err != nil {
		s.log.Error("failed to get local apps info", log.Error(err))
		return
	}
	info := ForwardInfo{
		Apps: apps.Value,
		Status: baetyl.Inspect{
			Time: time.Now(),
		},
	}
	data, err := json.Marshal(info)
	if err != nil {
		s.log.Error("failed to marshal report info", log.Error(err))
		return
	}

	resp, err := s.sendRequest("POST", s.cfg.Cloud.Report.URL, data)
	if err != nil {
		s.log.Error("failed to send report data", log.Error(err))
		return
	}
	data, err = http.HandleResponse(resp)
	if err != nil {
		s.log.Error("failed to send report data", log.Error(err))
		return
	}

	var res BackwardInfo
	err = json.Unmarshal(data, &res)
	if err != nil {
		s.log.Error("error to unmarshal response data returned", log.Error(err))
		return
	}
	if res.Delta != nil {
		e := &Event{
			Trace: res.Metadata["trace"].(string),
			Type:  res.Metadata["type"].(string),
		}
		e.Content = res.Delta
		select {
		case oe := <-s.events:
			s.log.Warn("discard old event", log.Any("event", *oe))
			s.events <- e
		case s.events <- e:
		case <-s.tomb.Dying():
		}
	}
}

func (s *sync) sendRequest(method, path string, body []byte) (*gohttp.Response, error) {
	url := fmt.Sprintf("%s%s", s.cfg.Cloud.HTTP.Address, path)
	r := bytes.NewReader(body)
	req, err := gohttp.NewRequest(method, url, r)
	if err != nil {
		return nil, err
	}
	header := map[string]string{
		"Content-Type": "application/json",
	}
	if s.node != nil {
		// for report
		header[common.HeaderKeyNodeNamespace] = s.node.Namespace
		header[common.HeaderKeyNodeName] = s.node.Name
	} else if s.batch != nil {
		// for active
		header[common.HeaderKeyBatchNamespace] = s.batch.Namespace
		header[common.HeaderKeyBatchName] = s.batch.Name
	}
	s.log.Debug("request", log.Any("method", method),
		log.Any("path", path), log.Any("body", body),
		log.Any("header", header))
	req.Header = gohttp.Header{}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	res, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
