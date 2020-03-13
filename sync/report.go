package sync

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/mqtt"
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
	var apps map[string]string
	err := s.store.Get(common.DefaultAppsKey, &apps)
	if err != nil {
		s.log.Error("failed to get local apps info", log.Error(err))
		return
	}
	info := config.ForwardInfo{
		Apps: apps,
		// Status: &baetyl.Inspect{
		// 	Time: time.Now(),
		// },
	}
	data, err := json.Marshal(info)
	if err != nil {
		s.log.Error("failed to marshal report info", log.Error(err))
		return
	}

	resp, err := s.http.Post(s.cfg.Cloud.Report.URL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		s.log.Error("failed to send report data", log.Error(err))
		return
	}
	data, err = http.HandleResponse(resp)
	if err != nil {
		s.log.Error("failed to send report data", log.Error(err))
		return
	}

	var res config.BackwardInfo
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
		pkt := mqtt.NewPublish()
		pkt.Message.Topic = common.InternalEventTopic
		pkt.Message.QOS = mqtt.QOS(1)
		payload, err := json.Marshal(e)
		if err != nil {
			s.log.Error("failed marshal payload", log.Any("payload", payload))
			return
		}
		pkt.Message.Payload = payload
		err = s.mqtt.Send(pkt)
		if err != nil {
			s.log.Error("failed to send mqtt msg", log.Any("mqtt", pkt))
		}
	}
}

// func (s *sync) sendRequest(method, path string, body []byte) ([]byte, error) {
// 	header := map[string]string{
// 		"Content-Type": "application/json",
// 	}
// 	if s.node != nil {
// 		// for report
// 		header[common.HeaderKeyNodeNamespace] = s.node.Namespace
// 		header[common.HeaderKeyNodeName] = s.node.Name
// 	} else if s.batch != nil {
// 		// for active
// 		header[common.HeaderKeyBatchNamespace] = s.batch.Namespace
// 		header[common.HeaderKeyBatchName] = s.batch.Name
// 	}
// 	s.log.Debug("request", log.Any("method", method),
// 		log.Any("path", path), log.Any("body", body),
// 		log.Any("header", header))
// 	return s.http.Post(method, path, body, header)
// }
