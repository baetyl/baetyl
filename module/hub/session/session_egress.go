package session

import (
	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/module/hub/common"
)

func (s *session) publish(msg common.Message) {
	pub := new(packet.Publish)
	pub.Message.QOS = byte(msg.TargetQOS)
	pub.Message.Topic = msg.TargetTopic
	pub.Message.Payload = msg.Payload
	pub.Message.Retain = msg.Retain
	if msg.TargetQOS == 1 {
		pid := s.pids.Set(&msg)
		pub.ID = packet.ID(pid)
	}
	if err := s.send(pub, true); err != nil {
		s.close(true)
	}
}

func (s *session) republish(msg common.Message) {
	if msg.TargetQOS != 1 {
		s.log.Errorf("Unexpcted: qos must be 1")
	}
	pid := s.pids.Get(msg.SequenceID)
	if pid == 0 {
		s.log.Errorf("Cannot find packet to re-publish")
		return
	}
	pub := new(packet.Publish)
	pub.ID = packet.ID(pid)
	pub.Dup = true
	pub.Message.QOS = 1
	pub.Message.Topic = msg.TargetTopic
	pub.Message.Payload = msg.Payload
	if err := s.send(pub, true); err != nil {
		s.close(true)
	}
	s.log.Debugf("Resend message (pid=%d)", pid)
}

// func (s *session) trace() func() {
// 	start := time.Now()
// 	return func() {
// 		s.log.Debugf("Send message elapsed time: %v", time.Since(start))
// 	}
// }
