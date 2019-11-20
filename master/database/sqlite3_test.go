package database

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"os"
	"path"
	"testing"

	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	dir, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	_, err = New(Conf{Driver: "sqlite2", Source: path.Join(dir, "kv.db")})
	assert.Error(t, err)

	_, err = New(Conf{Driver: "sqlite3", Source: path.Join(dir, "kv.db")})
	assert.NoError(t, err)
}

func TestConf(t *testing.T) {
	conf := Conf{Driver: "sqlite3", Source: path.Join("test", "kv.db")}
	db := sqldb{nil, conf}
	assert.Equal(t, db.Conf(), conf)
}

func TestDatabaseSQLiteKV(t *testing.T) {
	dir, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := New(Conf{Driver: "sqlite3", Source: path.Join(dir, "kv.db")})
	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, "sqlite3", db.Conf().Driver)
	defer db.Close()

	k1 := baetyl.KV{
		Key:   []byte("k1"),
		Value: []byte("k1data"),
	}

	// Get: k1 does not exist
	v, err := db.Get(k1.Key)
	assert.NoError(t, err)
	assert.Empty(t, v.Key)
	assert.Empty(t, v.Value)

	// Put: k1 does not exist
	err = db.Set(&k1)
	assert.NoError(t, err)

	// Get: k1 exists
	v, err = db.Get(k1.Key)
	assert.NoError(t, err)
	assert.Equal(t, k1.Key, v.Key)
	assert.Equal(t, k1.Value, v.Value)

	// Put: k1 again
	err = db.Set(&k1)
	assert.NoError(t, err)

	// Put: key is empty
	err = db.Set(&baetyl.KV{})
	assert.Error(t, err)

	// Put: value is empty
	err = db.Set(&baetyl.KV{Key: []byte("baetyl")})
	assert.NoError(t, err)

	// Del: del k1
	err = db.Del(k1.Key)
	assert.NoError(t, err)

	// Del: del k1
	err = db.Del([]byte("ss"))
	assert.NoError(t, err)

	// list db
	vs, err := db.List([]byte("/"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 0)

	kv1 := &baetyl.KV{Key: []byte("/k/1"), Value: []byte("/k/1/data")}
	kv2 := &baetyl.KV{Key: []byte("/k/2"), Value: []byte("/k/2/data")}
	kv3 := &baetyl.KV{Key: []byte("/s/3"), Value: []byte("/d/3/data")}

	// put url-like key
	err = db.Set(kv1)
	assert.NoError(t, err)

	// put url-like key
	err = db.Set(kv2)
	assert.NoError(t, err)

	// put url-like key
	err = db.Set(kv3)
	assert.NoError(t, err)

	// list db
	vs, err = db.List([]byte("/"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 3)
	assert.Equal(t, vs.Kvs[0].Key, kv1.Key)
	assert.Equal(t, vs.Kvs[1].Key, kv2.Key)
	assert.Equal(t, vs.Kvs[2].Key, kv3.Key)
	assert.Equal(t, vs.Kvs[0].Value, kv1.Value)
	assert.Equal(t, vs.Kvs[1].Value, kv2.Value)
	assert.Equal(t, vs.Kvs[2].Value, kv3.Value)

	vs, err = db.List([]byte("/k"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 2)
	assert.Equal(t, vs.Kvs[0].Key, kv1.Key)
	assert.Equal(t, vs.Kvs[1].Key, kv2.Key)
	assert.Equal(t, vs.Kvs[0].Value, kv1.Value)
	assert.Equal(t, vs.Kvs[0].Value, kv1.Value)

	vs, err = db.List([]byte("/k/"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 2)
	assert.Equal(t, vs.Kvs[0].Key, kv1.Key)
	assert.Equal(t, vs.Kvs[1].Key, kv2.Key)
	assert.Equal(t, vs.Kvs[0].Value, kv1.Value)
	assert.Equal(t, vs.Kvs[0].Value, kv1.Value)

	vs, err = db.List([]byte("/kx/"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 0)

	err = db.Del(kv1.Key)
	assert.NoError(t, err)

	// list db
	vs, err = db.List([]byte("/"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 2)
	assert.Equal(t, vs.Kvs[0].Key, kv2.Key)
	assert.Equal(t, vs.Kvs[1].Key, kv3.Key)
	assert.Equal(t, vs.Kvs[0].Value, kv2.Value)
	assert.Equal(t, vs.Kvs[1].Value, kv3.Value)

	// delete k3
	err = db.Del(kv3.Key)
	assert.NoError(t, err)

	// list db
	vs, err = db.List([]byte("/kx"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 0)

	// delete k2
	err = db.Del(kv2.Key)
	assert.NoError(t, err)

	// list db
	vs, err = db.List([]byte("/"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 0)

	// test Chinese
	kvc1 := &baetyl.KV{Key: []byte("/陈/张"), Value: []byte("/陈/张里")}
	kvc2 := &baetyl.KV{Key: []byte("/陈/王"), Value: []byte("/陈/王里")}
	kvc3 := &baetyl.KV{Key: []byte("/李/王"), Value: []byte("/李/王里")}

	// put url-like key
	err = db.Set(kvc1)
	assert.NoError(t, err)

	// put url-like key
	err = db.Set(kvc2)
	assert.NoError(t, err)

	// put url-like key
	err = db.Set(kvc3)
	assert.NoError(t, err)

	// list db
	vs, err = db.List([]byte("/陈"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 2)
	assert.Equal(t, vs.Kvs[0].Key, kvc1.Key)
	assert.Equal(t, vs.Kvs[1].Key, kvc2.Key)
	assert.Equal(t, vs.Kvs[0].Value, kvc1.Value)
	assert.Equal(t, vs.Kvs[0].Value, kvc1.Value)

	vs, err = db.List([]byte("/陈/"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 2)
	assert.Equal(t, vs.Kvs[0].Key, kvc1.Key)
	assert.Equal(t, vs.Kvs[1].Key, kvc2.Key)
	assert.Equal(t, vs.Kvs[0].Value, kvc1.Value)
	assert.Equal(t, vs.Kvs[0].Value, kvc1.Value)

	vs, err = db.List([]byte("/赵/"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 0)

	vs, err = db.List([]byte("/李/"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 1)
	assert.Equal(t, vs.Kvs[0].Key, kvc3.Key)
	assert.Equal(t, vs.Kvs[0].Value, kvc3.Value)

	err = db.Del(kvc1.Key)
	assert.NoError(t, err)

	vs, err = db.List([]byte("/陈"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 1)
	assert.Equal(t, vs.Kvs[0].Key, kvc2.Key)
	assert.Equal(t, vs.Kvs[0].Value, kvc2.Value)

	err = db.Del(kvc3.Key)
	assert.NoError(t, err)

	vs, err = db.List([]byte("/李"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 0)

	vs, err = db.List([]byte("/赵"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 0)

	err = db.Del(kvc2.Key)
	assert.NoError(t, err)

	vs, err = db.List([]byte("/陈"))
	assert.NoError(t, err)
	assert.Len(t, vs.Kvs, 0)
}

func BenchmarkDatabaseSQLite(b *testing.B) {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(b, err)
	defer os.RemoveAll(dir)

	db, err := New(Conf{Driver: "sqlite3", Source: path.Join(dir, "t.db")})
	assert.NoError(b, err)
	assert.NotNil(b, db)
	defer db.Close()

	// list db
	vs, err := db.List([]byte("/"))
	assert.NoError(b, err)
	assert.Len(b, vs.Kvs, 0)

	k1 := []byte("/")
	b.ResetTimer()
	b.Run("put", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := bytes.Join([][]byte{k1, []byte("/"), int32ToBytes(i)}, []byte(""))
			db.Set(&baetyl.KV{Key: key, Value: key})
		}
	})
	b.Run("get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := bytes.Join([][]byte{k1, []byte("/"), int32ToBytes(i)}, []byte(""))
			db.Get(key)
		}
	})
	b.Run("del", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := bytes.Join([][]byte{k1, []byte("/"), int32ToBytes(i)}, []byte(""))
			db.Del(key)
		}
	})
}

func int32ToBytes(i int) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(i))
	return buf
}
