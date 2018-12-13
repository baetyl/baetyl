package server

import (
	"github.com/256dpi/gomqtt/transport"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/trans"
	"github.com/baidu/openedge/trans/mqtt"
	"github.com/baidu/openedge/utils"
	"github.com/sirupsen/logrus"
)

// Handle handles connection
type Handle func(transport.Conn)

// Manager manager of servers
type Manager struct {
	servers []transport.Server
	handle  Handle
	tomb    utils.Tomb
	log     *logrus.Entry
}

// NewManager creates a server manager
func NewManager(addrs []string, cert trans.Certificate, handle Handle) (*Manager, error) {
	launcher, err := mqtt.NewLauncher(cert)
	if err != nil {
		return nil, err
	}
	m := &Manager{
		servers: make([]transport.Server, 0),
		handle:  handle,
		log:     logger.WithFields("server", "manager"),
	}
	for _, addr := range addrs {
		svr, err := launcher.Launch(addr)
		if err != nil {
			m.Close()
			return nil, err
		}
		m.servers = append(m.servers, svr)
	}
	return m, nil
}

// Start starts all servers
func (m *Manager) Start() {
	for _, item := range m.servers {
		svr := item
		m.tomb.Go(func() error {
			for {
				conn, err := svr.Accept()
				if err != nil {
					if !m.tomb.Alive() {
						return nil
					}
					m.log.WithError(err).Errorf("failed to accept connection")
					continue
				}
				go m.handle(conn)
			}
		})
	}
}

// Close closes server manager
func (m *Manager) Close() error {
	m.tomb.Kill(nil)
	for _, svr := range m.servers {
		svr.Close()
	}
	return m.tomb.Wait()
}
