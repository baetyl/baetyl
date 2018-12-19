package agent

import (
	"fmt"
	"io/ioutil"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/module/http"
	"github.com/baidu/openedge/module/logger"
	"github.com/baidu/openedge/module/mqtt"
)

// topics from cloud
const (
	CloudForward  = "$baidu/iot/edge/%s/core/forward"
	CloudBackward = "$baidu/iot/edge/%s/core/backward"
)

// Agent agent of edge cloud
type Agent struct {
	conf Config
	http *http.Client
	mqtt *mqtt.Dispatcher
	log  *logger.Entry
}

// NewAgent creates a new agent
func NewAgent(c Config) (*Agent, error) {
	httpcli, err := http.NewClient(c.OpenAPI)
	if err != nil {
		return nil, err
	}
	return &Agent{
		conf: c,
		http: httpcli,
		mqtt: mqtt.NewDispatcher(c.MQTTClient),
		log:  logger.WithFields("cloud", "agent"),
	}, nil
}

// Start starts agent
func (a *Agent) Start(reload func(string) map[string]interface{}) error {
	h := mqtt.Handler{}
	h.ProcessPublish = func(p *packet.Publish) error {
		e := NewEvent(p.Message.Payload)
		logger.Debugln("backward event:", e)
		switch e.Type {
		case SyncConfig:
			var report map[string]interface{}
			err := a.sync(e.Detail.Version, e.Detail.DownloadURL)
			if err != nil {
				logger.WithError(err).Errorf("failed to download new config package")
				report = map[string]interface{}{"reload_error": err.Error()}
			} else {
				report = reload(e.Detail.Version)
			}
			a.Report(report)
		default:
			logger.Warnf("event type unexpected")
		}
		if p.Message.QOS == 1 {
			puback := packet.NewPuback()
			puback.ID = p.ID
			a.mqtt.Send(puback)
		}
		return nil
	}
	return a.mqtt.Start(h)
}

// Report reports info
func (a *Agent) Report(parts map[string]interface{}) {
	r := NewReport(parts)
	p := packet.NewPublish()
	p.Message.Topic = fmt.Sprintf(CloudForward, a.conf.ClientID)
	p.Message.Payload = r.Bytes()
	err := a.mqtt.Send(p)
	if err != nil {
		logger.WithError(err).Warnf("failed to report by mqtt")
	}
	err = a.report(a.conf.Key, p.Message.Payload)
	if err != nil {
		logger.WithError(err).Warnf("failed to report by https")
	}
}

// Close closes agent
func (a *Agent) Close() error {
	return a.mqtt.Close()
}

func (a *Agent) sync(version, url string) error {
	data, err := a.download(a.conf.Key, url)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(version+".zip", data, 0644)
}
