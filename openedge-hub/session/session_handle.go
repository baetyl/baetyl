package session

import (
	"fmt"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/module/mqtt"
	"github.com/baidu/openedge/openedge-hub/auth"
	"github.com/baidu/openedge/openedge-hub/common"
	"github.com/baidu/openedge/openedge-hub/utils"
	"github.com/docker/distribution/uuid"
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
			s.log.WithError(err).Warnf("failed to reveive message")
			s.close(true)
			return
		}
		if _, ok := p.(*packet.Connect); !ok && s.authorizer == nil {
			s.log.Errorf("only connect packet allowed before auth")
			s.close(true)
			return
		}
		s.log.Debugln("received:", p)
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
			err = fmt.Errorf("packet (%v) not supported", pack)
		}
		if err != nil {
			s.log.Errorf(err.Error())
			s.close(true)
			break
		}
	}
}

func (s *session) onConnect(p *packet.Connect) error {
	s.log = s.log.WithFields(common.LogClient, p.ClientID)
	if p.Version != packet.Version31 && p.Version != packet.Version311 {
		s.sendConnack(packet.InvalidProtocolVersion)
		return fmt.Errorf("MQTT protocol version (%d) invalid", p.Version)
	}
	// TODO: test
	if tlsconn, ok := mqtt.GetTLSConn(s.conn); ok && (p.Username == "" || p.Password == "") {
		sn, err := mqtt.GetClientCertSerialNumber(tlsconn)
		if err != nil {
			s.sendConnack(packet.NotAuthorized)
			return fmt.Errorf("client certificate invalid: %s", err.Error())
		}
		authorizer := s.manager.auth.AuthenticateCert(sn)
		if authorizer == nil {
			s.sendConnack(packet.NotAuthorized)
			return fmt.Errorf("client certificate (sn:%s) not permitted", sn)
		}
		s.authorizer = authorizer
	} else {
		authorizer := s.manager.auth.AuthenticateAccount(p.Username, p.Password)
		if authorizer == nil {
			s.sendConnack(packet.BadUsernameOrPassword)
			return fmt.Errorf("username (%s) or password not permitted", p.Username)
		}
		s.authorizer = authorizer
	}
	if !utils.IsClientID(p.ClientID) {
		s.sendConnack(packet.IdentifierRejected)
		return fmt.Errorf("client ID (%s) invalid", p.ClientID)
	}
	if p.Will != nil {
		// TODO: remove?
		if !common.PubTopicValidate(p.Will.Topic) {
			return fmt.Errorf("will topic (%s) invalid", p.Will.Topic)
		}
		if !s.authorizer.Authorize(auth.Publish, p.Will.Topic) {
			s.sendConnack(packet.NotAuthorized)
			return fmt.Errorf("will topic (%s) not permitted", p.Will.Topic)
		}
		if p.Will.QOS > 1 {
			return fmt.Errorf("will QOS (%d) not supported", p.Will.QOS)
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
		return fmt.Errorf("failed to create session rule: %s", err.Error())
	}
	subs, err := s.manager.recorder.getSubs(s.id)
	if err != nil {
		return err
	}
	if s.clean {
		for _, sub := range subs {
			err = s.manager.recorder.removeSub(s.id, sub.Topic)
			if err != nil {
				return err
			}
		}
		s.log.Debugf("session state cleaned")
	} else {
		// bce-iot-5347
		// Re-check subscriptions, if subscription not permit, log error and skip
		rv := s.genSubAck(subs)
		for i, sub := range subs {
			if rv[i] == packet.QOSFailure {
				s.log.Errorf("failed to resubscribe topic (%s)", sub.Topic)
				err = s.manager.recorder.removeSub(s.id, sub.Topic)
				if err != nil {
					return err
				}
				continue
			}
			s.subs[sub.Topic] = sub
			err := s.manager.rules.AddSinkSub(s.id, s.id, uint32(sub.QOS), sub.Topic, uint32(sub.QOS), "")
			if err != nil {
				return fmt.Errorf("failed to resubscribe: %s", err.Error())
			}
			s.log.Infof("topic (%s) resubscribed", sub.Topic)
		}
		s.log.Debugf("session state resumed")
	}
	err = s.saveWillMessage(p)
	if err != nil {
		return err
	}
	err = s.sendConnack(packet.ConnectionAccepted)
	if err != nil {
		return err
	}
	err = s.manager.rules.StartRule(s.id)
	if err != nil {
		return err
	}
	s.log.Infof("session connected")
	return nil
}

func (s *session) onPublish(p *packet.Publish) error {
	if _, ok := s.permittedPublishTopics[p.Message.Topic]; !ok {
		// TODO: remove?
		if !common.PubTopicValidate(p.Message.Topic) {
			return fmt.Errorf("publish topic (%s) invalid", p.Message.Topic)
		}
		if !s.authorizer.Authorize(auth.Publish, p.Message.Topic) {
			return fmt.Errorf("publish topic (%s) not permitted", p.Message.Topic)
		}
		s.permittedPublishTopics[p.Message.Topic] = struct{}{}
	}
	if p.Message.QOS > 1 {
		return fmt.Errorf("publish QOS (%d) not supported", p.Message.QOS)
	}
	err := s.retainMessage(&p.Message)
	if err != nil {
		return err
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
		s.log.Warnf("puback(pid=%d) not found", p.ID)
	}
	return nil
}

func (s *session) onSubscribe(p *packet.Subscribe) error {
	ack := packet.NewSuback()
	rv := s.genSubAck(p.Subscriptions)
	for i, sub := range p.Subscriptions {
		if rv[i] == packet.QOSFailure {
			s.log.Errorf("failed to subscribe topic (%s)", sub.Topic)
			continue
		}
		if _, ok := s.subs[sub.Topic]; !ok {
			err := s.manager.rules.AddSinkSub(s.id, s.id, uint32(sub.QOS), sub.Topic, uint32(sub.QOS), "")
			if err != nil {
				return err
			}
			s.log.Infof("topic (%s) subscribed", sub.Topic)
			s.subs[sub.Topic] = sub
			if !s.clean {
				err := s.manager.recorder.addSub(s.id, sub)
				if err != nil {
					return err
				}
			}
		} else {
			if s.subs[sub.Topic].QOS != sub.QOS {
				// s.manager.rules.RemoveSinkSub(s.id, sub.Topic)
				err := s.manager.rules.AddSinkSub(s.id, s.id, uint32(sub.QOS), sub.Topic, uint32(sub.QOS), "")
				if err != nil {
					return err
				}
				s.log.Infof("topic (%s) subscribed", sub.Topic)
				s.subs[sub.Topic] = sub
				if !s.clean {
					err := s.manager.recorder.removeSub(s.id, sub.Topic)
					if err != nil {
						return err
					}
					err = s.manager.recorder.addSub(s.id, sub)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	ack.ID = p.ID
	ack.ReturnCodes = rv
	err := s.send(ack, true)
	if err != nil {
		return err
	}
	return s.sendRetainMessage(p)
}

func (s *session) onUnsubscribe(p *packet.Unsubscribe) error {
	ack := packet.NewUnsuback()
	for _, topic := range p.Topics {
		if _, ok := s.subs[topic]; ok {
			err := s.manager.rules.RemoveSinkSub(s.id, topic)
			if err != nil {
				s.log.Errorf(err.Error())
			}
			delete(s.subs, topic)
			if !s.clean {
				s.manager.recorder.removeSub(s.id, topic)
			}
			s.log.Infof("topic (%s) unsubscribed", topic)
		} else {
			s.log.Warnf("topic (%s) not subscribed yet", topic)
		}
	}
	ack.ID = p.ID
	return s.send(ack, true)
}

func (s *session) onPingreq(p *packet.Pingreq) error {
	return s.send(packet.NewPingresp(), true)
}
