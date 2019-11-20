package database

import (
	"errors"
	"io"
)

// Factories of database
var Factories = map[string]func(conf Conf) (DB, error){}

// KV kv object
type KV struct {
	Key   []byte
	Value []byte
}

// DB the backend database
type DB interface {
	Conf() Conf

	PutKV(key, value []byte) error
	GetKV(key []byte) (result KV, err error)
	DelKV(key []byte) error
	ListKV(prefix []byte) (results []KV, err error)

	io.Closer
}

// Conf the configuration of database
type Conf struct {
	Driver string
	Source string
}

// New engine by given name
func New(conf Conf) (DB, error) {
	if f, ok := Factories[conf.Driver]; ok {
		return f(conf)
	}
	return nil, errors.New("no such kind database")
}
