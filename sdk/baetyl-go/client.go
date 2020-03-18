package baetyl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

func (c *ctximpl) setupMaster() error {
	c.m.Transport = &http.Transport{
		TLSClientConfig: c.tls,
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return net.Dial("unix", c.cfg.LegacyAPIServer.Address)
		},
	}
	c.m.Timeout = time.Duration(c.cfg.LegacyAPIServer.TimeoutInSeconds)
	return nil
}

/*
// UpdateSystem updates and reloads config
func (c *ctximpl) UpdateSystem(trace, tp, path string) error {
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
func (c *ctximpl) InspectSystem() (*Inspect, error) {
	body, err := c.cli.Get(c.ver + "/system/inspect")
	if err != nil {
		return nil, err
	}
	s := new(Inspect)
	err = json.Unmarshal(body, s)
	return s, err
}
*/
// GetAvailablePort gets available port
func (c *ctximpl) GetAvailablePort() (string, error) {
	resp, err := c.m.Get(c.cfg.LegacyAPIServer.Address + "/ports/available")
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}
	info := make(map[string]string)
	err = json.Unmarshal(data, &info)
	if err != nil {
		return "", err
	}
	port, ok := info["port"]
	if !ok {
		return "", fmt.Errorf("invalid response, port not found")
	}
	return port, nil
}

/*
// ReportInstance reports the stats of the instance of the service
func (c *ctximpl) ReportInstance(serviceName, instanceName string, stats map[string]interface{}) error {
	data, err := json.Marshal(stats)
	if err != nil {
		return err
	}
	_, err = c.cli.Put(data, c.ver+"/services/%s/instances/%s/report", serviceName, instanceName)
	return err
}

// StartInstance starts a new service instance with dynamic config
func (c *ctximpl) StartInstance(serviceName, instanceName string, dynamicConfig map[string]string) error {
	data, err := json.Marshal(dynamicConfig)
	if err != nil {
		return err
	}
	_, err = c.cli.Put(data, c.ver+"/services/%s/instances/%s/start", serviceName, instanceName)
	return err
}

// StopInstance stops a service instance
func (c *ctximpl) StopInstance(serviceName, instanceName string) error {
	_, err := c.cli.Put(nil, c.ver+"/services/%s/instances/%s/stop", serviceName, instanceName)
	return err
}

// SetKV set kv
func (c *ctximpl) SetKV(kv apiserver.KV) error {
	_, err := c.KV.Set(context.Background(), &kv, grpc.WaitForReady(true))
	return err
}

// SetKVConext set kv which supports context
func (c *ctximpl) SetKVConext(ctx context.Context, kv apiserver.KV) error {
	_, err := c.KV.Set(ctx, &kv, grpc.WaitForReady(true))
	return err
}

// GetKV get kv
func (c *ctximpl) GetKV(k []byte) (*apiserver.KV, error) {
	return c.KV.Get(context.Background(), &apiserver.KV{Key: k}, grpc.WaitForReady(true))
}

// GetKVConext get kv which supports context
func (c *ctximpl) GetKVConext(ctx context.Context, k []byte) (*apiserver.KV, error) {
	return c.KV.Get(ctx, &apiserver.KV{Key: k}, grpc.WaitForReady(true))
}

// DelKV del kv
func (c *ctximpl) DelKV(k []byte) error {
	_, err := c.KV.Del(context.Background(), &apiserver.KV{Key: k}, grpc.WaitForReady(true))
	return err
}

// DelKVConext del kv which supports context
func (c *ctximpl) DelKVConext(ctx context.Context, k []byte) error {
	_, err := c.KV.Del(ctx, &apiserver.KV{Key: k}, grpc.WaitForReady(true))
	return err
}

// ListKV list kv with prefix
func (c *ctximpl) ListKV(p []byte) ([]*apiserver.KV, error) {
	kvs, err := c.KV.List(context.Background(), &apiserver.KV{Key: p}, grpc.WaitForReady(true))
	if err != nil {
		return nil, err
	}
	return kvs.Kvs, nil
}

// ListKVContext list kv with prefix which supports context
func (c *ctximpl) ListKVContext(ctx context.Context, p []byte) ([]*apiserver.KV, error) {
	kvs, err := c.KV.List(ctx, &apiserver.KV{Key: p}, grpc.WaitForReady(true))
	if err != nil {
		return nil, err
	}
	return kvs.Kvs, nil
}
*/
