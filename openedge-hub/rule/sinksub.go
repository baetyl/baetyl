package rule

import (
	"github.com/baidu/openedge/openedge-hub/common"
)

// sinksub subscription of sink
type sinksub struct {
	id      string
	qos     uint32
	topic   string
	tqos    uint32
	ttopic  string
	channel *msgchan
}

// Newsinksub creates a new subscription of sink
func newSinkSub(subid string, subqos uint32, subtopic string, pubqos uint32, pubtopic string, channel *msgchan) *sinksub {
	return &sinksub{
		id:      subid,
		qos:     subqos,
		topic:   subtopic,
		tqos:    pubqos,
		ttopic:  pubtopic,
		channel: channel,
	}
}

// ID returns id of sinksub
func (s *sinksub) ID() string {
	return s.id
}

// QOS returns qos of sinksub
func (s *sinksub) QOS() uint32 {
	return s.qos
}

// Topic returns topic of sinksub
func (s *sinksub) Topic() string {
	return s.topic
}

// TargetQOS returns target qos of sinksub
func (s *sinksub) TargetQOS() uint32 {
	return s.tqos
}

// TargetTopic returns target topic of sinksub
func (s *sinksub) TargetTopic() string {
	return s.ttopic
}

// Flow flows message
func (s *sinksub) Flow(msg common.Message) {
	// set target topic
	if s.ttopic != "" {
		msg.TargetTopic = s.ttopic
	} else {
		msg.TargetTopic = msg.Topic
	}
	// qos only can decrease without increasing
	sqos := s.qos
	if msg.QOS < sqos {
		sqos = msg.QOS
	}
	if sqos == 0 {
		msg.TargetQOS = 0
		s.channel.putQ0(&msg)
	} else {
		msg.TargetQOS = s.tqos
		s.channel.putQ1(&msg)
	}
}
