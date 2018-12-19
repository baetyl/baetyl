package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/utils"
	"github.com/docker/go-connections/sockets"
)

// Client client of http server
type Client struct {
	*http.Client
	Addr *url.URL
	Conf *config.HTTPClient
}

// NewClient creates a new http client
func NewClient(cc config.HTTPClient) (*Client, error) {
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
func (c *Client) Send(method, url string, headers Headers, body []byte) (Headers, []byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, err
	}
	req.Header = headers
	res, err := c.Do(req)
	if err != nil {
		return nil, nil, err
	}
	var resBody []byte
	if res.Body != nil {
		defer res.Body.Close()
		resBody, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, nil, err
		}
	}
	if res.StatusCode >= 400 {
		errMessage := string(resBody)
		return nil, nil, fmt.Errorf("[%d] %s", res.StatusCode, strings.TrimRight(errMessage, "\n"))
	}
	return res.Header, resBody, nil
}
