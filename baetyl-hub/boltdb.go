package hub

import (
	"time"

	"github.com/baetyl/baetyl/baetyl-hub/utils"
	bolt "github.com/etcd-io/bbolt"
)

// boltDB use boltdb to persist data
type boltDB struct {
	*bolt.DB
	bucket []byte
}

func newBoltDB(path string) (*boltDB, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}
	return &boltDB{DB: db, bucket: []byte(".self")}, nil
}

// Sequence returns the sequence id
func (p *boltDB) Sequence() (sid uint64, err error) {
	err = p.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(p.bucket)
		if b != nil {
			sid = b.Sequence()
		}
		return nil
	})
	return
}

// Put puts a KV
func (p *boltDB) Put(key, value []byte) error {
	return p.BucketPut(p.bucket, key, value)
}

// Get gets a KV by key
func (p *boltDB) Get(key []byte) (value []byte, err error) {
	return p.BucketGet(p.bucket, key)
}

// Fetch fetches a KV by offset
func (p *boltDB) Fetch(offset []byte) (key, value []byte, err error) {
	err = p.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(p.bucket)
		if b == nil {
			return nil
		}
		ik, iv := b.Cursor().Seek(offset)
		if len(ik) == 0 || len(iv) == 0 {
			return nil
		}
		key = make([]byte, len(ik))
		value = make([]byte, len(iv))
		copy(key, ik)   // copy
		copy(value, iv) // copy
		return nil
	})
	return
}

// Delete deletes a KV
func (p *boltDB) Delete(key []byte) error {
	return p.BucketDelete(p.bucket, key)
}

// Clean cleans all KVs before timestamp
func (p *boltDB) Clean(timestamp uint64) (count uint64, err error) {
	err = p.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(p.bucket)
		if b == nil {
			return nil
		}
		c := b.Cursor()
		for k, _ := c.First(); k != nil && utils.U64(k[8:]) < timestamp; k, _ = c.Next() {
			err = b.Delete(k)
			if err != nil {
				return err
			}
			count++
		}
		return nil
	})
	return
}

// BatchPut puts KVs in batch mode
func (p *boltDB) BatchPut(kvs []*KV) error {
	return p.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(p.bucket)
		if err != nil {
			return err
		}
		for _, kv := range kvs {
			err = b.Put(kv.Key, kv.Value)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// BatchPutV puts values in batch mode
func (p *boltDB) BatchPutV(vs [][]byte) error {
	return p.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(p.bucket)
		if err != nil {
			return err
		}
		// key = sid + ts (16 bytes)
		ts := uint64(time.Now().Unix())
		for _, v := range vs {
			sid, err := b.NextSequence()
			if err != nil {
				return err
			}
			err = b.Put(utils.U64U64ToB(sid, ts), v)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// BucketPut puts a KV into bucket
func (p *boltDB) BucketPut(bucket, key, value []byte) error {
	return p.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		return b.Put(key, value)
	})
}

// BucketGet gets a KV from bucket by key
func (p *boltDB) BucketGet(bucket, key []byte) (value []byte, err error) {
	err = p.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		iv := b.Get(key)
		if len(iv) == 0 {
			return nil
		}
		value = make([]byte, len(iv))
		copy(value, iv) // copy
		return nil
	})
	return
}

// BatchFetch fetches KVs by offset in batch mode
func (p *boltDB) BatchFetch(offset []byte, size int) ([]*KV, error) {
	res := make([]*KV, 0)
	err := p.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(p.bucket)
		if b == nil {
			return nil
		}
		i := int(0)
		c := b.Cursor()
		for ik, iv := c.Seek(offset); i < size && len(ik) != 0 && len(iv) != 0; ik, iv = c.Next() {
			key := make([]byte, len(ik))
			value := make([]byte, len(iv))
			copy(key, ik)   // copy
			copy(value, iv) // copy
			res = append(res, &KV{Key: key, Value: value})
			i++
		}
		return nil
	})
	return res, err
}

// BucketDelete deletes a KV in bucket by key
func (p *boltDB) BucketDelete(bucket, key []byte) error {
	return p.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		return b.Delete(key)
	})
}

// BucketList lists all KVs in bucket
func (p *boltDB) BucketList(bucket []byte) (map[string][]byte, error) {
	res := make(map[string][]byte)
	err := p.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		if b == nil {
			return nil
		}
		return b.ForEach(func(ik, iv []byte) error {
			v := make([]byte, len(iv))
			copy(v, iv) // copy
			res[string(ik)] = v
			return nil
		})
	})
	return res, err
}

// Close closes database
func (p *boltDB) Close() {
	p.DB.Close()
}
