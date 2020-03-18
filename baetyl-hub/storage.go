package hub

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// KV key-value pair
type KV struct {
	Key   []byte
	Value []byte
}

// Database persistence interfaces
type Database interface {
	Sequence() (uint64, error)
	// Put(key, value []byte) error
	Get(key []byte) ([]byte, error)
	// Fetch(offset []byte) ([]byte, []byte, error)
	Delete(key []byte) error
	Clean(timestamp uint64) (uint64, error)
	Close()

	BatchPut(kvs []*KV) error
	BatchPutV(vs [][]byte) error
	BatchFetch(offset []byte, size int) ([]*KV, error)

	BucketPut(bucket, key, value []byte) error
	BucketGet(bucket, key []byte) ([]byte, error)
	BucketList(bucket []byte) (map[string][]byte, error)
	BucketDelete(bucket, key []byte) error
}

type storage struct {
	dir string
	dbs map[string]Database
	sync.Mutex
}

func (h *hub) startStorage() error {
	dir := h.cfg.Storage.Dir
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to make dir (%s): %s", dir, err.Error())
	}
	h.storage = &storage{dir: dir, dbs: make(map[string]Database)}
	return nil
}

func (h *hub) stopStorage() {
	h.storage.Lock()
	defer h.storage.Unlock()
	for _, db := range h.storage.dbs {
		db.Close()
	}
}

func (f *storage) newDB(name string) (Database, error) {
	if name == "" {
		return nil, fmt.Errorf("name (%s) invalid", name)
	}
	f.Lock()
	defer f.Unlock()
	if db, ok := f.dbs[name]; ok {
		return db, nil
	}
	db, err := newBoltDB(filepath.Join(f.dir, name))
	if err != nil {
		return nil, err
	}
	f.dbs[name] = db
	return db, nil
}
