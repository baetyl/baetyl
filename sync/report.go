package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-go/link"
	gohttp "net/http"
	"time"

	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
)

// Report reports info
func (s *sync) Report(msg link.Message) error {
	var report map[string]interface{}
	err := json.Unmarshal(msg.Content, &report)
	apps, ok := report["apps"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("apps does not exist")
	}
	info := ForwardInfo{
		Apps: apps,
		Status: baetyl.Inspect{
			Time: time.Now(),
		},
	}
	data, err := json.Marshal(info)
	if err != nil {
		s.log.Error("failed to marshal report info", log.Error(err))
		return err
	}
	resp, err := s.sendRequest("POST", s.cfg.Cloud.Report.URL, data)
	if err != nil {
		s.log.Error("failed to send report data", log.Error(err))
		return err
	}
	data, err = http.HandleResponse(resp)
	if err != nil {
		s.log.Error("failed to send report data", log.Error(err))
		return err
	}
	var res BackwardInfo
	err = json.Unmarshal(data, &res)
	if err != nil {
		s.log.Error("error to unmarshal response data returned", log.Error(err))
		return err
	}
	if res.Delta != nil {
		_, err = s.shadow.Desire(res.Delta)
		if err != nil {
			return err
		}
	}
	return nil
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
	header[common.HeaderKeyNodeNamespace] = s.node.Namespace
	header[common.HeaderKeyNodeName] = s.node.Name
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
