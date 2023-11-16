package prober

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"k8s.io/component-base/version"
	utilio "k8s.io/utils/io"
)

const (
	maxRespBodyLength = 10 * 1 << 10 // 10KB
)

// NewHTTPProber creates Prober that will skip TLS verification while probing.
// followNonLocalRedirects configures whether the prober should follow redirects to a different hostname.
//
//	If disabled, redirects to other hosts will trigger a warning result.
func NewHTTPProber(followNonLocalRedirects bool) HTTPProber {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	return NewWithTLSConfig(tlsConfig, followNonLocalRedirects)
}

// NewWithTLSConfig takes tls config as parameter.
// followNonLocalRedirects configures whether the prober should follow redirects to a different hostname.
//
//	If disabled, redirects to other hosts will trigger a warning result.
func NewWithTLSConfig(config *tls.Config, followNonLocalRedirects bool) HTTPProber {
	// We do not want the probe use node's local proxy set.
	transport := &http.Transport{
		TLSClientConfig:   config,
		DisableKeepAlives: true,
		Proxy:             http.ProxyURL(nil),
	}
	return httpProber{transport, followNonLocalRedirects}
}

// HTTPProber is an interface that defines the Probe function for doing HTTP readiness/liveness checks.
type HTTPProber interface {
	Probe(url *url.URL, headers http.Header, timeout time.Duration) (ProbeResult, string, error)
}

type httpProber struct {
	transport               *http.Transport
	followNonLocalRedirects bool
}

// Probe returns a ProbeRunner capable of running an HTTP check.
func (pr httpProber) Probe(url *url.URL, headers http.Header, timeout time.Duration) (ProbeResult, string, error) {
	pr.transport.DisableCompression = true // removes Accept-Encoding header
	client := &http.Client{
		Timeout:       timeout,
		Transport:     pr.transport,
		CheckRedirect: redirectChecker(pr.followNonLocalRedirects),
	}
	return DoHTTPProbe(url, headers, client)
}

// GetHTTPInterface is an interface for making HTTP requests, that returns a response and error.
type GetHTTPInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

// DoHTTPProbe checks if a GET request to the url succeeds.
// If the HTTP response code is successful (i.e. 400 > code >= 200), it returns Success.
// If the HTTP response code is unsuccessful or HTTP communication fails, it returns Failure.
// This is exported because some other packages may want to do direct HTTP probes.
func DoHTTPProbe(url *url.URL, headers http.Header, client GetHTTPInterface) (ProbeResult, string, error) {
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		// Convert errors into failures to catch timeouts.
		return Failure, err.Error(), nil
	}
	if headers == nil {
		headers = http.Header{}
	}
	if _, ok := headers["User-Agent"]; !ok {
		// explicitly set User-Agent so it's not set to default Go value
		v := version.Get()
		headers.Set("User-Agent", fmt.Sprintf("kube-probe/%s.%s", v.Major, v.Minor))
	}
	if _, ok := headers["Accept"]; !ok {
		// Accept header was not defined. accept all
		headers.Set("Accept", "*/*")
	} else if headers.Get("Accept") == "" {
		// Accept header was overridden but is empty. removing
		headers.Del("Accept")
	}
	req.Header = headers
	req.Host = headers.Get("Host")
	res, err := client.Do(req)
	if err != nil {
		// Convert errors into failures to catch timeouts.
		return Failure, err.Error(), nil
	}
	defer res.Body.Close()
	b, err := utilio.ReadAtMost(res.Body, maxRespBodyLength)
	if err != nil {
		if err == utilio.ErrLimitReached {
			log.L().Debug(fmt.Sprintf("Non fatal body truncation for %s, Response: %v", url.String(), *res))
		} else {
			return Failure, "", err
		}
	}
	body := string(b)
	if res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusBadRequest {
		if res.StatusCode >= http.StatusMultipleChoices { // Redirect
			log.L().Warn(fmt.Sprintf("Probe succeeded for %s, Response: %v", url.String(), *res))
			return Warning, body, nil
		}
		log.L().Debug(fmt.Sprintf("Probe succeeded for %s, Response: %v", url.String(), *res))
		return Success, body, nil
	}
	log.L().Warn(fmt.Sprintf("Probe failed for %s with request headers %v, response body: %v", url.String(), headers, body))
	return Failure, fmt.Sprintf("HTTP probe failed with statuscode: %d", res.StatusCode), nil
}

func redirectChecker(followNonLocalRedirects bool) func(*http.Request, []*http.Request) error {
	if followNonLocalRedirects {
		return nil // Use the default http client checker.
	}

	return func(req *http.Request, via []*http.Request) error {
		if req.URL.Hostname() != via[0].URL.Hostname() {
			return http.ErrUseLastResponse
		}
		// Default behavior: stop after 10 redirects.
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}
}
