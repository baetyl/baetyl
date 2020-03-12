package store

import (
	"context"
	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-go/kv"
	"github.com/baetyl/baetyl-go/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type stateDriver struct {
	conn *grpc.ClientConn
	KV   kv.KVServiceClient
	ctx  context.Context
}

const StateDriverName = "state"

func NewStateDriver(cfg config.StateConfig) (Driver, error) {
	ctx := context.Background()
	opts := []grpc.DialOption{
		grpc.WithBlock(),
	}
	if cfg.Key != "" || cfg.Cert != "" {
		tlsCfg, err := utils.NewTLSConfigClient(cfg.Certificate)
		if err != nil {
			return nil, err
		}
		if !cfg.InsecureSkipVerify {
			tlsCfg.ServerName = cfg.Name
		}
		creds := credentials.NewTLS(tlsCfg)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.DialContext(ctx, cfg.Address, opts...)
	if err != nil {
		return nil, err
	}
	return &stateDriver{
		ctx:  context.Background(),
		conn: conn,
		KV:   kv.NewKVServiceClient(conn),
	}, nil
}

func (s *stateDriver) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func (s *stateDriver) Create(key []byte, val []byte) error {
	_, err := s.KV.Set(s.ctx, &kv.KV{Key: string(key), Value: val})
	if err != nil {
		return err
	}
	return nil
}

func (s *stateDriver) Update(key []byte, val []byte) error {
	_, err := s.KV.Set(s.ctx, &kv.KV{Key: string(key), Value: val})
	if err != nil {
		return err
	}
	return nil
}

func (s *stateDriver) Delete(key []byte) error {
	_, err := s.KV.Del(s.ctx, &kv.KV{Key: string(key)})
	if err != nil {
		return err
	}
	return nil
}

func (s *stateDriver) Get(key []byte) ([]byte, error) {
	kv, err := s.KV.Get(s.ctx, &kv.KV{
		Key: string(key),
	})
	if err != nil {
		return nil, err
	}
	return kv.Value, nil
}

func (s *stateDriver) List(filter func([]byte) bool) ([]byte, error) {
	return nil, nil
}

func (s *stateDriver) Query(labels map[string]string) ([]byte, error) {
	return nil, nil
}

func (s *stateDriver) Name() string {
	return StateDriverName
}
