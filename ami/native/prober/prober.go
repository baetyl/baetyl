package prober

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"k8s.io/api/core/v1"
)

// ProbeResult is a string used to handle the results for probing container readiness/liveness
type ProbeResult string

const (
	// Success ProbeResult
	Success ProbeResult = "success"
	// Warning ProbeResult. Logically success, but with additional debugging information attached.
	Warning ProbeResult = "warning"
	// Failure ProbeResult
	Failure ProbeResult = "failure"
	// Unknown ProbeResult
	Unknown ProbeResult = "unknown"

	maxProbeRetries = 3
	localhost       = "127.0.0.1"
)

// Prober helps to check the liveness/readiness/startup of a container.
type prober struct {
	// probe types needs different httpprobe instances so they don't
	// share a connection pool which can cause collisions to the
	// same host:port and transient failures. See #49740.
	livenessHTTP HTTPProber
	tcp          TCPProber
}

func newProber() *prober {
	const followNonLocalRedirects = false
	return &prober{
		livenessHTTP: NewHTTPProber(followNonLocalRedirects),
		tcp:          NewTCPProber(),
	}
}

func (pb *prober) probe(appName string, p *v1.Probe) (ProbeResult, error) {
	var err error
	var output string
	result := Unknown
	for i := 0; i < maxProbeRetries; i++ {
		result, output, err = pb.runProbe(appName, p)
		if err == nil {
			break
		}
	}
	if err != nil {
		log.L().Error("Probe errored", log.Error(err))
		return Failure, err
	}
	if result != Success && result != Warning {
		log.L().Error("Probe failed", log.Any("output", output))
		return Failure, err
	}
	if result == Warning {
		log.L().Warn("Probe succeeded with a warning", log.Any("output", output))
	}
	log.L().Debug("Probe succeeded")
	return Success, nil
}

func (pb *prober) runProbe(appName string, p *v1.Probe) (ProbeResult, string, error) {
	timeout := time.Duration(p.TimeoutSeconds) * time.Second
	if p.HTTPGet != nil {
		scheme := strings.ToLower(string(p.HTTPGet.Scheme))
		host := p.HTTPGet.Host
		if host == "" {
			host = localhost
		}
		port := p.HTTPGet.Port
		if port.String() == "" {
			return Unknown, "", errors.New("No port selected")
		}
		path := p.HTTPGet.Path
		log.L().Debug("HTTP-Probe Host", log.Any("host", host), log.Any("port", port), log.Any("path", path))
		url := formatURL(scheme, host, port.IntValue(), path)
		headers := buildHeader(p.HTTPGet.HTTPHeaders)
		return pb.livenessHTTP.Probe(url, headers, timeout)
	}
	if p.TCPSocket != nil {
		port := p.HTTPGet.Port
		if port.String() == "" {
			return Unknown, "", errors.New("No port selected")
		}
		host := p.TCPSocket.Host
		if host == "" {
			host = localhost
		}
		log.L().Debug("TCP-Probe Host", log.Any("host", host), log.Any("port", port))
		return pb.tcp.Probe(host, port.IntValue(), timeout)
	}
	return Unknown, "", fmt.Errorf("missing probe handler for %s", appName)
}

// formatURL formats a URL from args.  For testability.
func formatURL(scheme string, host string, port int, path string) *url.URL {
	u, err := url.Parse(path)
	// Something is busted with the path, but it's too late to reject it. Pass it along as is.
	if err != nil {
		u = &url.URL{
			Path: path,
		}
	}
	u.Scheme = scheme
	u.Host = net.JoinHostPort(host, strconv.Itoa(port))
	return u
}

// buildHeaderMap takes a list of HTTPHeader <name, value> string
// pairs and returns a populated string->[]string http.Header map.
func buildHeader(headerList []v1.HTTPHeader) http.Header {
	headers := make(http.Header)
	for _, header := range headerList {
		headers[header.Name] = append(headers[header.Name], header.Value)
	}
	return headers
}
