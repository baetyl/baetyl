package server

import (
	"fmt"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
	"github.com/baetyl/baetyl/protocol/mqtt"
	"github.com/baetyl/baetyl/utils"
	"github.com/stretchr/testify/assert"
)

func TestMqttTcp(t *testing.T) {
	addr := []string{"tcp://:0", "tcp://127.0.0.1:0"}
	m, err := NewManager(addr, utils.Certificate{}, func(conn transport.Conn) {
		p, err := conn.Receive()
		assert.NoError(t, err)
		err = conn.Send(p, false)
		assert.NoError(t, err)
	})
	assert.NoError(t, err)
	defer m.Close()
	m.Start()
	time.Sleep(time.Millisecond * 100)

	dailer, err := mqtt.NewDialer(utils.Certificate{}, 0)
	request := packet.NewConnect()
	request.ClientID = m.servers[0].Addr().String()
	conn, err := dailer.Dial(getURL(m.servers[0], "tcp"))
	assert.NoError(t, err)
	err = conn.Send(request, false)
	assert.NoError(t, err)
	response, err := conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, request.String(), response.String())
	conn.Close()

	request.ClientID = m.servers[1].Addr().String()
	conn, err = dailer.Dial(getURL(m.servers[1], "tcp"))
	assert.NoError(t, err)
	err = conn.Send(request, true)
	assert.NoError(t, err)
	response, err = conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, request.String(), response.String())
	conn.Close()
}

func TestMqttTcpTls(t *testing.T) {
	addr := []string{"ssl://localhost:0"}
	cert := utils.Certificate{
		CA:   "./testcert/ca.pem",
		Key:  "./testcert/server.key",
		Cert: "./testcert/server.pem",
	}
	count := int32(0)
	m, err := NewManager(addr, cert, func(conn transport.Conn) {
		c := atomic.AddInt32(&count, 1)
		p, err := conn.Receive()

		assert.NoError(t, err)
		assert.NotNil(t, p)

		ok := mqtt.IsTwoWayTLS(conn)
		if c == 3 {
			assert.Truef(t, ok, "count: %d", c)
		} else {
			assert.Falsef(t, ok, "count: %d", c)
		}

		err = conn.Send(p, false)
		assert.NoError(t, err)
	})
	assert.NoError(t, err)
	defer m.Close()
	m.Start()
	time.Sleep(time.Millisecond * 100)

	url := getURL(m.servers[0], "ssl")
	request := packet.NewConnect()
	request.ClientID = m.servers[0].Addr().String()

	// count: 1
	dailer, err := mqtt.NewDialer(utils.Certificate{Insecure: true}, 0)
	assert.NoError(t, err)
	conn, err := dailer.Dial(url)
	assert.NoError(t, err)
	err = conn.Send(request, false)
	assert.NoError(t, err)
	response, err := conn.Receive()
	assert.NoError(t, err)
	conn.Close()

	// count: 2
	dailer, err = mqtt.NewDialer(utils.Certificate{
		CA:       "./testcert/ca.pem",
		Insecure: true,
	}, 0)
	assert.NoError(t, err)
	conn, err = dailer.Dial(url)
	assert.NoError(t, err)
	err = conn.Send(request, false)
	assert.NoError(t, err)
	response, err = conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, request.String(), response.String())
	conn.Close()

	// count: 3
	dailer, err = mqtt.NewDialer(utils.Certificate{
		CA:       "./testcert/ca.pem",
		Key:      "./testcert/testssl2.key",
		Cert:     "./testcert/testssl2.pem",
		Insecure: true,
	}, 0)
	assert.NoError(t, err)
	conn, err = dailer.Dial(url)
	assert.NoError(t, err)
	err = conn.Send(request, false)
	assert.NoError(t, err)
	response, err = conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, request.String(), response.String())
	conn.Close()
}

func TestMqttWebSocket(t *testing.T) {
	addr := []string{"ws://localhost:0", "ws://127.0.0.1:0/mqtt"}
	cert := utils.Certificate{}
	m, err := NewManager(addr, cert, func(conn transport.Conn) {
		p, err := conn.Receive()
		assert.NoError(t, err)
		err = conn.Send(p, false)
		assert.NoError(t, err)
	})
	assert.NoError(t, err)
	defer m.Close()
	m.Start()
	time.Sleep(time.Millisecond * 100)

	dailer, err := mqtt.NewDialer(utils.Certificate{Insecure: true}, 0)
	assert.NoError(t, err)
	request := packet.NewConnect()
	request.ClientID = m.servers[0].Addr().String()
	conn, err := dailer.Dial(getURL(m.servers[0], "ws"))
	assert.NoError(t, err)
	err = conn.Send(request, false)
	assert.NoError(t, err)
	response, err := conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, request.String(), response.String())
	conn.Close()

	request.ClientID = m.servers[1].Addr().String()
	conn, err = dailer.Dial(getURL(m.servers[1], "ws") + "/mqtt")
	assert.NoError(t, err)
	err = conn.Send(request, true)
	assert.NoError(t, err)
	response, err = conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, request.String(), response.String())
	conn.Close()

	request.ClientID = m.servers[1].Addr().String() + "-1"
	conn, err = dailer.Dial(getURL(m.servers[1], "ws") + "/notexist")
	assert.NoError(t, err)
	err = conn.Send(request, false)
	assert.NoError(t, err)
	response, err = conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, request.String(), response.String())
	conn.Close()
}

func TestMqttWebSocketTls(t *testing.T) {
	addr := []string{"wss://localhost:0/mqtt"}
	cert := utils.Certificate{
		CA:   "./testcert/ca.pem",
		Key:  "./testcert/server.key",
		Cert: "./testcert/server.pem",
	}
	count := int32(0)
	m, err := NewManager(addr, cert, func(conn transport.Conn) {
		c := atomic.AddInt32(&count, 1)
		fmt.Println(count, conn.LocalAddr())
		p, err := conn.Receive()
		assert.NoError(t, err)
		assert.NotNil(t, p)

		ok := mqtt.IsTwoWayTLS(conn)
		if c == 3 {
			assert.Truef(t, ok, "count: %d", c)
		} else {
			assert.Falsef(t, ok, "count: %d", c)
		}

		err = conn.Send(p, false)
		assert.NoError(t, err)
	})
	assert.NoError(t, err)
	defer m.Close()
	m.Start()
	time.Sleep(time.Millisecond * 100)

	url := getURL(m.servers[0], "wss") + "/mqtt"
	request := packet.NewConnect()
	request.ClientID = m.servers[0].Addr().String()

	dailer, err := mqtt.NewDialer(utils.Certificate{}, 0)
	assert.NoError(t, err)
	conn, err := dailer.Dial(url)
	assert.Nil(t, conn)
	switch err.Error() {
	case "x509: certificate signed by unknown authority":
	case "x509: cannot validate certificate for 127.0.0.1 because it doesn't contain any IP SANs":
	default:
		assert.FailNow(t, "error expected")
	}

	// count: 1
	dailer, err = mqtt.NewDialer(utils.Certificate{Insecure: true}, 0)
	assert.NoError(t, err)
	conn, err = dailer.Dial(url)
	assert.NoError(t, err)
	err = conn.Send(request, false)
	assert.NoError(t, err)
	response, err := conn.Receive()
	assert.NoError(t, err)
	conn.Close()

	// count: 2
	dailer, err = mqtt.NewDialer(utils.Certificate{
		CA:       "./testcert/ca.pem",
		Insecure: true,
	}, 0)
	assert.NoError(t, err)
	conn, err = dailer.Dial(url)
	assert.NoError(t, err)
	err = conn.Send(request, false)
	assert.NoError(t, err)
	response, err = conn.Receive()
	assert.NoError(t, err)
	conn.Close()

	// count: 3
	dailer, err = mqtt.NewDialer(utils.Certificate{
		CA:       "./testcert/ca.pem",
		Key:      "./testcert/testssl2.key",
		Cert:     "./testcert/testssl2.pem",
		Insecure: true,
	}, 0)
	assert.NoError(t, err)
	conn, err = dailer.Dial(url)
	assert.NoError(t, err)
	err = conn.Send(request, false)
	assert.NoError(t, err)
	response, err = conn.Receive()
	assert.NoError(t, err)
	assert.Equal(t, request.String(), response.String())
	conn.Close()
}

func TestServerException(t *testing.T) {
	addr := []string{"tcp://:28767", "tcp://:28767"}
	_, err := NewManager(addr, utils.Certificate{}, nil)
	switch err.Error() {
	case "listen tcp :28767: bind: address already in use":
	case "listen tcp :28767: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted.":
	default:
		assert.FailNow(t, "error expected")
	}
	addr = []string{"tcp://:28767", "ssl://:28767"}
	_, err = NewManager(addr, utils.Certificate{}, nil)
	assert.EqualError(t, err, "tls: neither Certificates nor GetCertificate set in Config")
	addr = []string{"ws://:28767/v1", "wss://:28767/v2"}
	_, err = NewManager(addr, utils.Certificate{}, nil)
	assert.EqualError(t, err, "tls: neither Certificates nor GetCertificate set in Config")
	addr = []string{"ws://:28767/v1", "ws://:28767/v1"}
	_, err = NewManager(addr, utils.Certificate{}, nil)
	switch err.Error() {
	case "listen tcp :28767: bind: address already in use":
	case "listen tcp :28767: bind: Only one usage of each socket address (protocol/network address/port) is normally permitted.":
	default:
		assert.FailNow(t, "error expected")
	}

	// TODO: test more special case
	// addr = []string{"ws://:28767/v1", "ws://0.0.0.0:28767/v2"}
	// addr = []string{"ws://localhost:28767/v1", "ws://127.0.0.1:28767/v2"}
}

func getPort(s transport.Server) string {
	_, port, _ := net.SplitHostPort(s.Addr().String())
	return port
}

func getURL(s transport.Server, protocol string) string {
	return fmt.Sprintf("%s://%s", protocol, s.Addr().String())
}
