package hub

import (
	"path/filepath"

	"github.com/256dpi/gomqtt/transport"
	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/protocol/mqtt"
	"github.com/baetyl/baetyl/utils"
)

type server struct {
	ss   []transport.Server
	tomb utils.Tomb
	log  logger.Logger
}

func (h *hub) runServer() error {
	addrs := h.cfg.Listen
	cert := h.cfg.Certificate
	if !filepath.IsAbs(cert.CA) {
		cert.CA = filepath.Join(configDir, cert.CA)
	}
	if !filepath.IsAbs(cert.Cert) {
		cert.Cert = filepath.Join(configDir, cert.Cert)
	}
	if !filepath.IsAbs(cert.Key) {
		cert.Key = filepath.Join(configDir, cert.Key)
	}
	launcher, err := mqtt.NewLauncher(cert)
	if err != nil {
		return err
	}
	m := &server{
		ss:  make([]transport.Server, 0),
		log: logger.WithField("manager", "server"),
	}
	h.server = m
	for _, addr := range addrs {
		svr, err := launcher.Launch(addr)
		if err != nil {
			h.stopServer()
			return err
		}
		m.ss = append(m.ss, svr)
	}
	for _, item := range m.ss {
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
				go h.sess.handle(conn)
			}
		})
	}
	return nil
}

func (h *hub) stopServer() error {
	h.server.tomb.Kill(nil)
	for _, svr := range h.server.ss {
		svr.Close()
	}
	return h.server.tomb.Wait()
}
