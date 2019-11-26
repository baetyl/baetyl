package session

import (
	"github.com/256dpi/gomqtt/packet"
	"github.com/baetyl/baetyl/baetyl-hub/common"
)

func (s *session) publish(msg common.Message) {
	pub := new(packet.Publish)
	pub.Message.QOS = packet.QOS(msg.TargetQOS)
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
	s.log.Debugf("message (pid=%d) sent", pub.ID)
}

func (s *session) republish(msg common.Message) {
	if msg.TargetQOS != 1 {
		s.log.Errorf("unexpcted: qos must be 1")
	}
	pid := s.pids.Get(msg.SequenceID)
	if pid == 0 {
		s.log.Errorf("failed to find packet to republish")
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
	s.log.Debugf("message (pid=%d) resent", pid)
}
