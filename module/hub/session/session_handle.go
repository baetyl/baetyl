package session

import (
	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/module/hub/auth"
	"github.com/baidu/openedge/module/hub/common"
	"github.com/baidu/openedge/module/hub/utils"
	"github.com/baidu/openedge/trans/mqtt"
	"github.com/docker/distribution/uuid"
	"github.com/juju/errors"
)

// Handle handles mqtt connection
func (s *session) Handle() {
	var err error
	var p packet.Generic
	for {
		p, err = s.conn.Receive()
		if err != nil {
			if !s.tomb.Alive() {
				return
			}
			s.log.WithError(err).Warn("Reveive message failed")
			s.close(true)
			return
		}
		if _, ok := p.(*packet.Connect); !ok && s.authorizer == nil {
			s.log.Error("Only connect packet allowed before auth")
			s.close(true)
			return
		}
		s.log.Debugln("Received:", p)
		switch pack := p.(type) {
		case *packet.Connect:
			err = s.onConnect(pack)
		case *packet.Publish:
			err = s.onPublish(pack)
		case *packet.Puback:
			err = s.onPuback(pack)
		case *packet.Subscribe:
			err = s.onSubscribe(pack)
		case *packet.Pingreq:
			err = s.onPingreq(pack)
		case *packet.Pingresp:
			err = nil // just ignore
		case *packet.Disconnect:
			s.close(false)
			return
		case *packet.Unsubscribe:
			err = s.onUnsubscribe(pack)
		default:
			err = errors.NotSupportedf("Packet (%v)", pack)
		}
		if err != nil {
			s.log.Debug(err.Error())
			s.close(true)
			break
		}
	}
}

func (s *session) onConnect(p *packet.Connect) error {
	s.log.Data[common.LogClient] = p.ClientID
	if p.Version != packet.Version31 && p.Version != packet.Version311 {
		s.log.WithField(common.LogMQTTVersion, p.Version).Error("MQTT protocol version invalid")
		return s.sendConnack(packet.InvalidProtocolVersion)
	}
	// TODO: test
	if tlsconn, ok := mqtt.GetTLSConn(s.conn); ok && (p.Username == "" || p.Password == "") {
		sn, err := mqtt.GetClientCertSerialNumber(tlsconn)
		if err != nil {
			s.log.WithError(err).Error("Client certificate invalid")
			return s.sendConnack(packet.NotAuthorized)
		}
		authorizer := s.manager.auth.AuthenticateCert(sn)
		if authorizer == nil {
			s.log.WithField("serialnumber", sn).Error("Client certificate not permitted")
			return s.sendConnack(packet.NotAuthorized)
		}
		s.authorizer = authorizer
	} else {
		authorizer := s.manager.auth.AuthenticateAccount(p.Username, p.Password)
		if authorizer == nil {
			s.log.WithField("username", p.Username).Error("Username or password not permitted")
			return s.sendConnack(packet.BadUsernameOrPassword)
		}
		s.authorizer = authorizer
	}
	if !utils.IsClientID(p.ClientID) {
		s.log.Error("Client ID invalid")
		return s.sendConnack(packet.IdentifierRejected)
	}
	if p.Will != nil {
		// TODO: remove?
		if !common.PubTopicValidate(p.Will.Topic) {
			s.log.WithField(common.LogWillTopic, p.Will.Topic).Error("Will topic invalid")
			return errors.Errorf("Will topic (%s) invalid", p.Will.Topic)
		}
		if !s.authorizer.Authorize(auth.Publish, p.Will.Topic) {
			s.log.WithField("topic", p.Will.Topic).Error("Will topic not permitted")
			return s.sendConnack(packet.NotAuthorized)
		}
		if p.Will.QOS > 1 {
			s.log.WithField(common.LogMessageQOS, p.Will.QOS).Error("QOS not supported")
			return errors.NotSupportedf("QOS (%d)", p.Will.QOS)
		}
	}
	var err error
	s.clientID = p.ClientID
	s.clean = p.CleanSession
	if p.ClientID == "" {
		s.id = common.PrefixTmp + uuid.Generate().String()
		s.clean = true
	} else {
		s.id = common.PrefixSess + p.ClientID
	}
	err = s.manager.register(s)
	if err != nil {
		s.log.WithError(err).Error("Create session rule failed")
		return errors.Annotate(err, "Create session rule failed")
	}
	subs, err := s.manager.recorder.getSubs(s.id)
	if err != nil {
		return errors.Trace(err)
	}
	if s.clean {
		for _, sub := range subs {
			err = s.manager.recorder.removeSub(s.id, sub.Topic)
			if err != nil {
				return errors.Trace(err)
			}
		}
		s.log.Debugf("Clean session state successfully")
	} else {
		// bce-iot-5347
		// Re-check subscriptions, if subscription not permit, log error and skip
		rv := s.genSubAck(subs)
		for i, sub := range subs {
			if rv[i] == packet.QOSFailure {
				s.log.WithField("topic", sub.Topic).Error("Failed to resubscribe topic")
				err = s.manager.recorder.removeSub(s.id, sub.Topic)
				if err != nil {
					return errors.Trace(err)
				}
				continue
			}
			s.subs[sub.Topic] = sub
			err := s.manager.rules.AddSinkSub(s.id, s.id, uint32(sub.QOS), sub.Topic, uint32(sub.QOS), "")
			if err != nil {
				s.log.WithError(err).Error("Resubscribe failed")
				return errors.Annotate(err, "Resubscribe failed")
			}
			s.log.WithField(common.LogSinkTopic, sub.Topic).Info("Topic resubscribed")
		}
		s.log.Debugf("Resume session state successfully")
	}
	err = s.saveWillMessage(p)
	if err != nil {
		return errors.Trace(err)
	}
	err = s.sendConnack(packet.ConnectionAccepted)
	if err != nil {
		s.log.WithError(err).Error("Send connect ack failed")
		return errors.Trace(err)
	}
	err = s.manager.rules.StartRule(s.id)
	if err != nil {
		s.log.WithError(err).Error("Start session rule failed")
		return errors.Trace(err)
	}
	s.log.Info("session connected successfully")
	return nil
}

func (s *session) onPublish(p *packet.Publish) error {
	if _, ok := s.permittedPublishTopics[p.Message.Topic]; !ok {
		// TODO: remove?
		if !common.PubTopicValidate(p.Message.Topic) {
			s.log.WithField("topic", p.Message.Topic).Error("Publish topic invalid")
			return errors.Errorf("Topic (%s) invalid", p.Message.Topic)
		}
		if !s.authorizer.Authorize(auth.Publish, p.Message.Topic) {
			s.log.WithField("Topic", p.Message.Topic).Error("Publish topic not permitted")
			return errors.Errorf("Topic (%s) not permitted", p.Message.Topic)
		}
		s.permittedPublishTopics[p.Message.Topic] = struct{}{}
	}
	if p.Message.QOS > 1 {
		s.log.WithField(common.LogMessageQOS, p.Message.QOS).Error("QOS not supported")
		return errors.NotSupportedf("QOS (%d)", p.Message.QOS)
	}
	err := s.retainMessage(&p.Message)
	if err != nil {
		return errors.Trace(err)
	}
	msg := common.NewMessage(uint32(p.Message.QOS), p.Message.Topic, p.Message.Payload, s.clientID)
	if p.Message.QOS == 1 {
		msg.SetCallbackPID(uint32(p.ID), s.callback)
	}
	s.manager.flow(msg)
	return nil
}

func (s *session) onPuback(p *packet.Puback) error {
	// s.log.Debugf("Receive puback: pid=%d", p.ID)
	if !s.pids.Ack(p.ID) {
		s.log.Warnf("Puback(pid=%d) not found", p.ID)
	}
	return nil
}

func (s *session) onSubscribe(p *packet.Subscribe) error {
	ack := packet.NewSuback()
	rv := s.genSubAck(p.Subscriptions)
	for i, sub := range p.Subscriptions {
		if rv[i] == packet.QOSFailure {
			s.log.WithField("topic", sub.Topic).Error("Failed to subscribe topic")
			continue
		}
		if _, ok := s.subs[sub.Topic]; !ok {
			err := s.manager.rules.AddSinkSub(s.id, s.id, uint32(sub.QOS), sub.Topic, uint32(sub.QOS), "")
			if err != nil {
				return errors.Trace(err)
			}
			s.log.WithField("topic", sub.Topic).Info("Topic subscribed")
			s.subs[sub.Topic] = sub
			if !s.clean {
				err := s.manager.recorder.addSub(s.id, sub)
				if err != nil {
					return errors.Trace(err)
				}
			}
		} else {
			if s.subs[sub.Topic].QOS != sub.QOS {
				// s.manager.rules.RemoveSinkSub(s.id, sub.Topic)
				err := s.manager.rules.AddSinkSub(s.id, s.id, uint32(sub.QOS), sub.Topic, uint32(sub.QOS), "")
				if err != nil {
					return errors.Trace(err)
				}
				s.log.WithField("topic", sub.Topic).Info("Topic subscribed")
				s.subs[sub.Topic] = sub
				if !s.clean {
					err := s.manager.recorder.removeSub(s.id, sub.Topic)
					if err != nil {
						return errors.Trace(err)
					}
					err = s.manager.recorder.addSub(s.id, sub)
					if err != nil {
						return errors.Trace(err)
					}
				}
			}
		}
	}
	ack.ID = p.ID
	ack.ReturnCodes = rv
	err := s.send(ack, true)
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Trace(s.sendRetainMessage(p))
}

func (s *session) onUnsubscribe(p *packet.Unsubscribe) error {
	ack := packet.NewUnsuback()
	for _, topic := range p.Topics {
		if _, ok := s.subs[topic]; ok {
			err := s.manager.rules.RemoveSinkSub(s.id, topic)
			if err != nil {
				s.log.WithError(err).Warn("Remove subscription from session rule failed")
			}
			delete(s.subs, topic)
			if !s.clean {
				s.manager.recorder.removeSub(s.id, topic)
			}
			s.log.WithField(common.LogUnSubTopic, topic).Info("Unsubscribe topic successfully")
		} else {
			s.log.WithField(common.LogUnSubTopic, topic).Warn("Topic not subscribed yet")
		}
	}
	ack.ID = p.ID
	return errors.Trace(s.send(ack, true))
}

func (s *session) onPingreq(p *packet.Pingreq) error {
	return errors.Trace(s.send(packet.NewPingresp(), true))
}
