package mqtt

import (
	"crypto/tls"
	"net"

	"github.com/256dpi/gomqtt/transport"
)

// IsTwoWayTLS check two-way tls connection
func IsTwoWayTLS(conn transport.Conn) bool {
	var inner net.Conn
	if tcps, ok := conn.(*transport.NetConn); ok {
		inner = tcps.UnderlyingConn()
	} else if wss, ok := conn.(*transport.WebSocketConn); ok {
		inner = wss.UnderlyingConn().UnderlyingConn()
	}
	tlsconn, ok := inner.(*tls.Conn)
	if !ok {
		return false
	}
	state := tlsconn.ConnectionState()
	if !state.HandshakeComplete {
		return false
	}
	return len(state.PeerCertificates) > 0
}
