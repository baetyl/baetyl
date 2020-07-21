package store

import (
	"encoding/json"
	"os"
	"path"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	bh "github.com/timshannon/bolthold"
	bolt "go.etcd.io/bbolt"
)

// NewBoltHold creates a new bolt hold
func NewBoltHold(filename string) (*bh.Store, error) {
	err := os.MkdirAll(path.Dir(filename), 0755)
	if err != nil {
		return nil, errors.Trace(err)
	}
	ops := &bh.Options{}
	ops.Options = bolt.DefaultOptions
	ops.Timeout = time.Second * 10
	ops.Encoder = func(value interface{}) ([]byte, error) {
		return json.Marshal(value)
	}
	ops.Decoder = func(data []byte, value interface{}) error {
		return json.Unmarshal(data, value)
	}
	sto, err := bh.Open(filename, 0666, ops)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return sto, nil
}
