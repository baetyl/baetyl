package master

import (
	"github.com/baetyl/baetyl/master/database"
)

// PutKV put key and value into SQL DB
func (m *Master) PutKV(key string, value []byte) error {
	return m.db.PutKV(key, value)
}

// GetKV gets value by key from SQL DB
func (m *Master) GetKV(key string) (result database.KV, err error) {
	return m.db.GetKV(key)
}

// DelKV deletes key and value from SQL DB
func (m *Master) DelKV(key string) error {
	return m.db.DelKV(key)
}

// ListKV list kvs under the prefix
func (m *Master) ListKV(prefix string) (results []database.KV, err error) {
	return m.db.ListKV(prefix)
}
