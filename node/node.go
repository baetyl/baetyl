package node

import (
	"encoding/json"
	"runtime"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	routing "github.com/qiangxue/fasthttp-routing"
	bh "github.com/timshannon/bolthold"
	bolt "go.etcd.io/bbolt"
)

const OfflineDuration = 40 * time.Second

// Node node
type Node struct {
	id    []byte
	store *bh.Store
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
	s := &Node{
		id:    []byte("baetyl-edge-node"),
		store: store,
	}
	err := s.insert(m)
	if err != nil && errors.Cause(err) != bh.ErrKeyExists {
		return nil, errors.Trace(err)
	}
	// report some core info
	_, err = s.Report(m.Report)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return s, nil
}

// Get returns node model
func (s *Node) Get() (m *v1.Node, err error) {
	err = s.store.Bolt().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.id)
		prev := b.Get(s.id)
		if len(prev) == 0 {
			return errors.Trace(bh.ErrNotFound)
		}
		m = &v1.Node{}
		return errors.Trace(json.Unmarshal(prev, m))
	})
	return
}

// Desire update shadow desired data, then return the delta of desired and reported data
func (s *Node) Desire(desired v1.Desire) (delta v1.Desire, err error) {
	err = s.store.Bolt().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.id)
		prev := b.Get(s.id)
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
		err = b.Put(s.id, curr)
		if err != nil {
			return errors.Trace(err)
		}
		delta, err = m.Desire.Diff(m.Report)
		return errors.Trace(err)
	})
	return
}

// Report update shadow reported data, then return the delta of desired and reported data
func (s *Node) Report(reported v1.Report) (delta v1.Desire, err error) {
	err = s.store.Bolt().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.id)
		prev := b.Get(s.id)
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
		err = b.Put(s.id, curr)
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
func (s *Node) GetStatus(ctx *routing.Context) error {
	node, err := s.Get()
	if err != nil {
		http.RespondMsg(ctx, 500, "UnknownError", err.Error())
		return nil
	}

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

// Get insert the whole shadow data
func (s *Node) insert(m *v1.Node) error {
	return s.store.Bolt().Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(s.id)
		if err != nil {
			return errors.Trace(err)
		}
		data := b.Get(s.id)
		if len(data) != 0 {
			return errors.Trace(bh.ErrKeyExists)
		}
		data, err = json.Marshal(m)
		if err != nil {
			return errors.Trace(err)
		}
		return errors.Trace(b.Put(s.id, data))
	})
}
