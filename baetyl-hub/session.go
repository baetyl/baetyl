package hub

import (
	"fmt"
	"sync"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
	"github.com/baetyl/baetyl/baetyl-hub/common"
	"github.com/baetyl/baetyl/baetyl-hub/router"
	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/utils"
)

// session session of a client
// ingress data flow: client -> session(onPublish) -> broker -> database -> session(Ack)
// egress data flow: broker(rule) -> session(doQ0/doQ1) -> client -> session(onPuback)
type session struct {
	id       string
	clean    bool
	clientID string
	conn     transport.Conn
	subs     map[string]packet.Subscription
	manager  *sessionManager
	pids     *common.PacketIDS
	log      logger.Logger
	once     sync.Once
	tomb     utils.Tomb
	sync.Mutex

	authorizer *authorizer
	//  cache
	permittedPublishTopics map[string]struct{}
}

func newSession(conn transport.Conn, manager *sessionManager) *session {
	return &session{
		conn:                   conn,
		manager:                manager,
		subs:                   make(map[string]packet.Subscription),
		pids:                   common.NewPacketIDS(),
		log:                    logger.WithField("mqtt", "session"),
		permittedPublishTopics: make(map[string]struct{}),
	}
}

func (s *session) send(p packet.Generic, async bool) error {
	s.Lock()
	defer s.Unlock()
	err := s.conn.Send(p, async)
	if err != nil {
		return fmt.Errorf("failed to send message: %s", err.Error())
	}
	return nil
}

func (s *session) sendConnack(code packet.ConnackCode) error {
	ack := packet.Connack{
		SessionPresent: false, // TODO: to support
		ReturnCode:     code,
	}
	return s.send(&ack, false)
}

func (s *session) saveWillMessage(p *packet.Connect) error {
	if p.Will == nil {
		return nil
	}
	return s.manager.recorder.setWill(s.id, p.Will)
}

// TODO: need to send will message after client reconnected if baetyl panicked
// Situations in which the Will Message is published include, but are not limited to:
// * An I/O error or network failure detected by the Server.
// * The Client fails to communicate within the Keep Alive time.
// * The Client closes the Network Connection without first sending a DISCONNECT Packet. The Server closes the Network Connection because of a protocol error.
func (s *session) sendWillMessage() {
	msg, err := s.manager.recorder.getWill(s.id)
	if err != nil {
		s.log.WithError(err).Errorf("failed to get will message")
	}
	if msg == nil {
		return
	}
	err = s.retainMessage(msg)
	if err != nil {
		s.log.WithError(err).Errorf("failed to retain will message")
	}
	s.manager.flow(common.NewMessage(uint32(msg.QOS), msg.Topic, msg.Payload, s.clientID))
}

func (s *session) retainMessage(msg *packet.Message) error {
	if len(msg.Payload) == 0 {
		return s.manager.recorder.removeRetained(msg.Topic)
	}
	return s.manager.recorder.setRetained(msg.Topic, msg)
}

// TODO: 存在安全问题？未做账号隔离？云端也存在这个问题
func (s *session) sendRetainMessage(p *packet.Subscribe) error {
	msgs, err := s.manager.recorder.getRetained()
	if err != nil || len(msgs) == 0 {
		return err
	}
	t := router.NewTrie()
	for _, sub := range p.Subscriptions {
		t.Add(router.NewNopSinkSub(s.id, uint32(sub.QOS), sub.Topic, uint32(sub.QOS), ""))
	}
	// TODO: improve and test, to resend if not acked?
	for _, msg := range msgs {
		if ok, qos := t.IsMatch(msg.Topic); ok {
			m := common.NewMessage(uint32(msg.QOS), msg.Topic, msg.Payload, s.clientID)
			if qos > m.QOS {
				m.TargetQOS = m.QOS
			} else {
				m.TargetQOS = qos
			}
			m.TargetTopic = msg.Topic
			m.Retain = true
			s.publish(*m)
		}
	}
	return nil
}

func (s *session) genSubAck(subs []packet.Subscription) []packet.QOS {
	rv := make([]packet.QOS, len(subs))
	for i, sub := range subs {
		if !common.SubTopicValidate(sub.Topic) {
			s.log.Errorf("subscribe topic (%s) invalid", sub.Topic)
			rv[i] = packet.QOSFailure
		} else if !s.authorizer.authorize(authSubscribe, sub.Topic) {
			s.log.Errorf("subscribe topic (%s) not permitted", sub.Topic)
			rv[i] = packet.QOSFailure
		} else if sub.QOS > 1 {
			s.log.Errorf("subscribe QOS (%d) not supported", sub.QOS)
			rv[i] = packet.QOSFailure
		} else {
			rv[i] = sub.QOS
		}
	}
	return rv
}

// Close closes this session, only called by session manager
func (s *session) close(will bool) {
	s.once.Do(func() {
		s.tomb.Kill(nil)
		s.log.Infof("session closing, messages (unack): %d", s.pids.Size())
		defer s.log.Infof("session closed, messages (unack): %d", s.pids.Size())
		s.manager.remove(s.id)
		if will {
			s.sendWillMessage()
		}
		s.conn.Close()
		s.manager.recorder.removeWill(s.id)
	})
}
