package mqtt

import (
	"crypto/tls"
	"net"

	"github.com/256dpi/gomqtt/transport"
	"github.com/juju/errors"
)

// GetTLSConn gets tls connection if tls enabled
func GetTLSConn(conn transport.Conn) (*tls.Conn, bool) {
	var innerconn net.Conn
	if netconn, ok := conn.(*transport.NetConn); ok {
		innerconn = netconn.UnderlyingConn()
	} else if wsconn, ok := conn.(*transport.WebSocketConn); ok {
		innerconn = wsconn.UnderlyingConn().UnderlyingConn()
	}
	tlsconn, ok := innerconn.(*tls.Conn)
	return tlsconn, ok
}

// GetClientCertSerialNumber gets client certificate serial number if tls enabled
// TODO: test not pass, cannot get serial number from client connection
func GetClientCertSerialNumber(conn *tls.Conn) (string, error) {
	state := conn.ConnectionState()
	if !state.HandshakeComplete {
		return "", errors.Errorf("TLS handshake not completed")
	}
	length := len(state.PeerCertificates)
	if length == 0 {
		return "", errors.Errorf("Certidicate not found")
	}
	serialNumber := state.PeerCertificates[len(state.PeerCertificates)-1].SerialNumber
	return serialNumber.String(), nil
}
