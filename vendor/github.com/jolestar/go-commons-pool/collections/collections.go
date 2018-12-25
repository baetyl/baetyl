package collections

import (
	"reflect"
	"sync"
)

// Iterator interface for collection
// see LinkedBlockDequeIterator
type Iterator interface {
	HasNext() bool
	Next() interface{}
	Remove()
}

// SyncIdentityMap is a concurrent safe map
// use key's pointer as map key
type SyncIdentityMap struct {
	sync.RWMutex
	m map[uintptr]interface{}
}

// NewSyncMap return a new SyncIdentityMap
func NewSyncMap() *SyncIdentityMap {
	return &SyncIdentityMap{m: make(map[uintptr]interface{})}
}

// Get by key
func (m *SyncIdentityMap) Get(key interface{}) interface{} {
	m.RLock()
	keyPtr := genKey(key)
	value := m.m[keyPtr]
	m.RUnlock()
	return value
}

func genKey(key interface{}) uintptr {
	keyValue := reflect.ValueOf(key)
	return keyValue.Pointer()
}

// Put key and value to map
func (m *SyncIdentityMap) Put(key interface{}, value interface{}) {
	m.Lock()
	keyPtr := genKey(key)
	m.m[keyPtr] = value
	m.Unlock()
}

// Remove value by key
func (m *SyncIdentityMap) Remove(key interface{}) {
	m.Lock()
	keyPtr := genKey(key)
	delete(m.m, keyPtr)
	m.Unlock()
}

// Size return map len, and is concurrent safe
func (m *SyncIdentityMap) Size() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.m)
}

// Values copy all map's value to slice
func (m *SyncIdentityMap) Values() []interface{} {
	m.RLock()
	defer m.RUnlock()
	list := make([]interface{}, len(m.m))
	i := 0
	for _, v := range m.m {
		list[i] = v
		i++
	}
	return list
}
