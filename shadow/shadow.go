package shadow

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/baetyl/baetyl-core/models"
	jsonpatch "github.com/evanphx/json-patch"
	bh "github.com/timshannon/bolthold"
	bolt "go.etcd.io/bbolt"
)

// Shadow shadow
type Shadow struct {
	bk    []byte
	id    []byte
	depth int
	store *bh.Store
}

// NewShadow create a new shadow
func NewShadow(namespace, name string, store *bh.Store) (*Shadow, error) {
	m := &models.Shadow{
		Name:              name,
		Namespace:         namespace,
		CreationTimestamp: time.Now(),
		Reported:          map[string]interface{}{},
		Desired:           map[string]interface{}{},
	}
	s := &Shadow{
		bk:    []byte("baetyl-edge-shadow"),
		id:    []byte(name + "." + namespace),
		depth: 4,
		store: store,
	}
	err := s.insert(m)
	if err != nil && err != bh.ErrKeyExists {
		return nil, err
	}
	return s, nil
}

// Get returns shadow model
func (s *Shadow) Get() (m *models.Shadow, err error) {
	err = s.store.Bolt().View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bk)
		prev := b.Get(s.id)
		if len(prev) == 0 {
			return bh.ErrNotFound
		}
		m = &models.Shadow{}
		return json.Unmarshal(prev, m)
	})
	return
}

// Desire update shadow desired data, then return the delta of desired and reported data
func (s *Shadow) Desire(desired map[string]interface{}) (delta map[string]interface{}, err error) {
	err = s.store.Bolt().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bk)
		prev := b.Get(s.id)
		if len(prev) == 0 {
			return bh.ErrNotFound
		}
		m := &models.Shadow{}
		err := json.Unmarshal(prev, m)
		if err != nil {
			return err
		}
		if m.Desired == nil {
			m.Desired = desired
		} else {
			merge(m.Desired, desired, 0, s.depth)
		}
		curr, err := json.Marshal(m)
		if err != nil {
			return err
		}
		err = b.Put(s.id, curr)
		if err != nil {
			return err
		}
		delta, err = diff(m, s.depth)
		return err
	})
	return
}

// Report update shadow reported data, then return the delta of desired and reported data
func (s *Shadow) Report(reported map[string]interface{}) (delta map[string]interface{}, err error) {
	err = s.store.Bolt().Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bk)
		prev := b.Get(s.id)
		if len(prev) == 0 {
			return bh.ErrNotFound
		}
		m := &models.Shadow{}
		err := json.Unmarshal(prev, m)
		if err != nil {
			return err
		}
		if m.Reported == nil {
			m.Reported = reported
		} else {
			merge(m.Reported, reported, 0, s.depth)
		}
		curr, err := json.Marshal(m)
		if err != nil {
			return err
		}
		err = b.Put(s.id, curr)
		if err != nil {
			return err
		}
		delta, err = diff(m, s.depth)
		return err
	})
	return
}

// Get insert the whole shadow data
func (s *Shadow) insert(m *models.Shadow) error {
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

// merge right map into left map
func merge(left, right map[string]interface{}, depth, maxDepth int) {
	if depth >= maxDepth {
		return
	}
	for rk, rv := range right {
		lv, ok := left[rk]
		if !ok || reflect.TypeOf(rv).Kind() != reflect.Map || reflect.TypeOf(lv).Kind() != reflect.Map {
			left[rk] = rv
			continue
		}
		merge(lv.(map[string]interface{}), rv.(map[string]interface{}), depth+1, maxDepth)
	}
}

func diff(shadow *models.Shadow, maxDepth int) (map[string]interface{}, error) {
	var delta map[string]interface{}
	reported, err := json.Marshal(shadow.Reported)
	if err != nil {
		return delta, err
	}
	desired, err := json.Marshal(shadow.Desired)
	if err != nil {
		return delta, err
	}
	patch, err := jsonpatch.CreateMergePatch(reported, desired)
	if err != nil {
		return delta, err
	}
	err = json.Unmarshal(patch, &delta)
	if err != nil {
		return delta, err
	}
	clean(delta, 0, maxDepth)
	return delta, nil
}

func clean(m map[string]interface{}, depth, maxDepth int) {
	if depth >= maxDepth {
		return
	}
	for k, v := range m {
		if v == nil {
			delete(m, k)
			continue
		}
		bk := reflect.TypeOf(v).Kind()
		if bk != reflect.Map {
			continue
		}
		if vm, ok := v.(map[string]interface{}); ok {
			clean(vm, depth+1, maxDepth)
		}
	}
}
