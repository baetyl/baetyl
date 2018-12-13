package persist

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

// Factory persistence factory
type Factory struct {
	dir string
	dbs map[string]Database
	sync.Mutex
}

// NewFactory creates a persistence factory
func NewFactory(dir string) (*Factory, error) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}
	return &Factory{dir: dir, dbs: make(map[string]Database)}, nil
}

// NewDB creates a persistence database
func (f *Factory) NewDB(name string) (Database, error) {
	if name == "" {
		return nil, fmt.Errorf("name (%s) invalid", name)
	}
	f.Lock()
	defer f.Unlock()
	if db, ok := f.dbs[name]; ok {
		return db, nil
	}
	db, err := NewBoltDB(filepath.Join(f.dir, name))
	if err == nil {
		f.dbs[name] = db
	}
	return db, err
}

// Close close factory to release all databases
func (f *Factory) Close() {
	f.Lock()
	defer f.Unlock()
	for _, db := range f.dbs {
		db.Close()
	}
}
