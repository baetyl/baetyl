package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	openedge "github.com/baidu/openedge/api/go"

	"github.com/baidu/openedge/utils"
	"github.com/docker/go-connections/sockets"
)

// Client client of http server
type Client struct {
	*http.Client
	Addr *url.URL
	Conf *openedge.HTTPClientInfo
}

// NewClient creates a new http client
func NewClient(cc openedge.HTTPClientInfo) (*Client, error) {
	tls, err := utils.NewTLSClientConfig(cc.Certificate)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{
		TLSClientConfig: tls,
	}
	url, err := utils.ParseURL(cc.Address)
	if err != nil {
		return nil, err
	}
	sockets.ConfigureTransport(transport, url.Scheme, url.Host)
	if url.Scheme == "unix" {
		url.Host = "openedge"
	}
	url.Scheme = "http"
	return &Client{
		Client: &http.Client{
			Timeout:   cc.Timeout,
			Transport: transport,
		},
		Addr: url,
		Conf: &cc,
	}, nil
}

// Send sends request
func (c *Client) Send(method, url string, headers Headers, body io.Reader) (Headers, io.ReadCloser, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, nil, err
	}
	req.Header = headers
	res, err := c.Do(req)
	if err != nil {
		return nil, nil, err
	}
	if res.StatusCode >= 400 {
		count := 0
		message := make([]byte, 1024)
		if res.Body != nil {
			defer res.Body.Close()
			count, err = res.Body.Read(message)
			if err != nil {
				return nil, nil, err
			}
			message = message[:count]
		}
		return nil, nil, fmt.Errorf("[%d] %s", res.StatusCode, strings.TrimRight(string(message), "\n"))
	}
	return res.Header, res.Body, nil
}
