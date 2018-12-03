package session

import (
	"encoding/json"
	"sync"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/module/hub/common"
	"github.com/baidu/openedge/module/hub/persist"
	"github.com/juju/errors"
	"github.com/sirupsen/logrus"
)

// recorder records session info
type recorder struct {
	db  persist.Database
	log *logrus.Entry
	sync.Mutex
}

// NewRecorder creates a recorder
func newRecorder(db persist.Database) *recorder {
	return &recorder{
		db:  db,
		log: logger.WithFields(common.LogComponent, "recorder"),
	}
}

// AddSub adds subscription to id
func (c *recorder) addSub(id string, sub packet.Subscription) error {
	c.Lock()
	defer c.Unlock()

	subs, err := c.getsubs(id)
	if err != nil {
		return errors.Trace(err)
	}
	subs[sub.Topic] = sub
	creates, err := json.Marshal(subs)
	if err != nil {
		return errors.Annotate(err, "Marshal subscriptions failed")
	}
	err = c.db.BucketPut(common.BucketNameDotSubscription, []byte(id), creates)
	if err != nil {
		return errors.Annotate(err, "Persist subscriptions failed")
	}
	c.log.Debugf("Persist subscription successfully: qos=%d, topic=%s, id=%s", sub.QOS, sub.Topic, id)
	return nil
}

// RemoveSub removes subscription from id
func (c *recorder) removeSub(id, topic string) error {
	c.Lock()
	defer c.Unlock()

	subs, err := c.getsubs(id)
	if err != nil {
		return errors.Trace(err)
	}
	delete(subs, topic)
	creates, err := json.Marshal(subs)
	if err != nil {
		return errors.Annotate(err, "Marshal subscriptions failed")
	}
	err = c.db.BucketPut(common.BucketNameDotSubscription, []byte(id), creates)
	if err != nil {
		return errors.Annotate(err, "Persist subscriptions failed")
	}
	c.log.Debugf("Remove subscription successfully: topic=%s, id=%s", topic, id)
	return nil
}

// GetSubs gets subscriptions of id
func (c *recorder) getSubs(id string) ([]packet.Subscription, error) {
	c.Lock()
	defer c.Unlock()

	subs, err := c.getsubs(id)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res := make([]packet.Subscription, 0)
	for _, sub := range subs {
		res = append(res, sub)
	}
	c.log.Debugf("Get %d subscription(s) successfully: id=%s", len(res), id)
	return res, nil
}

func (c *recorder) getsubs(id string) (map[string]packet.Subscription, error) {
	olds, err := c.db.BucketGet(common.BucketNameDotSubscription, []byte(id))
	if err != nil {
		return nil, errors.Annotate(err, "Get subscriptions failed")
	}
	subs := make(map[string]packet.Subscription)
	if olds != nil {
		err = json.Unmarshal(olds, &subs)
		if err != nil {
			return nil, errors.Annotate(err, "Unmarshal subscriptions failed")
		}
	}
	return subs, nil
}

// SetRetained sets retained message of topic
func (c *recorder) setRetained(topic string, msg *packet.Message) error {
	c.Lock()
	defer c.Unlock()
	if !msg.Retain {
		return nil
	}
	value, err := json.Marshal(msg)
	if err != nil {
		return errors.Trace(err)
	}
	err = c.db.BucketPut(common.BucketNameDotRetained, []byte(topic), value)
	if err != nil {
		c.log.WithError(err).Errorf("Retain message failed: topic=%s", topic)
		return errors.Annotate(err, "Persist retained message failed")
	}
	c.log.Debugf("Retain message successfully: topic=%s", topic)
	return nil
}

// GetRetained gets retained message of all topics
func (c *recorder) getRetained() ([]*packet.Message, error) {
	l, err := c.db.BucketList(common.BucketNameDotRetained)
	if err != nil {
		c.log.WithError(err).Errorf("Get retaind message failed")
		return nil, errors.Annotatef(err, "Get retained message failed")
	}
	result := make([]*packet.Message, 0)
	for _, v := range l {
		var msg packet.Message
		err := json.Unmarshal(v, &msg)
		if err != nil {
			c.log.WithError(err).Warn("Unmarshal retaind message failed")
		}
		result = append(result, &msg)
	}
	c.log.Debugf("Get %d retaind message(s) successfully", len(result))
	return result, nil
}

// RemoveRetained removes retained message of topic
func (c *recorder) removeRetained(topic string) error {
	err := c.db.BucketDelete(common.BucketNameDotRetained, []byte(topic))
	if err != nil {
		c.log.WithError(err).Errorf("Remove retaind message failed: topic=%s", topic)
		return errors.Annotate(err, "Remove retained message failed")
	}
	c.log.Debugf("Remove retaind message successfully: topic=%s", topic)
	return nil
}

// SetWill sets will message of cleint
func (c *recorder) setWill(id string, msg *packet.Message) error {
	value, err := json.Marshal(msg)
	if err != nil {
		return errors.Annotate(err, "Marshal will messages failed")
	}
	err = c.db.BucketPut(common.BucketNameDotWill, []byte(id), value)
	if err != nil {
		c.log.WithError(err).Errorf("Persist will message failed: topic=%s, id=%s", msg.Topic, id)
		return errors.Annotate(err, "Persist will messages failed")
	}
	c.log.Debugf("Persist will message successfully: topic=%s, id=%s", msg.Topic, id)
	return nil
}

// GetWill gets will message of cleint
func (c *recorder) getWill(id string) (*packet.Message, error) {
	data, err := c.db.BucketGet(common.BucketNameDotWill, []byte(id))
	if err != nil {
		c.log.WithError(err).Errorf("Get will message failed: id=%s", id)
		return nil, errors.Annotate(err, "Get will messages failed")
	}
	if len(data) == 0 {
		return nil, nil
	}
	var msg packet.Message
	err = json.Unmarshal(data, &msg)
	if err != nil {
		return nil, errors.Annotate(err, "Unmarshal will messages failed")
	}
	c.log.Debugf("Get will message successfully: topic=%s, id=%s", msg.Topic, id)
	return &msg, nil
}

// RemoveWill removes will message of cleint
func (c *recorder) removeWill(id string) error {
	err := c.db.BucketDelete(common.BucketNameDotWill, []byte(id))
	if err != nil {
		c.log.WithError(err).Errorf("Remove will message failed: id=%s", id)
		return errors.Annotate(err, "Remove will messages failed")
	}
	c.log.Debugf("Remove will message successfully: id=%s", id)
	return nil
}
