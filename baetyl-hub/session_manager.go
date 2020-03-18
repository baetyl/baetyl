package hub

import (
	"github.com/256dpi/gomqtt/transport"
	"github.com/baetyl/baetyl/baetyl-hub/common"
	"github.com/baetyl/baetyl/baetyl-hub/config"
	"github.com/baetyl/baetyl/logger"
	cmap "github.com/orcaman/concurrent-map"
)

// sessionManager session manager
type sessionManager struct {
	auth     *authenticator
	recorder *recorder
	sessions cmap.ConcurrentMap
	flow     common.Flow
	conf     *config.Message
	rules    *ruleManager
	log      logger.Logger
}

func (h *hub) startSession() error {
	conf := &h.cfg
	flow := h.broker.Flow
	rules := h.rules
	pf := h.storage
	sessionDB, err := pf.newDB("session.db")
	if err != nil {
		return err
	}
	h.sess = &sessionManager{
		auth:     newAuthenticator(conf.Principals),
		rules:    rules,
		flow:     flow,
		conf:     &conf.Message,
		recorder: newRecorder(sessionDB),
		sessions: cmap.New(),
		log:      logger.WithField("manager", "session"),
	}
	return nil
}

func (m *sessionManager) handle(conn transport.Conn) {
	defer conn.Close()
	conn.SetReadLimit(int64(m.conf.Length.Max))
	newSession(conn, m).handle()
}

func (h *hub) stopSession() {
	h.sess.log.Infof("session manager closing")
	for item := range h.sess.sessions.IterBuffered() {
		item.Val.(*session).close(true)
	}
	h.sess.log.Infof("session manager closed")
}

// Called by session during onConnect
func (m *sessionManager) register(sess *session) error {
	if old, ok := m.sessions.Get(sess.id); ok {
		old.(*session).close(true)
	}
	m.sessions.Set(sess.id, sess)
	return m.rules.AddRuleSess(sess.id, !sess.clean, sess.publish, sess.republish)
}

// Called by session when error raises
func (m *sessionManager) remove(id string) {
	m.sessions.Remove(id)
	err := m.rules.RemoveRule(id)
	if err != nil {
		m.log.WithError(err).Debugf("failed to remove rule")
	}
}
