package prober

import (
	"net"
	"strconv"
	"time"

	"github.com/baetyl/baetyl-go/v2/log"
)

// NewTCPProber creates Prober.
func NewTCPProber() TCPProber {
	return tcpProber{}
}

// TCPProber is an interface that defines the Probe function for doing TCP readiness/liveness checks.
type TCPProber interface {
	Probe(host string, port int, timeout time.Duration) (ProbeResult, string, error)
}

type tcpProber struct{}

// Probe returns a ProbeRunner capable of running an TCP check.
func (pr tcpProber) Probe(host string, port int, timeout time.Duration) (ProbeResult, string, error) {
	return DoTCPProbe(net.JoinHostPort(host, strconv.Itoa(port)), timeout)
}

// DoTCPProbe checks that a TCP socket to the address can be opened.
// If the socket can be opened, it returns Success
// If the socket fails to open, it returns Failure.
// This is exported because some other packages may want to do direct TCP probes.
func DoTCPProbe(addr string, timeout time.Duration) (ProbeResult, string, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		// Convert errors to failures to handle timeouts.
		return Failure, err.Error(), nil
	}
	err = conn.Close()
	if err != nil {
		log.L().Error("Unexpected error closing TCP probe socket", log.Error(err))
	}
	return Success, "", nil
}
