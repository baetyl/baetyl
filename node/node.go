package node

import (
	"encoding/json"
	"os"
	"runtime"
	"time"

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

const (
	OfflineDuration = 40 * time.Second
	KeyNodeProps    = "nodeprops"
	KubeNodeName    = "KUBE_NODE_NAME"
)

var (
	ErrParseReport = errors.New("failed to parse report struct")
)

//go:generate mockgen -destination=../mock/node.go -package=mock -source=node.go Node

type Node interface {
	Get() (m *v1.Node, err error)
	Desire(desired v1.Desire, override bool) (delta v1.Delta, err error)
	Report(reported v1.Report, override bool) (delta v1.Delta, err error)
	GetStats(ctx *routing.Context) (interface{}, error)
	GetNodeProperties(ctx *routing.Context) (interface{}, error)
	UpdateNodeProperties(ctx *routing.Context) (interface{}, error)
}

// TODO define interface and implement
// Node node
type node struct {
	tomb  utils.Tomb
	log   *log.Logger
	id    []byte
	store *bh.Store
	mqtt  *mqtt.Client
}

// NewNode create a node with shadow
func NewNode(store *bh.Store) (Node, error) {
	m := &v1.Node{
		CreationTimestamp: time.Now(),
		Desire:            v1.Desire{},
		Report: v1.Report{
			"core": v1.CoreInfo{
				GoVersion:   runtime.Version(),
				BinVersion:  utils.VERSION,
				GitRevision: utils.REVISION,
			},
		},
	}
	n := &node{
		id:    []byte("baetyl-edge-node"),
		store: store,
		log:   log.With(log.Any("core", "node")),
	}
	err := n.insert(m)
	if err != nil && errors.Cause(err) != bh.ErrKeyExists {
		return nil, errors.Trace(err)
	}
	// report some core info
	_, err = n.Report(m.Report, false)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return n, nil
}

// Get returns node model
func (n *node) Get() (m *v1.Node, err error) {
	err = n.store.Bolt().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(n.id)
		prev := b.Get(n.id)
		if len(prev) == 0 {
			return errors.Trace(bh.ErrNotFound)
		}
		m = &v1.Node{}
		return errors.Trace(json.Unmarshal(prev, m))
	})
	return
}

// TODO remove override option
// Desire update shadow desired data, then return the delta of desired and reported data
func (n *node) Desire(desired v1.Desire, override bool) (delta v1.Delta, err error) {
	err = n.store.Bolt().Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(n.id)
		prev := bucket.Get(n.id)
		if len(prev) == 0 {
			return errors.Trace(bh.ErrNotFound)
		}
		m := &v1.Node{}
		err := json.Unmarshal(prev, m)
		if err != nil {
			return errors.Trace(err)
		}
		if m.Desire == nil || override {
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
		err = bucket.Put(n.id, curr)
		if err != nil {
			return errors.Trace(err)
		}
		delta, err = m.Desire.DiffWithNil(m.Report)
		return errors.Trace(err)
	})
	return
}

// TODO remove override option
// Report update shadow reported data, then return the delta of desired and reported data
func (n *node) Report(reported v1.Report, override bool) (delta v1.Delta, err error) {
	err = n.store.Bolt().Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(n.id)
		prev := bucket.Get(n.id)
		if len(prev) == 0 {
			return errors.Trace(bh.ErrNotFound)
		}
		m := &v1.Node{}
		err = json.Unmarshal(prev, m)
		if err != nil {
			return errors.Trace(err)
		}
		if m.Report == nil || override {
			m.Report = reported
		} else {
			err = m.Report.Merge(reported)
			if err != nil {
				return errors.Trace(err)
			}
			// since merge won't delete exist key-val, node info and stats should override
			if nodeInfo, ok := reported["node"]; ok {
				m.Report["node"] = nodeInfo
			}
			if nodeStats, ok := reported["nodestats"]; ok {
				m.Report["nodestats"] = nodeStats
			}
		}
		var curr []byte
		curr, err = json.Marshal(m)
		if err != nil {
			return errors.Trace(err)
		}
		err = bucket.Put(n.id, curr)
		if err != nil {
			return errors.Trace(err)
		}
		delta, err = m.Desire.DiffWithNil(m.Report)
		return errors.Trace(err)
	})
	return
}

// GetStatus get status
// TODO: add an error handling middleware like baetyl-cloud @chensheng
func (n *node) GetStats(ctx *routing.Context) (interface{}, error) {
	nodeStat, err := n.Get()
	if err != nil {
		return nil, errors.Trace(err)
	}
	nodeStat.Name = os.Getenv(context.KeyNodeName)
	view, err := nodeStat.View(OfflineDuration)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return view, nil
}

func (n *node) GetNodeProperties(ctx *routing.Context) (interface{}, error) {
	node, err := n.Get()
	if err != nil {
		return nil, errors.Trace(err)
	}
	desireProps := map[string]interface{}{}
	if node.Desire != nil {
		if props, ok := node.Desire[v1.KeyNodeProps]; ok && props != nil {
			if desireProps, ok = props.(map[string]interface{}); !ok {
				return nil, errors.Trace(errors.New("invalid node props of desire"))
			}
		}
	}
	reportProps := map[string]interface{}{}
	if node.Report != nil {
		if props, ok := node.Report[v1.KeyNodeProps]; ok && props != nil {
			if reportProps, ok = props.(map[string]interface{}); !ok {
				return nil, errors.Trace(errors.New("invalid node props of report"))
			}
		}
	}
	return map[string]interface{}{
		"report": reportProps,
		"desire": desireProps,
	}, nil
}

func (n *node) UpdateNodeProperties(ctx *routing.Context) (interface{}, error) {
	node, err := n.Get()
	if err != nil {
		return nil, errors.Trace(err)
	}
	var delta v1.Delta
	err = json.Unmarshal(ctx.Request.Body(), &delta)
	if err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil, errors.Trace(err)
	}
	for _, v := range delta {
		if _, ok := v.(string); v != nil && !ok {
			return nil, errors.Trace(errors.New("value is not string"))
		}
	}
	if node.Report == nil {
		node.Report = map[string]interface{}{}
	}
	propsVal := node.Report[v1.KeyNodeProps]
	if propsVal == nil {
		propsVal = map[string]interface{}{}
	}
	var oldReportProps v1.Report
	oldReportProps, ok := propsVal.(map[string]interface{})
	if !ok {
		return nil, errors.Trace(errors.New("old node props is invalid"))
	}
	newReportProps, err := oldReportProps.Patch(delta)
	if err != nil {
		return nil, errors.Trace(err)
	}
	// cast is necessary
	node.Report[v1.KeyNodeProps] = map[string]interface{}(newReportProps)
	if _, err = n.Report(node.Report, true); err != nil {
		return nil, errors.Trace(err)
	}
	return newReportProps, nil
}

// Get insert the whole shadow data
func (n *node) insert(m *v1.Node) error {
	return n.store.Bolt().Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(n.id)
		if err != nil {
			return errors.Trace(err)
		}
		data := b.Get(n.id)
		if len(data) != 0 {
			return errors.Trace(bh.ErrKeyExists)
		}
		data, err = json.Marshal(m)
		if err != nil {
			return errors.Trace(err)
		}
		return errors.Trace(b.Put(n.id, data))
	})
}
