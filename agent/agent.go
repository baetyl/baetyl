package agent

import (
	"fmt"
	"io/ioutil"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/config"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/trans/http"
	"github.com/baidu/openedge/trans/mqtt"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
)

// topics from cloud
const (
	CloudForward  = "$baidu/iot/edge/%s/core/forward"
	CloudBackward = "$baidu/iot/edge/%s/core/backward"
)

// Agent agent of edge cloud
type Agent struct {
	conf config.Cloud
	http *http.Client
	mqtt *mqtt.Dispatcher
	log  *logrus.Entry
}

// NewAgent creates a new agent
func NewAgent(c config.Cloud) (*Agent, error) {
	httpcli, err := http.NewClient(c.OpenAPI)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &Agent{
		conf: c,
		http: httpcli,
		mqtt: mqtt.NewDispatcher(c.ClientConfig),
		log:  logger.WithFields("cloud", "agent"),
	}, nil
}

// Start starts agent
func (a *Agent) Start(reload func(string) map[string]interface{}) error {
	return errors.Trace(a.mqtt.Start(func(pkt packet.Generic) {
		publish, ok := pkt.(*packet.Publish)
		if !ok {
			logger.Warnf("Packet unexpected")
			return
		}
		e := NewEvent(publish.Message.Payload)
		logger.Debugln("Backward event:", e)
		switch e.Type {
		case SyncConfig:
			var report map[string]interface{}
			err := a.sync(e.Detail.Version, e.Detail.DownloadURL)
			if err != nil {
				logger.WithError(err).Errorf("Failed to download new config package")
				report = map[string]interface{}{"reload_error": err.Error()}
			} else {
				report = reload(e.Detail.Version)
			}
			a.Report(report)
		default:
			logger.Warnf("Event type unexpected")
		}
		if publish.Message.QOS == 1 {
			puback := packet.NewPuback()
			puback.ID = publish.ID
			a.mqtt.Send(puback)
		}
	}))
}

// Report reports info
func (a *Agent) Report(parts map[string]interface{}) {
	r := NewReport(parts)
	p := packet.NewPublish()
	p.Message.Topic = fmt.Sprintf(CloudForward, a.conf.ClientID)
	p.Message.Payload = r.Bytes()
	err := a.mqtt.Send(p)
	if err != nil {
		logger.WithError(err).Warnf("Failed to report by mqtt")
	}
	err = a.report(a.conf.Key, p.Message.Payload)
	if err != nil {
		logger.WithError(err).Warnf("Failed to report by https")
	}
}

// Close closes agent
func (a *Agent) Close() error {
	return errors.Trace(a.mqtt.Close())
}

func (a *Agent) sync(version, url string) error {
	data, err := a.download(a.conf.Key, url)
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Trace(ioutil.WriteFile(version+".zip", data, 0644))
}
