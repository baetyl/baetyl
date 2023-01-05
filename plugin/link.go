package plugin

import (
	"fmt"

	"github.com/baetyl/baetyl-go/v2/spec/v1"
)

const (
	LinkStateSucceeded    = "Succeeded"
	LinkStateNodeNotFound = "NodeNotFound"
	LinkStateNetworkError = "NetworkError"
	LinkStateUnknown      = "Unknown"
)

var (
	// ErrLinkTLSConfigMissing certificate bidirectional authentication is required for connection with cloud
	ErrLinkTLSConfigMissing = fmt.Errorf("certificate bidirectional authentication is required for connection with cloud")
	ConfFile                string
)

//go:generate mockgen -destination=../mock/plugin/link.go -package=plugin -source=link.go Link

type Link interface {
	State() *v1.Message
	Receive() (<-chan *v1.Message, <-chan error)
	Request(msg *v1.Message) (*v1.Message, error)
	Send(msg *v1.Message) error
	IsAsyncSupported() bool
}
