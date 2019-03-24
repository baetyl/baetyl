package http

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

	var url *url.URL
	if c.Address != "" {
		url, err = utils.ParseURL(c.Address)
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
	return c.SendPath("GET", fmt.Sprintf(path, params...), nil, c.genHeader())
}

// Put sends put request
func (c *Client) Put(body []byte, path string, params ...interface{}) ([]byte, error) {
	return c.SendPath("PUT", fmt.Sprintf(path, params...), body, c.genHeader())
}

// Post sends post request
func (c *Client) Post(body []byte, path string, params ...interface{}) ([]byte, error) {
	return c.SendPath("POST", fmt.Sprintf(path, params...), body, c.genHeader())
}

// SendPath sends http request by path
func (c *Client) SendPath(method, path string, body []byte, header map[string]string) ([]byte, error) {
	url := fmt.Sprintf("%s://%s%s", c.url.Scheme, c.url.Host, path)
	res, err := c.SendUrl(method, url, bytes.NewBuffer(body), header)
	if err != nil {
		return nil, err
	}
	var resBody []byte
	if res != nil {
		defer res.Close()
		resBody, err = ioutil.ReadAll(res)
		if err != nil {
			return nil, err
		}
	}
	return resBody, nil
}

// SendUrl sends http request by url
func (c *Client) SendUrl(method, url string, body io.Reader, header map[string]string) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header = Headers{}
	if header != nil {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}
	res, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		var resBody []byte
		if res.Body != nil {
			defer res.Body.Close()
			resBody, _ = ioutil.ReadAll(res.Body)
		}
		return nil, fmt.Errorf("[%d] %s", res.StatusCode, strings.TrimRight(string(resBody), "\n"))
	}
	return res.Body, nil
}

func (c *Client) genHeader() map[string]string {
	header := map[string]string{"Content-Type": "application/json"}
	if c.cfg.Username != "" {
		header[headerKeyUsername] = c.cfg.Username
	}
	if c.cfg.Password != "" {
		header[headerKeyPassword] = c.cfg.Password
	}
	return header
}
