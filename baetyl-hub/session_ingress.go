package hub

import (
	"fmt"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baetyl/baetyl/baetyl-hub/common"
	"github.com/baetyl/baetyl/baetyl-hub/utils"
	"github.com/baetyl/baetyl/protocol/mqtt"
	"github.com/docker/distribution/uuid"
)

func (s *session) handle() {
	var err error
	var pkt packet.Generic
	for {
		pkt, err = s.conn.Receive()
		if err != nil {
			if !s.tomb.Alive() {
				return
			}
			s.log.WithError(err).Warnf("failed to reveive message")
			s.close(true)
			return
		}
		if _, ok := pkt.(*packet.Connect); !ok && s.authorizer == nil {
			s.log.Errorf("only connect packet is allowed before auth")
			s.close(true)
			return
		}
		switch p := pkt.(type) {
		case *packet.Connect:
			s.log.Debugln("received:", p.Type())
			err = s.onConnect(p)
		case *packet.Publish:
			s.log.Debugf("received: %s, pid: %d, qos: %d, topic: %s", p.Type(), p.ID, p.Message.QOS, p.Message.Topic)
			err = s.onPublish(p)
		case *packet.Puback:
			s.log.Debugf("received: %s, pid: %d", p.Type(), p.ID)
			err = s.onPuback(p)
		case *packet.Subscribe:
			s.log.Debugf("received: %s, subs: %v", p.Type(), p.Subscriptions)
			err = s.onSubscribe(p)
		case *packet.Pingreq:
			s.log.Debugln("received:", p.Type())
			err = s.onPingreq(p)
		case *packet.Pingresp:
			s.log.Debugln("received:", p.Type())
			err = nil // just ignore
		case *packet.Disconnect:
			s.log.Debugln("received:", p.Type())
			s.close(false)
			return
		case *packet.Unsubscribe:
			s.log.Debugf("received: %s, topics: %v", p.Type(), p.Topics)
			err = s.onUnsubscribe(p)
		default:
			err = fmt.Errorf("packet (%v) not supported", p)
		}
		if err != nil {
			s.log.Errorf(err.Error())
			s.close(true)
			break
		}
	}
}

func (s *session) onConnect(p *packet.Connect) error {
	s.log = s.log.WithField("client", p.ClientID)
	if p.Version != packet.Version31 && p.Version != packet.Version311 {
		s.sendConnack(packet.InvalidProtocolVersion)
		return fmt.Errorf("MQTT protocol version (%d) invalid", p.Version)
	}
	// username must set
	if p.Username == "" {
		s.sendConnack(packet.BadUsernameOrPassword)
		return fmt.Errorf("username not set")
	}
	if p.Password != "" {
		// if password is set, to use account auth
		s.authorizer = s.manager.auth.authenticateAccount(p.Username, p.Password)
		if s.authorizer == nil {
			s.sendConnack(packet.BadUsernameOrPassword)
			return fmt.Errorf("username (%s) or password not permitted", p.Username)
		}
	} else if mqtt.IsTwoWayTLS(s.conn) {
		// if it is two-way tls, to use cert auth
		s.authorizer = s.manager.auth.authenticateCert(p.Username)
		if s.authorizer == nil {
			s.sendConnack(packet.BadUsernameOrPassword)
			return fmt.Errorf("username (%s) is not permitted over tls", p.Username)
		}
	} else {
		s.sendConnack(packet.BadUsernameOrPassword)
		return fmt.Errorf("password not set")
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
		if !s.authorizer.authorize(authPublish, p.Will.Topic) {
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
		if !s.authorizer.authorize(authPublish, p.Message.Topic) {
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
	// s.log.Debugf("receive puback: pid=%d", p.ID)
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
			s.log.Infof("topic (%s) is unsubscribed", topic)
		} else {
			s.log.Warnf("topic (%s) is not subscribed yet", topic)
		}
	}
	ack.ID = p.ID
	return s.send(ack, true)
}

func (s *session) onPingreq(p *packet.Pingreq) error {
	return s.send(packet.NewPingresp(), true)
}

func (s *session) callback(pid uint32) {
	ack := packet.NewPuback()
	ack.ID = packet.ID(pid)
	err := s.send(ack, true)
	if err != nil {
		s.close(true)
		return
	}
	s.log.Debugf("puback(pid=%d) sent", pid)
}
