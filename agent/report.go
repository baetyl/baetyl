package agent

import (
	"encoding/json"
	"github.com/baetyl/baetyl-go/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"

	"github.com/baetyl/baetyl-core/common"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/utils"
)

type EventLink struct {
	Info  map[string]interface{}
	Trace string
	Type  string
}

func (a *Agent) Report() error {
	t := time.NewTicker(a.cfg.Remote.Report.Interval)
	for {
		select {
		case <-t.C:
			a.report()
		case <-a.tomb.Dying():
			return nil
		}
	}
}

// Report reports info
func (a *Agent) report() {
	defer utils.Trace("report", logger.Debugf)()

	// TODO get pod and node info from api server
	desire, ok := a.shadow.Desired.(map[string]interface{})
	if !ok {
		a.log.Error("shadow desire format error")
	}
	report, ok := a.shadow.Reported.(map[string]interface{})
	if !ok {
		a.log.Error("shadow report format error")
	}
	desireApps, ok := report["apps"].(map[string]string)
	if !ok {
		a.log.Error("shadow desire does not have apps info or format error")
	}
	reportApps, ok := desire["apps"].(map[string]string)
	if !ok {
		a.log.Error("shadow report does not have apps info or format error")
	}
	for name, ver := range desireApps {
		d, err := a.impl.Get(name, metav1.GetOptions{})
		if err != nil {
			a.log.Error("failed to get deployment", log.Any("name", name), log.Error(err))
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
		a.log.Error("failed to marshal report info", log.Error(err))
		return
	}
	var res config.BackwardInfo
	resData, err := a.sendRequest("POST", a.cfg.Remote.Report.URL, req)
	if err != nil {
		a.log.Error("failed to send report data", log.Error(err))
		return
	}
	err = json.Unmarshal(resData, &res)
	if err != nil {
		a.log.Error("error to unmarshal response data returned", log.Error(err))
		return
	}
	if a.node == nil {
		a.node = &node{
			Name:      res.Metadata[string(common.Node)].(string),
			Namespace: res.Metadata[common.KeyContextNamespace].(string),
		}
	}
	if res.Delta != nil {
		le := &EventLink{
			Trace: res.Metadata["trace"].(string),
			Type:  res.Metadata["type"].(string),
		}
		le.Info = res.Delta
		e := &Event{
			Time:    time.Time{},
			Type:    EventType(le.Type),
			Content: le,
		}
		select {
		case oe := <-a.events:
			a.log.Warn("discard old event", log.Any("old event", *oe))
			a.events <- e
		case a.events <- e:
		}
	}
}

func (a *Agent) sendRequest(method, path string, body []byte) ([]byte, error) {
	header := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	if a.node != nil {
		// for report
		header[common.HeaderKeyNodeNamespace] = a.shadow.Namespace
		header[common.HeaderKeyNodeName] = a.shadow.Name
	} else if a.batch != nil {
		// for active
		header[common.HeaderKeyBatchNamespace] = a.batch.Namespace
		header[common.HeaderKeyBatchName] = a.batch.Name
	}
	a.log.Debug("request", log.Any("method", method),
		log.Any("path", path), log.Any("body", body),
		log.Any("header", header))
	return a.http.SendPath(method, path, body, header)
}
