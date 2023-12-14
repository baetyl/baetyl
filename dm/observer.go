// Package dm 设备管理实现
package dm

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/256dpi/gomqtt/packet"
	dm "github.com/baetyl/baetyl-go/v2/dmcontext"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

const (
	KindGet             = "get"
	kindReport          = "report"
	kindPropertyReport  = "thing.property.post"
	kindLifecycleReport = "thing.lifecycle.post"

	DeviceTopicRe      = "\\$baetyl/device/(.+)/(.+)"
	BlinkDeviceTopicRe = "thing/(.+)/(.+)/(.+)/(.+)"
	BaetylTopicPrefix  = "$"
)

var (
	ErrIllegalTopic = errors.New("failed to parse topic")
)

type observer struct {
	msgCh chan *v1.Message
	log   *log.Logger
}

func newObserver(msgCh chan *v1.Message, log *log.Logger) mqtt.Observer {
	return &observer{
		msgCh: msgCh,
		log:   log,
	}
}

func parseDeviceTopic(topic string) (string, string, error) {
	if strings.HasPrefix(topic, BaetylTopicPrefix) {
		return parseTopic(topic)
	}
	return parseBlinkTopic(topic)
}

func parseTopic(topic string) (string, string, error) {
	r, err := regexp.Compile(DeviceTopicRe)
	if err != nil {
		return "", "", err
	}
	res := r.FindStringSubmatch(topic)
	if len(res) != 3 {
		return "", "", errors.Trace(ErrIllegalTopic)
	}
	return res[1], res[2], nil
}

func parseBlinkTopic(topic string) (string, string, error) {
	r, err := regexp.Compile(BlinkDeviceTopicRe)
	if err != nil {
		return "", "", err
	}
	res := r.FindStringSubmatch(topic)
	if len(res) != 5 {
		return "", "", errors.Trace(ErrIllegalTopic)
	}
	return res[2], fmt.Sprintf("thing.%s.%s", res[3], res[4]), nil
}

func (o *observer) OnPublish(pkt *packet.Publish) error {
	deviceName, kind, err := parseDeviceTopic(pkt.Message.Topic)
	if err != nil {
		o.log.Error("parse topic failed", log.Any("topic", pkt.Message.Topic))
		return nil
	}
	msg := v1.Message{Metadata: make(map[string]string)}
	switch kind {
	//case KindGet:
	//	msg.Kind = v1.MessageDeviceDesire
	//	msg.Metadata[KeyDevice] = deviceName
	case KindGet, kindReport, kindPropertyReport, kindLifecycleReport:
		if err = json.Unmarshal(pkt.Message.Payload, &msg); err != nil {
			o.log.Error("failed to get message", log.Any("device", deviceName), log.Error(err))
			return nil
		}
		msg.Metadata[dm.KeyDevice] = deviceName
	default:
		o.log.Error("get message from unexpected topic")
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		o.log.Error("failed to marshal msg")
	}
	msgStr := string(bytes)
	select {
	case o.msgCh <- &msg:
		o.log.Debug("observer receive downside message", log.Any("msg", msgStr))
	default:
		o.log.Error("failed to write downside message to channel", log.Any("msg", msgStr))
	}
	return nil
}

func (o *observer) OnPuback(_ *packet.Puback) error {
	return nil
}

func (o *observer) OnError(err error) {
	o.log.Error("receive mqtt message error", log.Error(err))
}
