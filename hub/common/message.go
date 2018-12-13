package common

import (
	"time"

	"github.com/baidu/openedge/hub/utils"
	"github.com/golang/protobuf/proto"
)

// Q0: it means the message with qos=0 published by clients or functions
// Q1: it means the message with qos=1 published by clients or functions
// Cb: it means the callback which called after message handled.
//     For message with qos=1 published by client, the callback used to send puback
// Ack: it means the acknowledgement which acknowledged after message acknowledged.
//      For message with qos=1 sent to client, it will acknowledged if client send back puback

// Flow flows message
type Flow func(*Message)

// Publish publishes message
type Publish func(Message)

// Message MQTT message with client ID
type Message struct {
	Persisted
	TargetQOS   uint32
	TargetTopic string
	Barrier     bool
	Retain      bool
	PacketID    uint32
	callbackPID func(uint32) // callback for PacketID
	SequenceID  uint64
	callbackSID func(uint64) // callback for SequenceID
	acknowledge *Acknowledge
}

// NewMessage creates a message
func NewMessage(qos uint32, topic string, payload []byte, clientID string) *Message {
	return &Message{
		Persisted: Persisted{
			QOS:      qos,
			Topic:    topic,
			Payload:  payload,
			ClientID: clientID,
		},
	}
}

// UnmarshalMessage creates a message by persisted data
func UnmarshalMessage(k, v []byte) (*Message, error) {
	msg := &Message{SequenceID: utils.U64(k)}
	err := proto.Unmarshal(v, &msg.Persisted)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// SetCallbackPID sets packet id and its callback
func (m *Message) SetCallbackPID(pid uint32, callback func(uint32)) {
	m.PacketID = pid
	m.callbackPID = callback
}

// CallbackPID calls the callback with packet id
func (m *Message) CallbackPID() {
	if m.callbackPID != nil {
		m.callbackPID(m.PacketID)
	}
}

// SetCallbackSID sets sequence id and its callback
func (m *Message) SetCallbackSID(callback func(uint64)) {
	m.callbackSID = callback
}

// SetAcknowledge sets acknowledge
func (m *Message) SetAcknowledge() {
	m.acknowledge = NewAcknowledge()
}

// SID returns sequence ID
func (m *Message) SID() uint64 {
	return m.SequenceID
}

// Ack acknowledges after message handled
func (m *Message) Ack() {
	if m.acknowledge != nil {
		m.acknowledge.Ack()
	}
}

// WaitTimeout waits until finish
func (m *Message) WaitTimeout(timeout time.Duration, republish Publish, cancel <-chan struct{}) {
	if m.acknowledge == nil {
		if m.callbackSID != nil {
			m.callbackSID(m.SequenceID)
		}
		return
	}
	for {
		select {
		case <-cancel:
			return
		case <-m.acknowledge.done:
			if m.callbackSID != nil {
				m.callbackSID(m.SequenceID)
			}
			return
		case <-time.After(timeout):
			if republish != nil {
				republish(*m)
			}
		}
	}
}
