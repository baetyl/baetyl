package plugin

import (
	"fmt"
	"github.com/baetyl/baetyl-go/v2/spec/v1"
)

var (
	// ErrSyncTLSConfigMissing certificate bidirectional authentication is required for connection with cloud
	ErrLinkTLSConfigMissing = fmt.Errorf("certificate bidirectional authentication is required for connection with cloud")
	ConfFile                string
)

//go:generate mockgen -destination=../mock/plugin/link.go -package=plugin -source=link.go
type Link interface {
	Receive() (<-chan *v1.Message, <-chan error)
	Request(msg *v1.Message) (*v1.Message, error)
	Send(msg *v1.Message) error
	IsAsyncSupported() bool
}
