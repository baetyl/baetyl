package shadow

import (
	"encoding/json"
	"time"

	shad "github.com/baetyl/baetyl-go/shadow"
	bh "github.com/timshannon/bolthold"
	bolt "go.etcd.io/bbolt"
)

// Shadow shadow
type Shadow struct {
	bk    []byte
	id    []byte
	store *bh.Store
}

// NewShadow create a new shadow
func NewShadow(namespace, name string, store *bh.Store) (*Shadow, error) {
	m := &shad.Spec{
		Name:              name,
		Namespace:         namespace,
		CreationTimestamp: time.Now(),
		Report:            shad.Report{},
		Desire:            shad.Desire{},
	}
	s := &Shadow{
		bk:    []byte("baetyl-edge-shadow"),
		id:    []byte(name + "." + namespace),
		store: store,
	}
	err := s.insert(m)
	if err != nil && err != bh.ErrKeyExists {
		return nil, err
	}
	return s, nil
}

// Get returns shadow model
func (s *Shadow) Get() (m *shad.Spec, err error) {
	err = s.store.Bolt().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bk)
		prev := b.Get(s.id)
		if len(prev) == 0 {
			return bh.ErrNotFound
		}
		m = &shad.Spec{}
		return json.Unmarshal(prev, m)
	})
	return
}

// Desire update shadow desired data, then return the delta of desired and reported data
func (s *Shadow) Desire(desired shad.Desire) (delta shad.Delta, err error) {
	err = s.store.Bolt().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bk)
		prev := b.Get(s.id)
		if len(prev) == 0 {
			return bh.ErrNotFound
		}
		m := &shad.Spec{}
		err := json.Unmarshal(prev, m)
		if err != nil {
			return err
		}
		if m.Desire == nil {
			m.Desire = desired
		} else {
			err = m.Desire.Merge(desired)
			if err != nil {
				return err
			}
		}
		curr, err := json.Marshal(m)
		if err != nil {
			return err
		}
		err = b.Put(s.id, curr)
		if err != nil {
			return err
		}
		delta, err = m.Desire.Diff(m.Report)
		return err
	})
	return
}

// Report update shadow reported data, then return the delta of desired and reported data
func (s *Shadow) Report(reported shad.Report) (delta shad.Delta, err error) {
	err = s.store.Bolt().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bk)
		prev := b.Get(s.id)
		if len(prev) == 0 {
			return bh.ErrNotFound
		}
		m := &shad.Spec{}
		err := json.Unmarshal(prev, m)
		if err != nil {
			return err
		}
		if m.Report == nil {
			m.Report = reported
		} else {
			err = m.Report.Merge(reported)
			if err != nil {
				return err
			}
		}
		curr, err := json.Marshal(m)
		if err != nil {
			return err
		}
		err = b.Put(s.id, curr)
		if err != nil {
			return err
		}
		delta, err = m.Desire.Diff(m.Report)
		return err
	})
	return
}

// Get insert the whole shadow data
func (s *Shadow) insert(m *shad.Spec) error {
	return s.store.Bolt().Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(s.bk)
		if err != nil {
			return err
		}
		data := b.Get(s.id)
		if len(data) != 0 {
			return bh.ErrKeyExists
		}
		data, err = json.Marshal(m)
		if err != nil {
			return err
		}
		return b.Put(s.id, data)
	})
}
