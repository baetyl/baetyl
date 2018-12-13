package session

import (
	"github.com/256dpi/gomqtt/transport"
	"github.com/baidu/openedge/hub/auth"
	"github.com/baidu/openedge/hub/common"
	"github.com/baidu/openedge/hub/config"
	"github.com/baidu/openedge/hub/persist"
	"github.com/baidu/openedge/hub/rule"
	"github.com/baidu/openedge/logger"
	"github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
)

// Manager session manager
type Manager struct {
	auth     *auth.Auth
	recorder *recorder
	sessions cmap.ConcurrentMap
	flow     common.Flow
	conf     *config.Message
	rules    *rule.Manager
	log      *logrus.Entry
}

// NewManager creates a session manager
func NewManager(conf *config.Config, flow common.Flow, rules *rule.Manager, pf *persist.Factory) (*Manager, error) {
	sessionDB, err := pf.NewDB("session.db")
	if err != nil {
		return nil, err
	}
	return &Manager{
		auth:     auth.NewAuth(conf.Principals),
		rules:    rules,
		flow:     flow,
		conf:     &conf.Message,
		recorder: newRecorder(sessionDB),
		sessions: cmap.New(),
		log:      logger.WithFields(common.LogComponent, "session_manager"),
	}, nil
}

// Handle handles connection
func (m *Manager) Handle(conn transport.Conn) {
	defer conn.Close()
	conn.SetReadLimit(int64(m.conf.Length.Max))
	newSession(conn, m).Handle()
}

// Close closes all sessions, called by rule manager
func (m *Manager) Close() {
	m.log.Info("session manager closing")
	for item := range m.sessions.IterBuffered() {
		item.Val.(*session).close(true)
	}
	m.log.Info("session manager closed")
}

// Called by session during onConnect
func (m *Manager) register(sess *session) error {
	if old, ok := m.sessions.Get(sess.id); ok {
		old.(*session).close(true)
	}
	m.sessions.Set(sess.id, sess)
	return m.rules.AddRuleSess(sess.id, !sess.clean, sess.publish, sess.republish)
}

// Called by session when error raises
func (m *Manager) remove(id string) {
	m.sessions.Remove(id)
	err := m.rules.RemoveRule(id)
	if err != nil {
		m.log.WithError(err).Debug("failed to remove rule")
	}
}
