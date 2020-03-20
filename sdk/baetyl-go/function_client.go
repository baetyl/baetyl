package baetyl

import (
	"time"

	"github.com/baetyl/baetyl/utils"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var callopt = grpc.FailFast(false)

// FunctionClientConfig function client config
type FunctionClientConfig struct {
	Address string `yaml:"address" json:"address"`
	Message struct {
		Length utils.Length `yaml:"length" json:"length" default:"{\"max\":4194304}"`
	} `yaml:"message" json:"message"`
	Backoff struct {
		Max time.Duration `yaml:"max" json:"max" default:"1m"`
	} `yaml:"backoff" json:"backoff"`
	Timeout time.Duration `yaml:"timeout" json:"timeout" default:"30s"`
}

// FClient client of functions server
type FClient struct {
	cli  FunctionClient
	cfg  FunctionClientConfig
	conn *grpc.ClientConn
}

// NewFClient creates a new client of functions server
func NewFClient(cc FunctionClientConfig) (*FClient, error) {
	ctx, cel := context.WithTimeout(context.Background(), cc.Timeout)
	defer cel()
	conn, err := grpc.DialContext(
		ctx,
		cc.Address,
		grpc.WithBlock(),
		grpc.WithInsecure(),
		grpc.WithTimeout(cc.Timeout),
		grpc.WithBackoffMaxDelay(cc.Backoff.Max),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(cc.Message.Length.Max))),
	)
	if err != nil {
		return nil, err
	}
	return &FClient{
		cfg:  cc,
		conn: conn,
		cli:  NewFunctionClient(conn),
	}, nil
}

// Call sends request to functions server
func (c *FClient) Call(msg *FunctionMessage) (*FunctionMessage, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), c.cfg.Timeout)
	defer cancel()
	return c.cli.Call(ctx, msg, callopt)
}

// Close closes the client
func (c *FClient) Close() error {
	return c.conn.Close()
}
