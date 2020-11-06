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
const CheckNodeTwinInterval = 40 * time.Second
const NodeTwinNotifyTopic = "$baetyl/nodetwin"
const KeyNodeTwin = "nodeTwin"

// Node node
type Node struct {
	tomb  utils.Tomb
	log   *log.Logger
	id    []byte
	store *bh.Store
	mqtt  *mqtt.Client
}

// NewNode create a node with shadow
func NewNode(store *bh.Store, ctx context.Context) (*Node, error) {
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
	if ctx != nil {
		// TODO support qos equal 1
		nod.mqtt, err = ctx.NewSystemBrokerClient(nil)
		if err != nil {
			return nil, err
		}
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

func (nod *Node) GetNodeTwin(ctx *routing.Context) error {
	node, err := nod.Get()
	if err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	desire, ok := node.Desire[KeyNodeTwin]
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

func (nod *Node) UpdateNodeTwin(ctx *routing.Context) error {
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
	val := node.Report[KeyNodeTwin]
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
	node.Report[KeyNodeTwin] = map[string]interface{}(newReport)
	if _, err = nod.Report(node.Report); err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}
	return nil
}

func (nod *Node) start() error {
	if nod.mqtt == nil {
		nod.log.Info("mqtt client is nil and won't check node twin")
		return nil
	}
	nod.log.Info("node starts to check node twin")
	defer nod.log.Info("node stop check node twin")

	t := time.NewTicker(CheckNodeTwinInterval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			err := nod.checkNodeTwin()
			if err != nil {
				nod.log.Error("failed to check node twin", log.Error(err))
			} else {
				nod.log.Debug("check node twin")
			}
		case <-nod.tomb.Dying():
			return nil
		}
	}
}

func (nod *Node) Start() {
	if nod.mqtt != nil {
		nod.mqtt.Start(nil)
	}
	nod.tomb.Go(nod.start)
}

func (nod *Node) Close() error {
	nod.tomb.Kill(nil)
	if err := nod.tomb.Wait(); err != nil {
		return err
	}
	if nod.mqtt != nil {
		return nod.mqtt.Close()
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
	val, ok := node.Report[KeyNodeTwin]
	if ok {
		report, ok = val.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid node twin of report")
		}
	} else {
		report = map[string]interface{}{}
	}
	val, ok = node.Desire[KeyNodeTwin]
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
	pkt.Message.Topic = NodeTwinNotifyTopic
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
