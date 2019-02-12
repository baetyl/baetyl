package engine

import (
	"io"

	"github.com/baidu/openedge/sdk-go/openedge"
)

// Service interfaces of service
type Service interface {
	Name() string
	Stats() openedge.ServiceStatus
	Stop()
}

// Instance interfaces of instance
type Instance interface {
	ID() string
	Name() string
	Supervisee
	io.Closer
}
