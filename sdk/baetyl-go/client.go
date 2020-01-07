package baetyl

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/baetyl/baetyl/protocol/http"
	"github.com/baetyl/baetyl/sdk/baetyl-go/api"
	"google.golang.org/grpc"
)

// HTTPClient client of http server
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
	c := http.ClientInfo{
		Address:  addr,
		Username: name,
		Password: token,
	}
	cli, err := NewClient(c, version)
	if err != nil {
		return nil, err
	}

	var gcli *api.Client
	if len(grpcAddr) != 0 {
		cc := api.ClientConfig{
			Address:  grpcAddr,
			Username: name,
			Password: token,
		}
		gcli, err = api.NewClient(cc)
		if err != nil {
			return nil, err
		}
	}
	return &Client{
		Client:     gcli,
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

// SetKV set kv
func (c *Client) SetKV(kv api.KV) error {
	_, err := c.KV.Set(context.Background(), &kv, grpc.WaitForReady(true))
	return err
}

// SetKVConext set kv which supports context
func (c *Client) SetKVConext(ctx context.Context, kv api.KV) error {
	_, err := c.KV.Set(ctx, &kv, grpc.WaitForReady(true))
	return err
}

// GetKV get kv
func (c *Client) GetKV(k []byte) (*api.KV, error) {
	return c.KV.Get(context.Background(), &api.KV{Key: k}, grpc.WaitForReady(true))
}

// GetKVConext get kv which supports context
func (c *Client) GetKVConext(ctx context.Context, k []byte) (*api.KV, error) {
	return c.KV.Get(ctx, &api.KV{Key: k}, grpc.WaitForReady(true))
}

// DelKV del kv
func (c *Client) DelKV(k []byte) error {
	_, err := c.KV.Del(context.Background(), &api.KV{Key: k}, grpc.WaitForReady(true))
	return err
}

// DelKVConext del kv which supports context
func (c *Client) DelKVConext(ctx context.Context, k []byte) error {
	_, err := c.KV.Del(ctx, &api.KV{Key: k}, grpc.WaitForReady(true))
	return err
}

// ListKV list kv with prefix
func (c *Client) ListKV(p []byte) ([]*api.KV, error) {
	kvs, err := c.KV.List(context.Background(), &api.KV{Key: p}, grpc.WaitForReady(true))
	if err != nil {
		return nil, err
	}
	return kvs.Kvs, nil
}

// ListKVContext list kv with prefix which supports context
func (c *Client) ListKVContext(ctx context.Context, p []byte) ([]*api.KV, error) {
	kvs, err := c.KV.List(ctx, &api.KV{Key: p}, grpc.WaitForReady(true))
	if err != nil {
		return nil, err
	}
	return kvs.Kvs, nil
}
