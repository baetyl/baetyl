package sync

import (
	"encoding/json"
	"github.com/256dpi/gomqtt/packet"
	"github.com/baetyl/baetyl-go/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/utils"
)

type Event struct {
	Content interface{}
	Trace   string
	Type    string
}

func (s *sync) reporting() error {
	t := time.NewTicker(s.cfg.Remote.Report.Interval)
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
	defer utils.Trace("report", logger.Debugf)()

	// TODO get pod and node info from api server
	desire, ok := s.shadow.Desired.(map[string]interface{})
	if !ok {
		s.log.Error("shadow desire format error")
	}
	report, ok := s.shadow.Reported.(map[string]interface{})
	if !ok {
		s.log.Error("shadow report format error")
	}
	desireApps, ok := report["apps"].(map[string]string)
	if !ok {
		s.log.Error("shadow desire does not have apps info or format error")
	}
	reportApps, ok := desire["apps"].(map[string]string)
	if !ok {
		s.log.Error("shadow report does not have apps info or format error")
	}
	for name, ver := range desireApps {
		d, err := s.impl.Get(name, metav1.GetOptions{})
		if err != nil {
			s.log.Error("failed to get deployment", log.Any("name", name), log.Error(err))
		}
		if ver != d.ResourceVersion {
			reportApps[name] = d.ResourceVersion
		}
	}
	// TODO update local shadow report

	info := config.ForwardInfo{
		Apps: reportApps,
	}
	req, err := json.Marshal(info)
	if err != nil {
		s.log.Error("failed to marshal report info", log.Error(err))
		return
	}
	var res config.BackwardInfo
	resData, err := s.sendRequest("POST", s.cfg.Remote.Report.URL, req)
	if err != nil {
		s.log.Error("failed to send report data", log.Error(err))
		return
	}
	err = json.Unmarshal(resData, &res)
	if err != nil {
		s.log.Error("error to unmarshal response data returned", log.Error(err))
		return
	}
	if s.node == nil {
		s.node = &node{
			Name:      res.Metadata[string(common.Node)].(string),
			Namespace: res.Metadata[common.KeyContextNamespace].(string),
		}
	}
	if res.Delta != nil {
		e := &Event{
			Trace: res.Metadata["trace"].(string),
			Type:  res.Metadata["type"].(string),
		}
		e.Content = res.Delta
		pkt := packet.NewPublish()
		pkt.Message.Topic = common.InternalEventTopic
		pkt.Message.QOS = packet.QOS(1)
		payload, err := json.Marshal(e)
		if err != nil {
			s.log.Error("failed marshal payload", log.Any("payload", payload))
			return
		}
		pkt.Message.Payload = payload
		err = s.mqtt.Send(pkt)
		if err != nil {
			s.log.Error("failed to send mqtt msg", log.Any("packet", pkt))
		}
	}
}

func (s *sync) sendRequest(method, path string, body []byte) ([]byte, error) {
	header := map[string]string{
		"Content-Type": "application/json",
	}
	if s.node != nil {
		// for report
		header[common.HeaderKeyNodeNamespace] = s.shadow.Namespace
		header[common.HeaderKeyNodeName] = s.shadow.Name
	} else if s.batch != nil {
		// for active
		header[common.HeaderKeyBatchNamespace] = s.batch.Namespace
		header[common.HeaderKeyBatchName] = s.batch.Name
	}
	s.log.Debug("request", log.Any("method", method),
		log.Any("path", path), log.Any("body", body),
		log.Any("header", header))
	return s.http.SendPath(method, path, body, header)
}
