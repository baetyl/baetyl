package rule

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/baidu/openedge/hub/common"
	"github.com/baidu/openedge/hub/config"
	"github.com/baidu/openedge/hub/router"
	"github.com/baidu/openedge/hub/utils"
	"github.com/baidu/openedge/logger"
	"github.com/orcaman/concurrent-map"
	"github.com/sirupsen/logrus"
)

const (
	initial = int32(0)
	started = int32(1)
	closed  = int32(2)
)

var errRuleManagerClosed = fmt.Errorf("rule manager already closed")

// Manager manages all rules of message routing
type Manager struct {
	status int32
	broker broker
	trieq0 *router.Trie
	rules  cmap.ConcurrentMap
	tomb   utils.Tomb
	log    *logrus.Entry
}

// NewManager creates a new rule manager
func NewManager(c []config.Subscription, b broker) (*Manager, error) {
	m := &Manager{
		broker: b,
		rules:  cmap.New(),
		trieq0: router.NewTrie(),
		log:    logger.WithFields(common.LogComponent, "rule_manager"),
	}
	m.rules.Set(common.RuleMsgQ0, newRuleQos0(m.broker, m.trieq0))
	m.rules.Set(common.RuleTopic, newRuleTopic(m.broker, m.trieq0))
	for _, sub := range c {
		err := m.AddSinkSub(common.RuleTopic, sub.Target.Topic, uint32(sub.Source.QOS), sub.Source.Topic, uint32(sub.Target.QOS), sub.Target.Topic)
		if err != nil {
			return nil, err
		}
	}
	if b.Config().Status.Logging.Enable {
		return m, m.tomb.Gos(m.logging)
	}
	return m, nil
}

// Start starts all rules
func (m *Manager) Start() {
	if !atomic.CompareAndSwapInt32(&m.status, initial, started) {
		return
	}
	for item := range m.rules.IterBuffered() {
		r := item.Val.(base)
		// r.log.Info("To start rule")
		if err := r.start(); err != nil {
			m.log.WithError(err).Infof("failed to start rule (%s)", r.uid())
		}
	}
}

// Close closes this manager
func (m *Manager) Close() {
	if !atomic.CompareAndSwapInt32(&m.status, started, closed) {
		return
	}
	m.log.Info("rule manager closing")
	defer m.log.Infof("rule manager closed, remaining offsets: %d", m.broker.OffsetChanLen())
	m.tomb.Kill()
	defer m.tomb.Wait()
	for item := range m.rules.IterBuffered() {
		r := item.Val.(base)
		// r.log.Info("To stop rule")
		r.stop()
	}
	// Wait all sinked messages are handled
	// TODO: how to handle the case of session closed by client during waiting
	for item := range m.rules.IterBuffered() {
		r := item.Val.(base)
		r.wait(false)
	}
	// wait all offsets persisted
	m.broker.WaitOffsetPersisted()
}

// AddRuleSess adds a new rule for session during running
func (m *Manager) AddRuleSess(id string, persistent bool, publish, republish common.Publish) error {
	if atomic.LoadInt32(&m.status) == closed {
		return errRuleManagerClosed
	}
	if _, ok := m.rules.Get(id); ok {
		return fmt.Errorf("rule (%s) exists", id)
	}
	m.rules.Set(id, newRuleSess(id, persistent, m.broker, m.trieq0, publish, republish))
	return nil
}

// StartRule starts a rule
func (m *Manager) StartRule(id string) error {
	if atomic.LoadInt32(&m.status) == closed {
		return errRuleManagerClosed
	} else if atomic.LoadInt32(&m.status) == initial {
		return nil
	}
	item, ok := m.rules.Get(id)
	if !ok {
		return fmt.Errorf("rule (%s) not found", id)
	}
	r := item.(base)
	// r.log.Info("To start rule")
	return r.start()
}

// RemoveRule removes a sink for session
func (m *Manager) RemoveRule(id string) error {
	if atomic.LoadInt32(&m.status) == closed {
		return errRuleManagerClosed
	}
	if item, ok := m.rules.Get(id); ok {
		m.rules.Remove(id)
		r := item.(base)
		// r.log.Info("To stop rule")
		r.stop()
		r.wait(true)
	}
	return nil
}

// AddSinkSub adds a sink subscription
func (m *Manager) AddSinkSub(ruleid, subid string, subqos uint32, subtopic string, pubqos uint32, pubtopic string) error {
	if atomic.LoadInt32(&m.status) == closed {
		return errRuleManagerClosed
	}
	item, ok := m.rules.Get(ruleid)
	if !ok {
		return fmt.Errorf("rule (%s) not found", ruleid)
	}
	r := item.(base)
	r.register(newSinkSub(subid, subqos, subtopic, pubqos, pubtopic, r.channel()))
	return nil
}

// RemoveSinkSub removes a sink subscription
func (m *Manager) RemoveSinkSub(id, topic string) error {
	if atomic.LoadInt32(&m.status) == closed {
		return errRuleManagerClosed
	}
	item, ok := m.rules.Get(id)
	if !ok {
		return fmt.Errorf("rule (%s) not found", id)
	}
	item.(base).remove(id, topic)
	return nil
}

func (m *Manager) logging() error {
	defer m.log.Debug("status logging task stopped")
	t := time.NewTicker(m.broker.Config().Status.Logging.Interval)
	defer t.Stop()
	for {
		select {
		case <-m.tomb.Dying():
			return nil
		case <-t.C:
			m.log.Infof("rule Status")
			m.log.Infof("  rule count: %d", m.rules.Count())
			for item := range m.rules.IterBuffered() {
				r := item.Val.(base)
				offsetPersisted := "<nil>"
				if v, _ := m.broker.OffsetPersisted(r.uid()); v != nil {
					offsetPersisted = strconv.FormatUint(*v, 10)
				}
				m.log.Infof("  persisted offset:%s, %s", offsetPersisted, r.info())
			}
		}
	}
}
