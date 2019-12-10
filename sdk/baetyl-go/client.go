package baetyl

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/baetyl/baetyl/protocol/http"
	"github.com/baetyl/baetyl/sdk/baetyl-go/api"
)

// HttpClient client of http server
// Deprecated: Use api.Client instead.
type HTTPClient struct {
	cli *http.Client
	ver string
}

// Client client of api server
type Client struct {
	*api.Client
	*HTTPClient
}

// NewEnvClient creates a new client by env
func NewEnvClient() (*Client, error) {
	addr := os.Getenv(EnvKeyMasterAPIAddress)
	grpcAddr := os.Getenv(EnvKeyMasterGRPCAPIAddress)
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
	if len(grpcAddr) == 0 {
		return nil, fmt.Errorf("Env (%s) not found", EnvKeyMasterGRPCAPIAddress)
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

	cc := api.ClientConfig{
		Address:  grpcAddr,
		Username: name,
		Password: token,
	}
	api, err := api.NewClient(cc)
	if err != nil {
		return nil, err
	}
	return &Client{
		Client:     api,
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

// SetKV set kv
func (c *Client) SetKV(key, value []byte) error {
	ctx := context.Background()
	kv := &api.KV{
		Key:   key,
		Value: value,
	}
	_, err := c.KV.Set(ctx, kv)
	return err
}

// GetKV get kv
func (c *Client) GetKV(key []byte) ([]byte, error) {
	ctx := context.Background()
	kv := &api.KV{
		Key: key,
	}
	kv, err := c.KV.Get(ctx, kv)
	return kv.Value, err
}

// DelKV del kv
func (c *Client) DelKV(key []byte) error {
	ctx := context.Background()
	kv := &api.KV{
		Key: key,
	}
	_, err := c.KV.Del(ctx, kv)
	return err
}

// ListKV list kv with prefix
func (c *Client) ListKV(key []byte) ([][]byte, error) {
	ctx := context.Background()
	kv := &api.KV{
		Key: key,
	}
	var values [][]byte
	kvs, err := c.KV.List(ctx, kv)
	for _, v := range kvs.Kvs {
		values = append(values, v.Value)
	}
	return values, err
}
