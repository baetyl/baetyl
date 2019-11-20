package database

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"os"
	"path"
	"testing"

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

	k1 := []byte("k1")
	k2 := []byte("k2")

	// k1 does not exist
	v, err := db.GetKV(k1)
	assert.NoError(t, err)
	assert.Empty(t, v.Key)
	assert.Nil(t, v.Value)

	// put k1
	err = db.PutKV(k1, k1)
	assert.NoError(t, err)

	// k1 exists
	v, err = db.GetKV(k1)
	assert.NoError(t, err)
	assert.Equal(t, k1, v.Key)
	assert.Equal(t, k1, v.Value)

	// put k1 again
	err = db.PutKV(k1, k2)
	assert.NoError(t, err)

	// k1 exists
	v, err = db.GetKV(k1)
	assert.NoError(t, err)
	assert.Equal(t, k1, v.Key)
	assert.Equal(t, k2, v.Value)

	// del k1
	err = db.DelKV(k1)
	assert.NoError(t, err)

	// k1 does not exist
	v, err = db.GetKV(k1)
	assert.NoError(t, err)
	assert.Empty(t, v.Key)
	assert.Nil(t, v.Value)

	// key is ""
	k1 = []byte("")
	err = db.PutKV(k1, k2)
	assert.NoError(t, err)

	// list db
	vs, err := db.ListKV([]byte("/"))
	assert.NoError(t, err)
	assert.Len(t, vs, 0)

	k1 = []byte("/k/1")
	k2 = []byte("/k/2")
	k3 := []byte("/kk/2")

	// put url-like key
	err = db.PutKV(k1, k1)
	assert.NoError(t, err)

	// put url-like key
	err = db.PutKV(k2, k2)
	assert.NoError(t, err)

	// put url-like key
	err = db.PutKV(k3, k3)
	assert.NoError(t, err)

	// list db
	vs, err = db.ListKV([]byte("/"))
	assert.NoError(t, err)
	assert.Len(t, vs, 3)
	assert.Equal(t, vs[0].Key, k1)
	assert.Equal(t, vs[0].Value, k1)
	assert.Equal(t, vs[1].Key, k2)
	assert.Equal(t, vs[1].Value, k2)
	assert.Equal(t, vs[2].Key, k3)
	assert.Equal(t, vs[2].Value, k3)

	vs, err = db.ListKV([]byte("/k"))
	assert.NoError(t, err)
	assert.Len(t, vs, 2)
	assert.Equal(t, vs[0].Key, k1)
	assert.Equal(t, vs[0].Value, k1)
	assert.Equal(t, vs[1].Key, k2)
	assert.Equal(t, vs[1].Value, k2)

	vs, err = db.ListKV([]byte("/k/"))
	assert.NoError(t, err)
	assert.Len(t, vs, 2)
	assert.Equal(t, vs[0].Key, k1)
	assert.Equal(t, vs[0].Value, k1)
	assert.Equal(t, vs[1].Key, k2)
	assert.Equal(t, vs[1].Value, k2)

	vs, err = db.ListKV([]byte("/kx/"))
	assert.NoError(t, err)
	assert.Len(t, vs, 0)

	// delete k1
	err = db.DelKV(k1)
	assert.NoError(t, err)

	// list db
	vs, err = db.ListKV([]byte("/"))
	assert.NoError(t, err)
	assert.Len(t, vs, 2)
	assert.Equal(t, vs[0].Key, k2)
	assert.Equal(t, vs[0].Value, k2)
	assert.Equal(t, vs[1].Key, k3)
	assert.Equal(t, vs[1].Value, k3)

	// delete k3
	err = db.DelKV(k3)
	assert.NoError(t, err)

	// list db
	vs, err = db.ListKV([]byte("/kx"))
	assert.NoError(t, err)
	assert.Len(t, vs, 0)

	// delete k2
	err = db.DelKV(k2)
	assert.NoError(t, err)

	// list db
	vs, err = db.ListKV([]byte("/"))
	assert.NoError(t, err)
	assert.Len(t, vs, 0)

	// test Chinese
	k1 = []byte("/陈/张")
	k2 = []byte("/陈/王")
	k3 = []byte("/赵/张")

	// put url-like key
	err = db.PutKV(k1, k1)
	assert.NoError(t, err)

	// put url-like key
	err = db.PutKV(k2, k2)
	assert.NoError(t, err)

	// put url-like key
	err = db.PutKV(k3, k3)
	assert.NoError(t, err)

	// list db
	vs, err = db.ListKV([]byte("/陈"))
	assert.NoError(t, err)
	assert.Len(t, vs, 2)
	assert.Equal(t, vs[0].Key, k1)
	assert.Equal(t, vs[0].Value, k1)
	assert.Equal(t, vs[1].Key, k2)
	assert.Equal(t, vs[1].Value, k2)

	vs, err = db.ListKV([]byte("/陈/"))
	assert.NoError(t, err)
	assert.Len(t, vs, 2)
	assert.Equal(t, vs[0].Key, k1)
	assert.Equal(t, vs[0].Value, k1)
	assert.Equal(t, vs[1].Key, k2)
	assert.Equal(t, vs[1].Value, k2)

	vs, err = db.ListKV([]byte("/赵/"))
	assert.NoError(t, err)
	assert.Len(t, vs, 1)
	assert.Equal(t, vs[0].Key, k3)
	assert.Equal(t, vs[0].Value, k3)

	vs, err = db.ListKV([]byte("/李/"))
	assert.NoError(t, err)
	assert.Len(t, vs, 0)

	// delete k1
	err = db.DelKV(k1)
	assert.NoError(t, err)

	// list db
	vs, err = db.ListKV([]byte("/陈"))
	assert.NoError(t, err)
	assert.Len(t, vs, 1)
	assert.Equal(t, vs[0].Key, k2)
	assert.Equal(t, vs[0].Value, k2)

	// delete k3
	err = db.DelKV(k3)
	assert.NoError(t, err)

	// list db
	vs, err = db.ListKV([]byte("/赵"))
	assert.NoError(t, err)
	assert.Len(t, vs, 0)

	// delete k2
	err = db.DelKV(k2)
	assert.NoError(t, err)

	// list db
	vs, err = db.ListKV([]byte("/陈"))
	assert.NoError(t, err)
	assert.Len(t, vs, 0)
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
	vs, err := db.ListKV([]byte("/"))
	assert.NoError(b, err)
	assert.Len(b, vs, 0)

	k1 := []byte("/")
	b.ResetTimer()
	b.Run("put", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := bytes.Join([][]byte{k1, []byte("/"), int32ToBytes(i)}, []byte(""))
			db.PutKV(key, key)
		}
	})
	b.Run("get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := bytes.Join([][]byte{k1, []byte("/"), int32ToBytes(i)}, []byte(""))
			db.GetKV(key)
		}
	})
	b.Run("del", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := bytes.Join([][]byte{k1, []byte("/"), int32ToBytes(i)}, []byte(""))
			db.DelKV(key)
		}
	})
}

func int32ToBytes(i int) []byte {
	var buf = make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(i))
	return buf
}
