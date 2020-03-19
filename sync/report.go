package sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/link"
	"github.com/baetyl/baetyl-go/log"
	gohttp "net/http"
)

// Report reports info
func (s *sync) Report(msg link.Message) error {
	resp, err := s.sendRequest("POST", s.cfg.Cloud.Report.URL, msg.Content)
	if err != nil {
		s.log.Error("failed to send report data", log.Error(err))
		return err
	}
	data, err := http.HandleResponse(resp)
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
		content, err := json.Marshal(res.Delta)
		if err != nil {
			return err
		}
		_, err = s.shadow.Desire(res.Delta)
		if err != nil {
			return err
		}
		msg := &link.Message{
			Context: link.Context{Topic: common.EngineAppEvent},
			Content: content,
		}
		err = s.cent.Trigger(msg)
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
