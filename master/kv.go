package master

import baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"

func (m *Master) SetKV(kv *baetyl.KV) error {
	return m.database.Set(kv)
}
func (m *Master) GetKV(key []byte) (*baetyl.KV, error) {
	return m.database.Get(key)
}
func (m *Master) DelKV(key []byte) error {
	return m.database.Del(key)
}
func (m *Master) ListKV(prefix []byte) (*baetyl.KVs, error) {
	return m.database.List(prefix)
}
