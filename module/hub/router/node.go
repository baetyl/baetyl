package router

import "github.com/baidu/openedge/module/hub/common"

// SinkSub subscription of sink
type SinkSub interface {
	ID() string // client id for session
	QOS() uint32
	Topic() string
	TargetQOS() uint32
	TargetTopic() string
	Flow(common.Message)
}

// NopSinkSub subscription of sink which does nothing
type NopSinkSub struct {
	id     string
	qos    uint32
	topic  string
	tqos   uint32
	ttopic string
}

// NewNopSinkSub creates a new subscription of sink which does nothing
func NewNopSinkSub(id string, qos uint32, topic string, tqos uint32, ttopic string) *NopSinkSub {
	return &NopSinkSub{id: id, qos: qos, topic: topic, tqos: tqos, ttopic: ttopic}
}

// ID returns the id of subscription
func (s *NopSinkSub) ID() string {
	return s.id
}

// QOS returns the qos of subscription
func (s *NopSinkSub) QOS() uint32 {
	return s.qos
}

// Topic returns the topic of subscribed topic
func (s *NopSinkSub) Topic() string {
	return s.topic
}

// TargetQOS returns the publish qos
func (s *NopSinkSub) TargetQOS() uint32 {
	return s.tqos
}

// TargetTopic returns the publish topic
func (s *NopSinkSub) TargetTopic() string {
	return s.ttopic
}

// Flow flows message to channel
func (s *NopSinkSub) Flow(common.Message) {
}

type node struct {
	children map[string]*node
	sinksubs map[string]SinkSub
}

func newNode() *node {
	return &node{
		children: make(map[string]*node),
		sinksubs: make(map[string]SinkSub),
	}
}

func (n *node) isEmpty() bool {
	return len(n.children) == 0 && len(n.sinksubs) == 0
}
