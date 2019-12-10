package api

import (
	"context"

	"github.com/baetyl/baetyl-go/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// NewClient creates a new client
func NewClient(conf ClientConfig) (*Client, error) {
	utils.SetDefaults(&conf)
	ctx, cel := context.WithTimeout(context.Background(), conf.Timeout)
	defer cel()

	opts := []grpc.DialOption{
		grpc.WithBlock(),
	}

	if conf.Key != "" || conf.Cert != "" {
		tlsCfg, err := utils.NewTLSConfigClient(conf.Certificate)
		if err != nil {
			return nil, err
		}
		if !conf.InsecureSkipVerify {
			tlsCfg.ServerName = conf.Name
		}
		creds := credentials.NewTLS(tlsCfg)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
		opts = append(opts, grpc.WithPerRPCCredentials(&passwordCred{
			Data: map[string]string{
				headerKeyUsername: conf.Username,
				headerKeyPassword: conf.Password,
			},
		}))
	}

	conn, err := grpc.DialContext(ctx, conf.Address, opts...)
	if err != nil {
		return nil, err
	}
	kv := NewKVServiceClient(conn)
	return &Client{
		conf: conf,
		conn: conn,
		KV:   kv,
	}, nil
}

// Close closes the client
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

type passwordCred struct {
	Data map[string]string
}

// GetRequestMetadata gets the current request metadata, refreshing tokens if required
func (c *passwordCred) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return c.Data, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security
func (c *passwordCred) RequireTransportSecurity() bool {
	return false
}
