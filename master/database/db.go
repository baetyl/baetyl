package database

import (
	"errors"
	"io"

	"github.com/baetyl/baetyl/sdk/baetyl-go/api"
)

// Factories of database
var Factories = map[string]func(conf Conf) (DB, error){}

// DB the backend database
type DB interface {
	Conf() Conf

	Set(kv *api.KV) error
	Get(key []byte) (*api.KV, error)
	Del(key []byte) error
	List(prefix []byte) (*api.KVs, error)

	io.Closer
}

// Conf the configuration of database
type Conf struct {
	Driver string
	Source string
}

// New KV database by given name
func New(conf Conf) (DB, error) {
	if f, ok := Factories[conf.Driver]; ok {
		return f(conf)
	}
	return nil, errors.New("no such kind database")
}
