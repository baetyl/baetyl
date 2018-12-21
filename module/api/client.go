package api

import (
	"encoding/json"
	"fmt"

	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/http"
	"github.com/baidu/openedge/module/logger"
)

// Client client of api server
type Client struct {
	*http.Client
	log *logger.Entry
}

// NewClient creates a new client
func NewClient(cc config.HTTPClient) (*Client, error) {
	cli, err := http.NewClient(cc)
	if err != nil {
		return nil, err
	}
	return &Client{
		Client: cli,
		log:    logger.WithFields("client", "api"),
	}, nil
}

// GetPortAvailable gets an available port
func (c *Client) GetPortAvailable(host string) (int, error) {
	_, resBody, err := c.Send("GET", fmt.Sprintf("%s://%s/ports/available?host=%s", c.Addr.Scheme, c.Addr.Host, host), c.newHeaders(), nil)
	if err != nil {
		return 0, err
	}
	b := map[string]int{}
	err = json.Unmarshal(resBody, &b)
	if err != nil {
		return 0, err
	}
	return b["port"], nil
}

// StartModule starts a module
func (c *Client) StartModule(m *config.Module) error {
	body, err := json.Marshal(m)
	if err != nil {
		return err
	}
	_, _, err = c.Send("PUT", c.newURL(m.UniqueName(), "start"), c.newHeaders(), body)
	return err
}

// StopModule stops a module
func (c *Client) StopModule(m *config.Module) error {
	_, _, err := c.Send("PUT", c.newURL(m.UniqueName(), "stop"), c.newHeaders(), nil)
	return err
}

func (c *Client) newURL(name, action string) string {
	return fmt.Sprintf("%s://%s/modules/%s/%s", c.Addr.Scheme, c.Addr.Host, name, action)
}

func (c *Client) newHeaders() http.Headers {
	h := http.Headers{}
	h.Set("Content-Type", "application/json")
	h.Set("x-iot-edge-username", c.Conf.Username)
	h.Set("x-iot-edge-password", c.Conf.Password)
	return h
}
