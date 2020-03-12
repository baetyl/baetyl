package store

import (
	bh "github.com/timshannon/bolthold"
)

// NewBoltHold creates a new bolt hold
func NewBoltHold(filename string) (Store, error) {
	return bh.Open(filename, 0666, nil)
}
