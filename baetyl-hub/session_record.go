package hub

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/baetyl/baetyl/baetyl-hub/common"
	"github.com/baetyl/baetyl/logger"

	"github.com/256dpi/gomqtt/packet"
)

// recorder records session info
type recorder struct {
	db  Database
	log logger.Logger
	sync.Mutex
}

// NewRecorder creates a recorder
func newRecorder(db Database) *recorder {
	return &recorder{
		db:  db,
		log: logger.WithField("session", "recorder"),
	}
}

// AddSub adds subscription to id
func (c *recorder) addSub(id string, sub packet.Subscription) error {
	c.Lock()
	defer c.Unlock()

	subs, err := c.getsubs(id)
	if err != nil {
		return err
	}
	subs[sub.Topic] = sub
	creates, err := json.Marshal(subs)
	if err != nil {
		return fmt.Errorf("failed to marshal subscriptions: %s", err.Error())
	}
	err = c.db.BucketPut(common.BucketNameDotSubscription, []byte(id), creates)
	if err != nil {
		return fmt.Errorf("failed to persist subscriptions: %s", err.Error())
	}
	c.log.Debugf("subscription persisted: qos=%d, topic=%s, id=%s", sub.QOS, sub.Topic, id)
	return nil
}

// RemoveSub removes subscription from id
func (c *recorder) removeSub(id, topic string) error {
	c.Lock()
	defer c.Unlock()

	subs, err := c.getsubs(id)
	if err != nil {
		return err
	}
	delete(subs, topic)
	creates, err := json.Marshal(subs)
	if err != nil {
		return fmt.Errorf("failed to marshal subscriptions: %s", err.Error())
	}
	err = c.db.BucketPut(common.BucketNameDotSubscription, []byte(id), creates)
	if err != nil {
		return fmt.Errorf("failed to persist subscriptions: %s", err.Error())
	}
	c.log.Debugf("subscription removed: topic=%s, id=%s", topic, id)
	return nil
}

// GetSubs gets subscriptions of id
func (c *recorder) getSubs(id string) ([]packet.Subscription, error) {
	c.Lock()
	defer c.Unlock()

	subs, err := c.getsubs(id)
	if err != nil {
		return nil, err
	}
	res := make([]packet.Subscription, 0)
	for _, sub := range subs {
		res = append(res, sub)
	}
	c.log.Debugf("%d subscription(s) got: id=%s", len(res), id)
	return res, nil
}

func (c *recorder) getsubs(id string) (map[string]packet.Subscription, error) {
	olds, err := c.db.BucketGet(common.BucketNameDotSubscription, []byte(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get subscriptions: %s", err.Error())
	}
	subs := make(map[string]packet.Subscription)
	if olds != nil {
		err = json.Unmarshal(olds, &subs)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal subscriptions: %s", err.Error())
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
		return err
	}
	err = c.db.BucketPut(common.BucketNameDotRetained, []byte(topic), value)
	if err != nil {
		c.log.WithError(err).Errorf("failed to persist retain message: topic=%s", topic)
		return fmt.Errorf("failed to persist retain message: %s", err.Error())
	}
	c.log.Debugf("retain message persisited: topic=%s", topic)
	return nil
}

// GetRetained gets retained message of all topics
func (c *recorder) getRetained() ([]*packet.Message, error) {
	l, err := c.db.BucketList(common.BucketNameDotRetained)
	if err != nil {
		c.log.WithError(err).Errorf("failed to get retaind message")
		return nil, fmt.Errorf("failed to get retain message: %s", err.Error())
	}
	result := make([]*packet.Message, 0)
	for _, v := range l {
		var msg packet.Message
		err := json.Unmarshal(v, &msg)
		if err != nil {
			c.log.WithError(err).Warnf("failed to unmarshal retain message")
		}
		result = append(result, &msg)
	}
	c.log.Debugf("%d retain message(s) got", len(result))
	return result, nil
}

// RemoveRetained removes retain message of topic
func (c *recorder) removeRetained(topic string) error {
	err := c.db.BucketDelete(common.BucketNameDotRetained, []byte(topic))
	if err != nil {
		c.log.WithError(err).Errorf("failed to remove retain message: topic=%s", topic)
		return fmt.Errorf("failed to remove retain message: %s", err.Error())
	}
	c.log.Debugf("retain message removed: topic=%s", topic)
	return nil
}

// SetWill sets will message of cleint
func (c *recorder) setWill(id string, msg *packet.Message) error {
	value, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal will message: %s", err.Error())
	}
	err = c.db.BucketPut(common.BucketNameDotWill, []byte(id), value)
	if err != nil {
		c.log.WithError(err).Errorf("failed to persist will message: topic=%s, id=%s", msg.Topic, id)
		return fmt.Errorf("failed to persist will message: %s", err.Error())
	}
	c.log.Debugf("will message persisted: topic=%s, id=%s", msg.Topic, id)
	return nil
}

// GetWill gets will message of cleint
func (c *recorder) getWill(id string) (*packet.Message, error) {
	data, err := c.db.BucketGet(common.BucketNameDotWill, []byte(id))
	if err != nil {
		c.log.WithError(err).Errorf("failed to get will message: id=%s", id)
		return nil, fmt.Errorf("failed to get will message: %s", err.Error())
	}
	if len(data) == 0 {
		return nil, nil
	}
	var msg packet.Message
	err = json.Unmarshal(data, &msg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal will message: %s", err.Error())
	}
	c.log.Debugf("will message got: topic=%s, id=%s", msg.Topic, id)
	return &msg, nil
}

// RemoveWill removes will message of cleint
func (c *recorder) removeWill(id string) error {
	err := c.db.BucketDelete(common.BucketNameDotWill, []byte(id))
	if err != nil {
		c.log.WithError(err).Errorf("failed to remove will message: id=%s", id)
		return fmt.Errorf("failed to remove will message: %s", err.Error())
	}
	c.log.Debugf("will message removed: id=%s", id)
	return nil
}
