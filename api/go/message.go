package openedge

// QoS constants
const (
	QoSAtMostOnce byte = 0
	QoSAtLeastOnce
	QoSExactOnce
)

// Message to send/receive
type Message struct {
	Topic   string `yaml:"topic" json:"topic"`
	QoS     byte   `yaml:"qos" json:"qos"`
	Payload []byte `yaml:"payload" json:"payload"`
}
