package engine

import (
	"net"
	"time"

	openedge "github.com/baidu/openedge/api/go"
)

// Service is a running instance of module
type Service interface {
	Info() *openedge.ServiceInfo
	Instances() []Instance
	Scale(replica int, grace time.Duration) error
	Stop(grace time.Duration) error
}

// Instance data of service
type Instance struct {
	ID      string
	Addr    net.Addr
	Service Service
}
