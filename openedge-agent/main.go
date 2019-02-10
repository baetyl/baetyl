package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/protocol/http"
	"github.com/baidu/openedge/protocol/mqtt"
	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/baidu/openedge/utils"
)

// mo agent module
type mo struct {
	cfg  Config
	key  []byte
	path string
	ctx  openedge.Context
	mqtt *mqtt.Dispatcher
	http *http.Client
	tomb utils.Tomb
}

const defaultConfigPath = "etc/openedge/service.yml"

func main() {
	openedge.Run(func(ctx openedge.Context) error {
		m, err := new(ctx)
		if err != nil {
			return err
		}
		defer m.close()
		err = m.start(ctx)
		if err != nil {
			return err
		}
		ctx.Wait()
		return nil
	})
}

func new(ctx openedge.Context) (*mo, error) {
	var cfg Config
	err := utils.LoadYAML(defaultConfigPath, &cfg)
	if err != nil {
		return nil, err
	}
	err = defaults(&cfg)
	if err != nil {
		return nil, err
	}
	key, err := ioutil.ReadFile(cfg.Remote.MQTT.Key)
	if err != nil {
		return nil, err
	}
	cli, err := http.NewClient(cfg.Remote.HTTP)
	if err != nil {
		return nil, err
	}
	return &mo{
		cfg:  cfg,
		key:  key,
		ctx:  ctx,
		http: cli,
		mqtt: mqtt.NewDispatcher(cfg.Remote.MQTT),
	}, nil
}

func (m *mo) start(ctx openedge.Context) error {
	err := m.mqtt.Start(m)
	if err != nil {
		return err
	}
	return m.tomb.Go(m.reporting)
}

func (m *mo) close() {
	m.tomb.Kill(nil)
	m.tomb.Wait()
	m.mqtt.Close()
}

func (m *mo) ProcessPublish(p *packet.Publish) error {
	e := NewEvent(p.Message.Payload)
	m.ctx.Log().Debugln("backward event:", e)
	switch e.Type {
	case Update:
		dataset, err := openedge.NewDatasetInfoFromBytes(e.Content)
		if err != nil {
			err := fmt.Errorf("update event invalid: %s", err.Error())
			m.ctx.Log().Errorf(err.Error())
			m.report(err.Error())
			break
		}
		if !isVersion(dataset.Version) {
			err := fmt.Errorf("update event invalid: dataset version invalid")
			m.ctx.Log().Errorf(err.Error())
			m.report(err.Error())
			break
		}
		err = m.ctx.UpdateSystem(dataset)
		if err != nil {
			err := fmt.Errorf("failed to update system: %s", err.Error())
			m.ctx.Log().Errorf(err.Error())
			m.report(err.Error())
		} else {
			m.report()
		}
	default:
		m.ctx.Log().Warnf("event type unexpected")
	}
	if p.Message.QOS == 1 {
		puback := packet.NewPuback()
		puback.ID = p.ID
		m.mqtt.Send(puback)
	}
	return nil
}

func (m *mo) ProcessPuback(p *packet.Puback) error {
	return nil
}

func (m *mo) ProcessError(err error) {
	m.ctx.Log().Errorf(err.Error())
}

func (m *mo) reporting() error {
	t := time.NewTicker(m.cfg.Remote.Report.Interval)
	m.report()
	defer m.report()
	for {
		select {
		case <-t.C:
			m.ctx.Log().Debugln("to report stats")
			m.report()
		case <-m.tomb.Dying():
			return nil
		}
	}
}

// Report reports info
func (m *mo) report(errors ...string) {
	defer trace("report")()

	i, err := m.ctx.InspectSystem()
	if err != nil {
		m.ctx.Log().WithError(err).Warnf("failed to inspect stats")
		i = openedge.NewInspect()
		errors = append(errors, err.Error())
	}
	i.Error = strings.Join(errors, ";")
	payload, err := json.Marshal(i)
	if err != nil {
		m.ctx.Log().WithError(err).Warnf("failed to marshal stats")
		return
	}
	m.ctx.Log().Debugln("stats", string(payload))
	p := packet.NewPublish()
	p.Message.Topic = m.cfg.Remote.Report.Topic
	p.Message.Payload = payload
	err = m.mqtt.Send(p)
	if err != nil {
		m.ctx.Log().WithError(err).Warnf("failed to report stats by mqtt")
	}
	err = m.send(p.Message.Payload)
	if err != nil {
		m.ctx.Log().WithError(err).Warnf("failed to report stats by https")
	}
}

func (m *mo) send(data []byte) error {
	body, key, err := m.encryptData(data)
	if err != nil {
		return err
	}
	header := map[string]string{
		"x-iot-edge-clientid": m.cfg.Remote.MQTT.ClientID,
		"x-iot-edge-key":      key,
		"Content-Type":        "application/x-www-form-urlencoded",
	}
	_, err = m.http.Send("POST", m.cfg.Remote.Report.URL, body, header)
	return err
}

func (m *mo) encryptData(data []byte) ([]byte, string, error) {
	aesKey := utils.NewAesKey()
	// encrypt data using AES
	body, err := utils.AesEncrypt(data, aesKey)
	if err != nil {
		return nil, "", err
	}
	// encrypt AES key using RSA
	k, err := utils.RsaPrivateEncrypt(aesKey, m.key)
	if err != nil {
		return nil, "", err
	}
	// encode key using BASE64
	key := base64.StdEncoding.EncodeToString(k)
	// encode body using BASE64
	body = []byte(base64.StdEncoding.EncodeToString(body))
	return body, key, nil
}

func defaults(c *Config) error {
	if c.Remote.MQTT.Address == "" {
		return fmt.Errorf("remote mqtt address missing")
	}
	if c.Remote.HTTP.CA == "" {
		return fmt.Errorf("remote http ca missing, must enable ssl")
	}
	if c.Remote.HTTP.Address == "" {
		if strings.Contains(c.Remote.MQTT.Address, "bj.baidubce.com") {
			c.Remote.HTTP.Address = "https://iotedge.bj.baidubce.com"
		} else if strings.Contains(c.Remote.MQTT.Address, "gz.baidubce.com") {
			c.Remote.HTTP.Address = "https://iotedge.gz.baidubce.com"
		} else {
			return fmt.Errorf("remote http address missing")
		}
	}
	c.Remote.Desire.Topic = fmt.Sprintf(c.Remote.Desire.Topic, c.Remote.MQTT.ClientID)
	c.Remote.Report.Topic = fmt.Sprintf(c.Remote.Report.Topic, c.Remote.MQTT.ClientID)
	c.Remote.MQTT.Subscriptions = append(c.Remote.MQTT.Subscriptions, mqtt.TopicInfo{QoS: 1, Topic: c.Remote.Desire.Topic})
	return nil
}

// IsVersion checks version
func isVersion(v string) bool {
	r := regexp.MustCompile("^[\\w\\.]+$")
	return r.MatchString(v)
}

func trace(name string) func() {
	start := time.Now()
	return func() {
		logger.Debugf("%s elapsed time: %v", name, time.Since(start))
	}
}
