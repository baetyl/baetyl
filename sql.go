package baetyl

import (
	"database/sql"

	schema "github.com/baetyl/baetyl/schema/v3"
)

var placeholderValue = "(?)"
var placeholderKeyValue = "(?,?)"
var dbSchema = map[string][]string{
	"sqlite3": []string{
		`CREATE TABLE IF NOT EXISTS kv (
			key TEXT PRIMARY KEY,
			value BLOB,
			ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP) WITHOUT ROWID`,
	},
}

func init() {
	RegisterDatabase("sqlite3", _new)
}

// sqldb the backend SQL Database to persist values
type sqldb struct {
	*sql.DB
	ctx Context
}

// New creates a new sql database
func _new(ctx Context, dbName string) (Database, error) {
	db, err := sql.Open(ctx.Config().Database.Driver, dbName)
	if err != nil {
		return nil, err
	}
	for _, v := range dbSchema[ctx.Config().Database.Driver] {
		if _, err = db.Exec(v); err != nil {
			db.Close()
			return nil, err
		}
	}
	return &sqldb{DB: db, ctx: ctx}, nil
}

// Conf returns the context
func (d *sqldb) Context() Context {
	return d.ctx
}

// Set put key and value into SQL Database
func (d *sqldb) Set(kv *schema.KV) error {
	stmt, err := d.Prepare("insert into kv(key,value) values (?,?) on conflict(key) do update set value=excluded.value")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(kv.Key, kv.Value)
	return err
}

// Get gets value by key from SQL Database
func (d *sqldb) Get(key []byte) (*schema.KV, error) {
	rows, err := d.Query("select value from kv where key=?", key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	kv := &schema.KV{Key: key}
	if rows.Next() {
		err = rows.Scan(&kv.Value)
		if err != nil {
			return nil, err
		}
		return kv, nil
	}
	return kv, nil
}

// Del deletes key and value from SQL Database
func (d *sqldb) Del(key []byte) error {
	_, err := d.Exec("delete from kv where key=?", key)
	return err
}

// List list kvs with the prefix
func (d *sqldb) List(prefix []byte) (*schema.KVs, error) {
	rows, err := d.Query("select key, value from kv where key like ?", append(prefix, byte('%')))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	kvs := &schema.KVs{}
	for rows.Next() {
		kv := new(schema.KV)
		err = rows.Scan(&kv.Key, &kv.Value)
		if err != nil {
			return nil, err
		}
		kvs.Kvs = append(kvs.Kvs, kv)
	}
	return kvs, nil
}
