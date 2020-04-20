package store

import (
	"encoding/json"
	"os"
	"path"

	bh "github.com/timshannon/bolthold"
)

// NewBoltHold creates a new bolt hold
func NewBoltHold(filename string) (*bh.Store, error) {
	err := os.MkdirAll(path.Dir(filename), 0755)
	if err != nil {
		return nil, err
	}
	ops := &bh.Options{
		Encoder: func(value interface{}) ([]byte, error) {
			return json.Marshal(value)
		},
		Decoder: func(data []byte, value interface{}) error {
			return json.Unmarshal(data, value)
		},
	}
	return bh.Open(filename, 0666, ops)
}
