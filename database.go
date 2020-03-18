package baetyl

import (
	"errors"
	"io"
	"path/filepath"

	schema "github.com/baetyl/baetyl/schema/v3"
)

type DatabaseFactory = func(Context, string) (Database, error)

const defaultDB = "kv.db"

var dfs = map[string]DatabaseFactory{}

// DB the backend database
type Database interface {
	io.Closer
	Context() Context

	Set(kv *schema.KV) error
	Get(key []byte) (*schema.KV, error)
	Del(key []byte) error
	List(prefix []byte) (*schema.KVs, error)
}

func (rt *runtime) openDB() error {
	if f, ok := dfs[rt.cfg.Database.Driver]; ok {
		db, err := f(rt, filepath.Join(rt.cfg.DataPath, defaultDB))
		if err != nil {
			return err
		}
		rt.db = db
		rt.log.Infoln("database started")
		return nil
	}
	return errors.New("no such kind database")
}

func (rt *runtime) closeDB() {
	rt.db.Close()
	rt.log.Infoln("database stopped")
}

// RegisterDatabase by name
func RegisterDatabase(name string, factory DatabaseFactory) DatabaseFactory {
	odf := dfs[name]
	dfs[name] = factory
	return odf
}
