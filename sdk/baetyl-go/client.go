package baetyl

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/baetyl/baetyl/protocol/http"
)

// HTTPClient client of http server
// Deprecated: Use api.Client instead.
type HTTPClient struct {
	cli *http.Client
	ver string
}

// Client client of api server
type Client struct {
	*HTTPClient
}

// NewEnvClient creates a new client by env
func NewEnvClient() (*Client, error) {
	addr := os.Getenv(EnvKeyMasterAPIAddress)
	name := os.Getenv(EnvKeyServiceName)
	token := os.Getenv(EnvKeyServiceToken)
	version := os.Getenv(EnvKeyMasterAPIVersion)
	if len(addr) == 0 {
		// TODO: remove, backward compatibility
		addr = os.Getenv(EnvMasterAPIKey)
		if len(addr) == 0 {
			return nil, fmt.Errorf("Env (%s) not found", EnvKeyMasterAPIAddress)
		}
		name = os.Getenv(EnvServiceNameKey)
		token = os.Getenv(EnvServiceTokenKey)
		version = os.Getenv(EnvMasterAPIVersionKey)
	}
	c := http.ClientInfo{
		Address:  addr,
		Username: name,
		Password: token,
	}
	cli, err := NewClient(c, version)
	if err != nil {
		return nil, err
	}

	return &Client{
		HTTPClient: cli,
	}, nil
}

// NewClient creates a new client
func NewClient(c http.ClientInfo, ver string) (*HTTPClient, error) {
	cli, err := http.NewClient(c)
	if err != nil {
		return nil, err
	}
	if ver != "" && !strings.HasPrefix(ver, "/") {
		ver = "/" + ver
	}
	return &HTTPClient{
		cli: cli,
		ver: ver,
	}, nil
}

// UpdateSystem updates and reloads config
func (c *Client) UpdateSystem(trace, tp, path string) error {
	data, err := json.Marshal(map[string]string{
		"type":  tp,
		"path":  path,
		"file":  path, // backward compatibility, master version < 0.1.4
		"trace": trace,
	})
	if err != nil {
		return err
	}
	_, err = c.cli.Put(data, c.ver+"/system/update")
	return err
}

// InspectSystem inspect all stats
func (c *Client) InspectSystem() (*Inspect, error) {
	body, err := c.cli.Get(c.ver + "/system/inspect")
	if err != nil {
		return nil, err
	}
	s := new(Inspect)
	err = json.Unmarshal(body, s)
	return s, err
}

// GetAvailablePort gets available port
func (c *Client) GetAvailablePort() (string, error) {
	res, err := c.cli.Get(c.ver + "/ports/available")
	if err != nil {
		return "", err
	}
	info := make(map[string]string)
	err = json.Unmarshal(res, &info)
	if err != nil {
		return "", err
	}
	port, ok := info["port"]
	if !ok {
		return "", fmt.Errorf("invalid response, port not found")
	}
	return port, nil
}

// ReportInstance reports the stats of the instance of the service
func (c *Client) ReportInstance(serviceName, instanceName string, stats map[string]interface{}) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	_, err = c.cli.Put(data, c.ver+"/services/%s/instances/%s/report", serviceName, instanceName)
	return err
}

// StartInstance starts a new service instance with dynamic config
func (c *Client) StartInstance(serviceName, instanceName string, dynamicConfig map[string]string) error {
	data, err := json.Marshal(dynamicConfig)
	if err != nil {
		return err
	}
	_, err = c.cli.Put(data, c.ver+"/services/%s/instances/%s/start", serviceName, instanceName)
	return err
}

// StopInstance stops a service instance
func (c *Client) StopInstance(serviceName, instanceName string) error {
	_, err := c.cli.Put(nil, c.ver+"/services/%s/instances/%s/stop", serviceName, instanceName)
	return err
}
