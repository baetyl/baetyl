package store

import (
	"encoding/json"

	bh "github.com/timshannon/bolthold"
)

// NewBoltHold creates a new bolt hold
func NewBoltHold(filename string) (*bh.Store, error) {
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
