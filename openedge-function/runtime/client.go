package runtime

import (
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

var callopt = grpc.FailFast(false)

// Client client of function server
type Client struct {
	cli  RuntimeClient
	conf ClientInfo
	conn *grpc.ClientConn
}

// NewClient creates a new client
func NewClient(cc ClientInfo) (*Client, error) {
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
	return &Client{
		conf: cc,
		conn: conn,
		cli:  NewRuntimeClient(conn),
	}, nil
}

// Handle sends request to function server
func (c *Client) Handle(in *Message) (*Message, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), c.conf.Timeout)
	defer cancel()
	return c.cli.Handle(ctx, in, callopt)
}

// Close closes the client
func (c *Client) Close() error {
	return c.conn.Close()
}
