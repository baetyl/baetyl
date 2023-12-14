// Package mqttlink 端云链接 mqtt 实现
package mqttlink

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	v2plugin "github.com/baetyl/baetyl-go/v2/plugin"
	specV1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"

	"github.com/baetyl/baetyl/v2/common"
	"github.com/baetyl/baetyl/v2/plugin"
)

var (
	ErrLinkTLSConfigMissing = errors.New("certificate bidirectional authentication is required for connection with cloud")
	ErrCertificateFormat    = errors.New("certificate format error")
	ErrInvalidCertificate   = errors.New("failed to parse node name from cert")
	ErrUnavailableAddress   = errors.New("there is no available address")
)

type mqttLink struct {
	msgCh      chan *specV1.Message
	observer   mqtt.Observer
	obsCh      chan *specV1.Message
	urls       []string
	keeper     common.SendKeeper
	cfg        Config
	cli        *mqtt.Client
	errCh      chan error
	ctx        context.Context
	state      *specV1.Message
	stateMutex sync.RWMutex
	cancel     context.CancelFunc
	log        *log.Logger
	ns         string
	name       string
}

func (mt *mqttLink) Close() error {
	close(mt.msgCh)
	mt.cancel()
	return mt.cli.Close()
}

func init() {
	v2plugin.RegisterFactory("mqttlink", New)
}

func New() (v2plugin.Plugin, error) {
	var cfg Config
	if err := utils.LoadYAML(plugin.ConfFile, &cfg); err != nil {
		return nil, errors.Trace(err)
	}
	log.L().Debug("config", log.Any("cfg", cfg))

	ctx, cancel := context.WithCancel(context.Background())

	link := &mqttLink{
		keeper: common.SendKeeper{},
		msgCh:  make(chan *specV1.Message, 1),
		errCh:  make(chan error, 1),
		cfg:    cfg,
		ctx:    ctx,
		state:  &specV1.Message{Kind: plugin.LinkStateUnknown, Content: specV1.LazyValue{Value: ""}},
		obsCh:  make(chan *specV1.Message, 1024),
		cancel: cancel,
		log:    log.With(log.Any("plugin", "mqttlink")),
	}
	tlsConfig, err := utils.NewTLSConfigClient(cfg.Node)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if len(tlsConfig.Certificates) == 1 && len(tlsConfig.Certificates[0].Certificate) == 1 {
		cert, err := x509.ParseCertificate(tlsConfig.Certificates[0].Certificate[0])
		if err != nil {
			return nil, errors.Trace(ErrCertificateFormat)
		}
		res := strings.SplitN(cert.Subject.CommonName, ".", 2)
		if len(res) != 2 || res[0] == "" || res[1] == "" {
			return nil, errors.Trace(ErrInvalidCertificate)
		}
		link.ns = res[0]
		link.name = res[1]
	}
	link.observer = newObserver(link.obsCh, link.log)
	if link.ns != "" && link.name != "" {
		link.cfg.MqttLink.Report.Topic = replaceTopic(link.cfg.MqttLink.Report.Topic, link.ns, link.name)
		link.cfg.MqttLink.Desire.Topic = replaceTopic(link.cfg.MqttLink.Desire.Topic, link.ns, link.name)
		link.cfg.MqttLink.Delta.Topic = replaceTopic(link.cfg.MqttLink.Delta.Topic, link.ns, link.name)
		link.cfg.MqttLink.DesireResponse.Topic = replaceTopic(link.cfg.MqttLink.DesireResponse.Topic, link.ns, link.name)
	}
	err = link.dial()
	if err != nil {
		return nil, errors.Trace(err)
	}
	go link.receiving()
	return link, nil
}

func replaceTopic(str, ns, n string) string {
	topic := strings.ReplaceAll(str, mqtt.TopicNamespace, ns)
	return strings.ReplaceAll(topic, mqtt.TopicNodeName, n)
}

func (mt *mqttLink) dial() error {
	mt.cfg.MqttLink.Subscriptions = append(mt.cfg.MqttLink.Subscriptions, mt.cfg.MqttLink.Delta, mt.cfg.MqttLink.DesireResponse)
	errs := []string{}
	for _, url := range strings.Split(mt.cfg.MqttLink.Address, ",") {
		if strings.TrimSpace(url) == "" {
			continue
		}
		mt.cfg.MqttLink.Address = url
		ops, err := mt.cfg.MqttLink.ToClientOptions()
		if err != nil {
			return errors.Trace(err)
		}
		if ops.TLSConfig == nil {
			return errors.Trace(ErrLinkTLSConfigMissing)
		}
		cli := mqtt.NewClient(ops)
		err = cli.Start(mt.observer)
		if err != nil {
			mt.log.Warn("failed to connect cloud", log.Any("url", url), log.Error(err))
			errs = append(errs, err.Error())
		} else {
			mt.cli = cli
			return nil
		}
	}
	mt.stateNotify(plugin.LinkStateNetworkError, strings.Join(errs, ";"))
	return errors.Trace(ErrUnavailableAddress)
}

func (mt *mqttLink) Send(msg *specV1.Message) error {
	if msg.Metadata == nil {
		msg.Metadata = make(map[string]string)
	}
	msg.Metadata["namespace"] = mt.ns
	msg.Metadata["name"] = mt.name
	pld, err := json.Marshal(msg)
	if err != nil {
		return errors.Trace(err)
	}
	var topic string
	switch msg.Kind {
	case specV1.MessageReport, specV1.MessageCMD, specV1.MessageData, specV1.MessageDeviceReport, specV1.MessageDeviceLifecycleReport:
		topic = mt.cfg.MqttLink.Report.Topic
	case specV1.MessageDesire, specV1.MessageDeviceDesire, specV1.MessageMultipleDeviceDesire:
		topic = mt.cfg.MqttLink.Desire.Topic
	default:
		return errors.New("invalid msg kind")
	}
	return mt.cli.Publish(mqtt.QOS(mt.cfg.MqttLink.Report.QOS), topic, pld, 0, false, false)
}

func (mt *mqttLink) receiving() {
	for {
		select {
		case msg := <-mt.obsCh:
			data, err := utils.ParseEnv(msg.Content.GetJSON())
			if err != nil {
				mt.log.Error("failed to parse env", log.Error(err))
				continue
			}
			msg.Content.SetJSON(data)
			if common.IsSyncMessage(msg) {
				err = mt.keeper.ReceiveResp(msg)
				if err != nil {
					mt.log.Error("failed to receive response", log.Error(err))
					continue
				}
			} else {
				if msg.Kind == specV1.MessageError {
					var errMsg string
					err = msg.Content.Unmarshal(&errMsg)
					if err != nil {
						mt.log.Error("failed to unmarshal error message", log.Error(err))
						continue
					}
					mt.log.Debug("get cloud error message", log.Any("errMsg", errMsg))
					select {
					case mt.errCh <- errors.New(errMsg):
					case <-mt.ctx.Done():
						return
					}
					if strings.HasSuffix(errMsg, fmt.Sprintf("The (node) resource (%s) is not found.", msg.Metadata["name"])) {
						mt.stateNotify(plugin.LinkStateNodeNotFound, errMsg)
					} else {
						mt.stateNotify(plugin.LinkStateSucceeded, errMsg)
					}
					continue
				}
				mt.stateNotify(plugin.LinkStateSucceeded, plugin.LinkStateSucceeded)
				select {
				case mt.msgCh <- msg:
				}
			}
		case <-mt.ctx.Done():
			return
		}
	}
}

func (mt *mqttLink) Receive() (<-chan *specV1.Message, <-chan error) {
	return mt.msgCh, mt.errCh
}

func (mt *mqttLink) IsAsyncSupported() bool {
	return true
}

func (mt *mqttLink) Request(msg *specV1.Message) (*specV1.Message, error) {
	res, err := mt.keeper.SendSync(msg, mt.cfg.MqttLink.Timeout, mt.Send)
	if err != nil {
		return nil, errors.Trace(err)
	}
	// encapsulation error message
	if res.Kind == specV1.MessageError {
		var errMsg string
		err = res.Content.Unmarshal(&errMsg)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return nil, errors.New(errMsg)
	}
	return res, nil
}

func (mt *mqttLink) State() *specV1.Message {
	mt.stateMutex.RLock()
	copyState := mt.state
	mt.stateMutex.RUnlock()
	return copyState
}

// stateNotify Lock and update the pointer of state
func (mt *mqttLink) stateNotify(kind, msg string) {
	mt.stateMutex.Lock()
	mt.state = &specV1.Message{
		Kind:     specV1.MessageKind(kind),
		Metadata: map[string]string{},
		Content:  specV1.LazyValue{Value: msg},
	}
	mt.stateMutex.Unlock()
}
