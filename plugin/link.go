package plugin

import "fmt"

var (
	// ErrSyncTLSConfigMissing certificate bidirectional authentication is required for connection with cloud
	ErrLinkTLSConfigMissing = fmt.Errorf("certificate bidirectional authentication is required for connection with cloud")
)

const (
	ReportKind = "report-kind"
	DesireKind = "desire-kind"
)

type Message struct {
	Kind     string            `yaml:"kind" json:"kind"`
	Metadata map[string]string `yaml:"meta" json:"meta"`
	Content  interface{}       `yaml:"content" json:"content"`
}

//go:generate mockgen -destination=../mock/plugin/link.go -package=plugin -source=link.go
type Link interface {
	Receive() (<-chan *Message, <-chan error)
	Request(msg *Message) (*Message, error)
	Send(msg *Message) error
	IsAsyncSupported() bool
}
