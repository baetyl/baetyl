package store

import (
	"encoding/json"
	"os"
	"path"

	"github.com/baetyl/baetyl-go/errors"
	bh "github.com/timshannon/bolthold"
)

// NewBoltHold creates a new bolt hold
func NewBoltHold(filename string) (*bh.Store, error) {
	err := os.MkdirAll(path.Dir(filename), 0755)
	if err != nil {
		return nil, errors.Trace(err)
	}
	ops := &bh.Options{
		Encoder: func(value interface{}) ([]byte, error) {
			return json.Marshal(value)
		},
		Decoder: func(data []byte, value interface{}) error {
			return json.Unmarshal(data, value)
		},
	}
	s, err := bh.Open(filename, 0666, ops)
	return s, errors.Trace(err)
}
