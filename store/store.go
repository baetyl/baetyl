package store

//go:generate mockgen -destination=../mock/store.go -package=mock github.com/baetyl/baetyl-core/store Store

// Store store interface
type Store interface {
	Get(key, result interface{}) error
	Insert(key, data interface{}) error
	Upsert(key interface{}, data interface{}) error
	Delete(key, dataType interface{}) error
}
