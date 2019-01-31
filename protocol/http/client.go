package http

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/baidu/openedge/utils"
	"github.com/creasty/defaults"
	"github.com/docker/go-connections/sockets"
)

const (
	headerKeyUsername = "x-openedge-username"
	headerKeyPassword = "x-openedge-password"
)

var errAccountUnauthorized = errors.New("account unauthorized")

// Client client of http server
type Client struct {
	cli *http.Client
	url *url.URL
	cfg ClientInfo
}

// NewClient creates a new http client
func NewClient(c ClientInfo) (*Client, error) {
	defaults.Set(&c)

	tls, err := utils.NewTLSClientConfig(c.Certificate)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{
		TLSClientConfig: tls,
	}
	url, err := utils.ParseURL(c.Address)
	if err != nil {
		return nil, err
	}
	sockets.ConfigureTransport(transport, url.Scheme, url.Host)
	if url.Scheme == "unix" {
		url.Host = "openedge"
	}
	if url.Scheme != "http" && url.Scheme != "https" {
		url.Scheme = "http"
	}
	return &Client{
		cfg: c,
		url: url,
		cli: &http.Client{
			Timeout:   c.Timeout,
			Transport: transport,
		},
	}, nil
}

// Get sends get request
func (c *Client) Get(path string, params ...interface{}) ([]byte, error) {
	return c.send("GET", fmt.Sprintf(path, params...), nil)
}

// Put sends put request
func (c *Client) Put(body []byte, path string, params ...interface{}) ([]byte, error) {
	return c.send("PUT", fmt.Sprintf(path, params...), body)
}

// Post sends post request
func (c *Client) Post(body []byte, path string, params ...interface{}) ([]byte, error) {
	return c.send("POST", fmt.Sprintf(path, params...), body)
}

func (c *Client) send(method, path string, body []byte) ([]byte, error) {
	url := fmt.Sprintf("%s://%s%s", c.url.Scheme, c.url.Host, path)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	headers := Headers{}
	headers.Set("Content-Type", "application/json")
	if c.cfg.Username != "" {
		headers.Set(headerKeyUsername, c.cfg.Username)
	}
	if c.cfg.Password != "" {
		headers.Set(headerKeyPassword, c.cfg.Password)
	}
	req.Header = headers
	res, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}
	var resBody []byte
	if res.Body != nil {
		defer res.Body.Close()
		resBody, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
	}
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("[%d] %s", res.StatusCode, strings.TrimRight(string(resBody), "\n"))
	}
	return resBody, nil
}
