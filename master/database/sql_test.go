package database

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	_, err := New(Conf{Driver: "sqlite2", Source: path.Join("test", "kv.db")})
	assert.Error(t, err)
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

	k1 := "k1"
	k2 := "k2"

	// k1 does not exist
	v, err := db.GetKV(k1)
	assert.NoError(t, err)
	assert.Empty(t, v.Key)
	assert.Nil(t, v.Value)

	// put k1
	err = db.PutKV(k1, []byte(k1))
	assert.NoError(t, err)

	// k1 exists
	v, err = db.GetKV(k1)
	assert.NoError(t, err)
	assert.Equal(t, k1, v.Key)
	assert.Equal(t, []byte(k1), v.Value)

	// put k1 again
	err = db.PutKV(k1, []byte(k2))
	assert.NoError(t, err)

	// k1 exists
	v, err = db.GetKV(k1)
	assert.NoError(t, err)
	assert.Equal(t, k1, v.Key)
	assert.Equal(t, []byte(k2), v.Value)

	// del k1
	err = db.DelKV(k1)
	assert.NoError(t, err)

	// k1 does not exist
	v, err = db.GetKV(k1)
	assert.NoError(t, err)
	assert.Empty(t, v.Key)
	assert.Nil(t, v.Value)

	// key is ""
	k1 = ""
	err = db.PutKV(k1, []byte(k2))
	assert.NoError(t, err)

	// list db
	vs, err := db.ListKV("/")
	assert.NoError(t, err)
	assert.Len(t, vs, 0)

	k1 = "/k/1"
	k2 = "/k/2"
	k3 := "/kk/2"

	// put url-like key
	err = db.PutKV(k1, []byte(k1))
	assert.NoError(t, err)

	// put url-like key
	err = db.PutKV(k2, []byte(k2))
	assert.NoError(t, err)

	// put url-like key
	err = db.PutKV(k3, []byte(k3))
	assert.NoError(t, err)

	// list db
	vs, err = db.ListKV("/")
	assert.NoError(t, err)
	assert.Len(t, vs, 3)
	assert.Equal(t, vs[0].Key, k1)
	assert.Equal(t, vs[0].Value, []byte(k1))
	assert.Equal(t, vs[1].Key, k2)
	assert.Equal(t, vs[1].Value, []byte(k2))
	assert.Equal(t, vs[2].Key, k3)
	assert.Equal(t, vs[2].Value, []byte(k3))

	vs, err = db.ListKV("/k")
	assert.NoError(t, err)
	assert.Len(t, vs, 2)
	assert.Equal(t, vs[0].Key, k1)
	assert.Equal(t, vs[0].Value, []byte(k1))
	assert.Equal(t, vs[1].Key, k2)
	assert.Equal(t, vs[1].Value, []byte(k2))

	vs, err = db.ListKV("/k/")
	assert.NoError(t, err)
	assert.Len(t, vs, 2)
	assert.Equal(t, vs[0].Key, k1)
	assert.Equal(t, vs[0].Value, []byte(k1))
	assert.Equal(t, vs[1].Key, k2)
	assert.Equal(t, vs[1].Value, []byte(k2))

	vs, err = db.ListKV("/kx/")
	assert.NoError(t, err)
	assert.Len(t, vs, 0)

	// delete k1
	err = db.DelKV(k1)
	assert.NoError(t, err)

	// list db
	vs, err = db.ListKV("/")
	assert.NoError(t, err)
	assert.Len(t, vs, 2)
	assert.Equal(t, vs[0].Key, k2)
	assert.Equal(t, vs[0].Value, []byte(k2))
	assert.Equal(t, vs[1].Key, k3)
	assert.Equal(t, vs[1].Value, []byte(k3))

	// delete k3
	err = db.DelKV(k3)
	assert.NoError(t, err)

	// list db
	vs, err = db.ListKV("/kx")
	assert.NoError(t, err)
	assert.Len(t, vs, 0)

	// delete k3
	err = db.DelKV(k2)
	assert.NoError(t, err)

	// list db
	vs, err = db.ListKV("/")
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
	vs, err := db.ListKV("/")
	assert.NoError(b, err)
	assert.Len(b, vs, 0)

	k1 := "/"
	b.ResetTimer()
	b.Run("put", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("%s/%d", k1, i)
			db.PutKV(key, []byte(key))
		}
	})
	b.Run("get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("%s/%d", k1, i)
			db.GetKV(key)
		}
	})
	b.Run("del", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("%s/%d", k1, i)
			db.DelKV(key)
		}
	})
}
