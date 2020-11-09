package node

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/256dpi/gomqtt/packet"
	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mqtt"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	routing "github.com/qiangxue/fasthttp-routing"
	bh "github.com/timshannon/bolthold"
	bolt "go.etcd.io/bbolt"
)

const OfflineDuration = 40 * time.Second
const NodePropsNotifyTopic = "$baetyl/node/props"
const KeyNodeProps = "nodeProps"

// Node node
type Node struct {
	tomb  utils.Tomb
	log   *log.Logger
	id    []byte
	store *bh.Store
	mqtt  *mqtt.Client
}

// NewNode create a node with shadow
func NewNode(store *bh.Store) (*Node, error) {
	m := &v1.Node{
		CreationTimestamp: time.Now(),
		Desire:            v1.Desire{},
		Report: v1.Report{
			"core": v1.CoreInfo{
				GoVersion:   runtime.Version(),
				BinVersion:  utils.VERSION,
				GitRevision: utils.REVISION,
			},
			"node":      nil,
			"nodestats": nil,
			"apps":      nil,
			"sysapps":   nil,
			"appstats":  nil,
		},
	}
	nod := &Node{
		id:    []byte("baetyl-edge-node"),
		store: store,
		log:   log.With(log.Any("core", "node")),
	}
	err := nod.insert(m)
	if err != nil && errors.Cause(err) != bh.ErrKeyExists {
		return nil, errors.Trace(err)
	}
	// report some core info
	_, err = nod.Report(m.Report)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return nod, nil
}

// Get returns node model
func (nod *Node) Get() (m *v1.Node, err error) {
	err = nod.store.Bolt().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(nod.id)
		prev := b.Get(nod.id)
		if len(prev) == 0 {
			return errors.Trace(bh.ErrNotFound)
		}
		m = &v1.Node{}
		return errors.Trace(json.Unmarshal(prev, m))
	})
	return
}

// Desire update shadow desired data, then return the delta of desired and reported data
func (nod *Node) Desire(desired v1.Desire) (delta v1.Desire, err error) {
	err = nod.store.Bolt().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(nod.id)
		prev := b.Get(nod.id)
		if len(prev) == 0 {
			return errors.Trace(bh.ErrNotFound)
		}
		m := &v1.Node{}
		err := json.Unmarshal(prev, m)
		if err != nil {
			return errors.Trace(err)
		}
		if m.Desire == nil {
			m.Desire = desired
		} else {
			err = m.Desire.Merge(desired)
			if err != nil {
				return errors.Trace(err)
			}
		}
		curr, err := json.Marshal(m)
		if err != nil {
			return errors.Trace(err)
		}
		err = b.Put(nod.id, curr)
		if err != nil {
			return errors.Trace(err)
		}
		delta, err = m.Desire.Diff(m.Report)
		return errors.Trace(err)
	})
	return
}

// Report update shadow reported data, then return the delta of desired and reported data
func (nod *Node) Report(reported v1.Report) (delta v1.Desire, err error) {
	err = nod.store.Bolt().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(nod.id)
		prev := b.Get(nod.id)
		if len(prev) == 0 {
			return errors.Trace(bh.ErrNotFound)
		}
		m := &v1.Node{}
		err := json.Unmarshal(prev, m)
		if err != nil {
			return errors.Trace(err)
		}
		if m.Report == nil {
			m.Report = reported
		} else {
			err = m.Report.Merge(reported)
			if err != nil {
				return errors.Trace(err)
			}
		}
		curr, err := json.Marshal(m)
		if err != nil {
			return errors.Trace(err)
		}
		err = b.Put(nod.id, curr)
		if err != nil {
			return errors.Trace(err)
		}
		delta, err = m.Desire.Diff(m.Report)
		return errors.Trace(err)
	})
	return
}

// GetStatus get status
// TODO: add an error handling middleware like baetyl-cloud @chensheng
func (nod *Node) GetStats(ctx *routing.Context) error {
	node, err := nod.Get()
	if err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	node.Name = os.Getenv(context.KeyNodeName)
	view, err := node.View(OfflineDuration)
	if err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	res, err := json.Marshal(view)
	if err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	http.Respond(ctx, 200, res)
	return nil
}

func (nod *Node) GetNodeProperties(ctx *routing.Context) error {
	node, err := nod.Get()
	if err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	desire, ok := node.Desire[KeyNodeProps]
	if !ok {
		http.RespondMsg(ctx, 500, "UnknownError", "node twin not exist")
		return nil
	}
	res, err := json.Marshal(desire)
	if err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	http.Respond(ctx, 200, res)
	return nil
}

func (nod *Node) UpdateNodeProperties(ctx *routing.Context) error {
	node, err := nod.Get()
	if err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	var delta v1.Delta
	err = json.Unmarshal(ctx.Request.Body(), &delta)
	if err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	var oldReport v1.Report
	val := node.Report[KeyNodeProps]
	if val == nil {
		val = map[string]interface{}{}
	}
	oldReport, ok := val.(map[string]interface{})
	if !ok {
		http.RespondMsg(ctx, 500, "UnknownError", "old node twin is invalid")
		return nil
	}
	newReport, err := oldReport.Patch(delta)
	if err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	node.Report[KeyNodeProps] = map[string]interface{}(newReport)
	if _, err = nod.Report(node.Report); err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	return nil
}

func (nod *Node) checkNodeTwin() error {
	node, err := nod.Get()
	if err != nil {
		return err
	}
	var report v1.Report
	var desire v1.Desire
	val, ok := node.Report[KeyNodeProps]
	if ok {
		report, ok = val.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid node twin of report")
		}
	} else {
		report = map[string]interface{}{}
	}
	val, ok = node.Desire[KeyNodeProps]
	if ok {
		desire, ok = val.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid node twin of desire")
		}
	} else {
		desire = map[string]interface{}{}
	}

	diff, err := desire.DiffWithNil(report)
	if err != nil {
		return err
	}
	if len(diff) == 0 {
		return nil
	}
	pld, err := json.Marshal(diff)
	if err != nil {
		return err
	}
	pkt := packet.NewPublish()
	pkt.Message.Topic = NodePropsNotifyTopic
	pkt.Message.Payload = pld
	err = nod.mqtt.Send(pkt)
	if err != nil {
		return err
	}
	return nil
}

// Get insert the whole shadow data
func (nod *Node) insert(m *v1.Node) error {
	return nod.store.Bolt().Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(nod.id)
		if err != nil {
			return errors.Trace(err)
		}
		data := b.Get(nod.id)
		if len(data) != 0 {
			return errors.Trace(bh.ErrKeyExists)
		}
		data, err = json.Marshal(m)
		if err != nil {
			return errors.Trace(err)
		}
		return errors.Trace(b.Put(nod.id, data))
	})
}
