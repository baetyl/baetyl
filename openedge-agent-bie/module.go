package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/http"
	"github.com/baidu/openedge/module/logger"
	"github.com/baidu/openedge/module/master"
	"github.com/baidu/openedge/module/mqtt"
	"github.com/baidu/openedge/module/utils"
)

// mo agent module
type mo struct {
	cfg  Config
	key  []byte
	path string
	http *http.Client
	mqtt *mqtt.Dispatcher
	cli  *master.Client
	tomb utils.Tomb
}

// New create a new module
func New(confFile string) (module.Module, error) {
	var cfg Config
	err := module.Load(&cfg, confFile)
	if err != nil {
		return nil, err
	}
	err = defaults(&cfg)
	if err != nil {
		return nil, err
	}
	err = logger.Init(cfg.Logger, "module", cfg.UniqueName())
	if err != nil {
		return nil, err
	}
	key, err := ioutil.ReadFile(cfg.Remote.MQTT.Key)
	if err != nil {
		return nil, err
	}
	httpcli, err := http.NewClient(cfg.Remote.HTTP)
	if err != nil {
		return nil, err
	}
	mcli, err := master.NewClient(cfg.API)
	if err != nil {
		return nil, err
	}
	p := path.Join("var", "db", "openedge", "volume", cfg.Name)
	err = os.MkdirAll(p, os.ModePerm)
	if err != nil {
		logger.Log.WithError(err).Errorf("failed to make dir: %s", p)
	}
	return &mo{
		cfg:  cfg,
		key:  key,
		cli:  mcli,
		http: httpcli,
		mqtt: mqtt.NewDispatcher(cfg.Remote.MQTT),
		path: p,
	}, nil
}

// Start starts module
func (m *mo) Start() error {
	h := mqtt.Handler{}
	h.ProcessError = func(err error) {
		logger.Log.Errorf(err.Error())
	}
	h.ProcessPublish = func(p *packet.Publish) error {
		e := NewEvent(p.Message.Payload)
		logger.Log.Debugln("backward event:", e)
		switch e.Type {
		case SyncConfig:
			if !isVersion(e.Detail.Version) {
				return fmt.Errorf("new config version invalid")
			}
			confFile, err := m.download(e.Detail.Version, e.Detail.DownloadURL)
			if err != nil {
				logger.Log.WithError(err).Errorf("failed to download new config package")
				m.report("error", err.Error())
				break
			}
			err = m.cli.Reload(confFile)
			if err != nil {
				logger.Log.WithError(err).Errorf("failed to download new config package")
				m.report("error", err.Error())
			} else {
				m.report()
			}
		default:
			logger.Log.Warnf("event type unexpected")
		}
		if p.Message.QOS == 1 {
			puback := packet.NewPuback()
			puback.ID = p.ID
			m.mqtt.Send(puback)
		}
		return nil
	}
	err := m.mqtt.Start(h)
	if err != nil {
		return err
	}
	return m.tomb.Go(m.reporting)
}

// Close closes module
func (m *mo) Close() {
	m.tomb.Kill(nil)
	m.tomb.Wait()
	m.mqtt.Close()
}

func (m *mo) reporting() error {
	t := time.NewTicker(m.cfg.Remote.Report.Interval)
	m.report()
	defer m.report()
	for {
		select {
		case <-t.C:
			logger.Log.Debugln("to report stats")
			m.report()
		case <-m.tomb.Dying():
			return nil
		}
	}
}

// Report reports info
func (m *mo) report(args ...string) {
	defer trace("report")()

	s, err := m.cli.Stats()
	if err != nil {
		logger.Log.WithError(err).Errorf("failed to get master stats")
		s = master.NewStats()
		s.Info["error"] = err.Error()
	}
	for index := 0; index < len(args)-1; index = index + 2 {
		s.Info[args[index]] = args[index+1]
	}
	payload, err := json.Marshal(s)
	if err != nil {
		logger.Log.Debugln("stats", string(payload))
		return
	}
	p := packet.NewPublish()
	p.Message.Topic = m.cfg.Remote.Report.Topic
	p.Message.Payload = payload
	err = m.mqtt.Send(p)
	if err != nil {
		logger.Log.WithError(err).Warnf("failed to report stats by mqtt")
	}
	err = m.send(p.Message.Payload)
	if err != nil {
		logger.Log.WithError(err).Warnf("failed to report stats by https")
	}
}

func (m *mo) send(data []byte) error {
	body, key, err := m.encryptData(data)
	if err != nil {
		return err
	}
	headers := http.Headers{}
	headers.Set("x-iot-edge-clientid", m.cfg.Remote.MQTT.ClientID)
	headers.Set("x-iot-edge-key", key)
	headers.Set("Content-Type", "application/x-www-form-urlencoded")
	url := fmt.Sprintf("%s://%s/%s", m.http.Addr.Scheme, m.http.Addr.Host, m.cfg.Remote.Report.URL)
	_, _, err = m.http.Send("POST", url, headers, body)
	return err
}

func (m *mo) download(version, downloadURL string) (string, error) {
	reqHeaders := http.Headers{}
	// reqHeaders.Set("x-iot-edge-clientid", m.cfg.Remote.MQTT.ClientID)
	// reqHeaders.Set("Content-Type", "application/octet-stream")
	_, resBody, err := m.http.Send("GET", downloadURL, reqHeaders, nil)
	if err != nil {
		return "", err
	}
	// data, err := m.decryptData(resHeaders.Get("x-iot-edge-key"), resBody)
	// if err != nil {
	// 	return "", err
	// }
	file := path.Join(m.path, version+".zip")
	return file, ioutil.WriteFile(file, resBody, 0644)
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

// func (m *mo) decryptData(key string, data []byte) ([]byte, error) {
// 	// decode key using BASE64
// 	k, err := base64.StdEncoding.DecodeString(key)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// decrypt AES key using RSA
// 	aesKey, err := utils.RsaPrivateDecrypt(k, m.key)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// decrypt data using AES
// 	decData, err := utils.AesDecrypt(data, aesKey)
// 	return decData, err
// }

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
	if c.API.Address == "" {
		c.API.Address = utils.GetEnv(module.EnvOpenEdgeMasterAPI)
	}
	c.API.Username = c.UniqueName()
	c.API.Password = utils.GetEnv(module.EnvOpenEdgeModuleToken)
	c.Remote.Desire.Topic = fmt.Sprintf(c.Remote.Desire.Topic, c.Remote.MQTT.ClientID)
	c.Remote.Report.Topic = fmt.Sprintf(c.Remote.Report.Topic, c.Remote.MQTT.ClientID)
	c.Remote.MQTT.Subscriptions = append(c.Remote.MQTT.Subscriptions, config.Subscription{QOS: 1, Topic: c.Remote.Desire.Topic})
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
		logger.Log.Debugf("%s elapsed time: %v", name, time.Since(start))
	}
}
