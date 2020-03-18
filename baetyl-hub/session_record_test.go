package hub

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/stretchr/testify/assert"
)

func TestRecordSubscription(t *testing.T) {
	pf, err := NewFactory("./var/db/")
	assert.NoError(t, err)
	dbName := strings.ToLower(t.Name()) + ".db"
	os.Remove("./var/db/" + dbName)
	db, err := pf.NewDB(dbName)
	assert.Nil(t, err)
	defer db.Close()

	c := newRecorder(db)
	assert.Nil(t, err)
	id := dbName
	sub, err := c.getSubs(id)
	assert.Nil(t, err)
	assert.Len(t, sub, 0)
	err = c.addSub(id, packet.Subscription{Topic: "topic-1", QOS: 1})
	assert.Nil(t, err)
	sub, err = c.getSubs(id)
	assert.Nil(t, err)
	assert.Len(t, sub, 1)
	assert.Equal(t, "topic-1", sub[0].Topic)
	assert.Equal(t, packet.QOS(1), sub[0].QOS)
	err = c.addSub(id, packet.Subscription{Topic: "topic-1", QOS: 0})
	assert.Nil(t, err)
	sub, err = c.getSubs(id)
	assert.Nil(t, err)
	assert.Len(t, sub, 1)
	assert.Equal(t, "topic-1", sub[0].Topic)
	assert.Equal(t, packet.QOS(0), sub[0].QOS)
	err = c.addSub(id, packet.Subscription{Topic: "topic-2", QOS: 0})
	assert.Nil(t, err)
	sub, err = c.getSubs(id)
	assert.Nil(t, err)
	assert.Len(t, sub, 2)
	err = c.removeSub(id, "topic-1")
	assert.Nil(t, err)
	err = c.removeSub(id, "topic-2")
	assert.Nil(t, err)
	err = c.removeSub(id, "topic-3")
	assert.Nil(t, err)
	sub, err = c.getSubs(id)
	assert.Nil(t, err)
	assert.Len(t, sub, 0)
}

// func TestRecordRetainMessage(t *testing.T) {
// 	pf, err := NewFactory("./var/db/")
// 	assert.NoError(t, err)
// 	dbName := strings.ToLower(t.Name()) + ".db"
// 	os.Remove("./var/db/" + dbName)
// 	db, err := pf.NewDB(dbName)
// 	assert.Nil(t, err)
// 	defer db.Close()

// 	topic := dbName
// 	c := NewRecorder(db)
// 	msg, err := c.getRetained()
// 	assert.Nil(t, err)
// 	assert.Len(t, msg, 0)

// 	rmsg := common.Message{
// 		Message: packet.Message{Topic: topic, QOS: 1, Payload: []byte("topic-1")},
// 	}
// 	err = c.setRetained(topic, &rmsg)
// 	assert.Nil(t, err)
// 	msg, err = c.getRetained()
// 	assert.Nil(t, err)
// 	assert.Len(t, msg, 1)
// 	assert.Equal(t, topic, msg[0].Topic)
// 	assert.Equal(t, []byte("topic-1"), msg[0].Payload)

// 	rmsg = common.Message{
// 		Message: packet.Message{Topic: "/china/shanghai", QOS: 0, Payload: []byte("topic-11")},
// 	}
// 	err = c.setRetained(rmsg.Topic, &rmsg)
// 	assert.Nil(t, err)
// 	rmsg = common.Message{
// 		Message: packet.Message{Topic: "/china/shanghai/pudong", QOS: 0, Payload: []byte("topic-11")},
// 	}
// 	err = c.setRetained(rmsg.Topic, &rmsg)
// 	assert.Nil(t, err)
// 	rmsg = common.Message{
// 		Message: packet.Message{Topic: "/china/beijing/haidian", QOS: 0, Payload: []byte("topic-11")},
// 	}
// 	err = c.setRetained(rmsg.Topic, &rmsg)
// 	assert.Nil(t, err)
// 	subTopics := []packet.Subscription{packet.Subscription{Topic: "/china/shanghai", QOS: 1},
// 		packet.Subscription{Topic: "/china/#", QOS: 1}}
// 	subpkt := packet.Subscribe{Subscriptions: subTopics}
// 	retainedMsgs, err := c.MatchRetained(&subpkt)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 3, len(retainedMsgs))

// 	c.removeRetained("/china/shanghai")
// 	retainedMsgs, err = c.MatchRetained(&subpkt)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 2, len(retainedMsgs))

// 	rmsg = common.Message{
// 		Message: packet.Message{Topic: "/china/shanghai/minhang", QOS: 0, Payload: []byte("topic-11")},
// 	}
// 	err = c.setRetained(rmsg.Topic, &rmsg)
// 	assert.Nil(t, err)
// 	rmsg = common.Message{
// 		Message: packet.Message{Topic: "/china/shanghai/huangpu", QOS: 0, Payload: []byte("topic-11")},
// 	}
// 	err = c.setRetained(rmsg.Topic, &rmsg)
// 	assert.Nil(t, err)
// 	subTopics = []packet.Subscription{packet.Subscription{Topic: "/china/+/+", QOS: 1}}
// 	subpkt = packet.Subscribe{Subscriptions: subTopics}
// 	retainedMsgs, err = c.MatchRetained(&subpkt)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 4, len(retainedMsgs))

// 	subTopics = []packet.Subscription{packet.Subscription{Topic: "/china/shanghai/+", QOS: 1}}
// 	subpkt = packet.Subscribe{Subscriptions: subTopics}
// 	retainedMsgs, err = c.MatchRetained(&subpkt)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 3, len(retainedMsgs))

// 	subTopics = []packet.Subscription{packet.Subscription{Topic: "/china/#", QOS: 1},
// 		packet.Subscription{Topic: "/china/+/+", QOS: 1}}
// 	subpkt = packet.Subscribe{Subscriptions: subTopics}
// 	retainedMsgs, err = c.MatchRetained(&subpkt)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 4, len(retainedMsgs))
// }

// func TestRecordRetainMessageUtf8(t *testing.T) {
// 	pf, err := NewFactory("./var/db/")
// 	assert.NoError(t, err)
// 	dbName := strings.ToLower(t.Name()) + ".db"
// 	os.Remove("./var/db/" + dbName)
// 	db, err := pf.NewDB(dbName)
// 	assert.Nil(t, err)
// 	defer db.Close()

// 	topic := dbName
// 	c := NewRecorder(db)
// 	msg, err := c.getRetained()
// 	assert.Nil(t, err)
// 	assert.Len(t, msg, 0)

// 	rmsg := common.Message{
// 		Message: packet.Message{Topic: topic, QOS: 1, Payload: []byte("topic-1")},
// 	}
// 	err = c.setRetained(topic, &rmsg)
// 	assert.Nil(t, err)
// 	msg, err = c.getRetained()
// 	assert.Nil(t, err)
// 	assert.Len(t, msg, 1)
// 	assert.Equal(t, topic, msg[0].Topic)
// 	assert.Equal(t, []byte("topic-1"), msg[0].Payload)

// 	rmsg = common.Message{
// 		Message: packet.Message{Topic: "/中国/上海", QOS: 0, Payload: []byte("topic-11")},
// 	}
// 	err = c.setRetained(rmsg.Topic, &rmsg)
// 	assert.Nil(t, err)
// 	rmsg = common.Message{
// 		Message: packet.Message{Topic: "/中国/上海/浦东", QOS: 0, Payload: []byte("topic-11")},
// 	}
// 	err = c.setRetained(rmsg.Topic, &rmsg)
// 	assert.Nil(t, err)
// 	rmsg = common.Message{
// 		Message: packet.Message{Topic: "/中国/北京/海淀", QOS: 0, Payload: []byte("topic-11")},
// 	}
// 	err = c.setRetained(rmsg.Topic, &rmsg)
// 	assert.Nil(t, err)
// 	subTopics := []packet.Subscription{packet.Subscription{Topic: "/china/shanghai", QOS: 1},
// 		packet.Subscription{Topic: "/中国/#", QOS: 1}}
// 	subpkt := packet.Subscribe{Subscriptions: subTopics}
// 	retainedMsgs, err := c.MatchRetained(&subpkt)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 3, len(retainedMsgs))

// 	c.removeRetained("/中国/上海")
// 	retainedMsgs, err = c.MatchRetained(&subpkt)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 2, len(retainedMsgs))

// 	rmsg = common.Message{
// 		Message: packet.Message{Topic: "/中国/上海/minhang", QOS: 0, Payload: []byte("topic-11")},
// 	}
// 	err = c.setRetained(rmsg.Topic, &rmsg)
// 	assert.Nil(t, err)
// 	rmsg = common.Message{
// 		Message: packet.Message{Topic: "/中国/上海/huangpu", QOS: 0, Payload: []byte("topic-11")},
// 	}
// 	err = c.setRetained(rmsg.Topic, &rmsg)
// 	assert.Nil(t, err)
// 	subTopics = []packet.Subscription{packet.Subscription{Topic: "/中国/+/+", QOS: 1}}
// 	subpkt = packet.Subscribe{Subscriptions: subTopics}
// 	retainedMsgs, err = c.MatchRetained(&subpkt)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 4, len(retainedMsgs))

// 	subTopics = []packet.Subscription{packet.Subscription{Topic: "/中国/上海/+", QOS: 1}}
// 	subpkt = packet.Subscribe{Subscriptions: subTopics}
// 	retainedMsgs, err = c.MatchRetained(&subpkt)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 3, len(retainedMsgs))

// 	subTopics = []packet.Subscription{packet.Subscription{Topic: "/中国/#", QOS: 1},
// 		packet.Subscription{Topic: "/中国/+/+", QOS: 1}}
// 	subpkt = packet.Subscribe{Subscriptions: subTopics}
// 	retainedMsgs, err = c.MatchRetained(&subpkt)
// 	assert.Nil(t, err)
// 	assert.Equal(t, 4, len(retainedMsgs))
// }

func TestRecordWillMessage(t *testing.T) {
	pf, err := NewFactory("./var/db/")
	assert.NoError(t, err)
	dbName := strings.ToLower(t.Name()) + ".db"
	os.Remove("./var/db/" + dbName)
	db, err := pf.NewDB(dbName)
	assert.Nil(t, err)
	defer db.Close()

	topic := dbName
	c := newRecorder(db)
	id := fmt.Sprintf("client_%d", time.Now().Nanosecond())
	msg, err := c.getWill(id)
	assert.Nil(t, msg)

	msg = &packet.Message{Topic: topic, QOS: 1, Retain: true, Payload: []byte("abc")}
	err = c.setWill(id, msg)
	msg, err = c.getWill(id)
	assert.Nil(t, err)
	assert.Equal(t, msg.Topic, topic)
	assert.Equal(t, msg.Payload, []byte("abc"))

	err = c.removeWill(id)
	assert.Nil(t, err)
	msg, err = c.getWill(id)
	assert.Nil(t, msg)
}
