package database

import (
	"database/sql"
	"strings"
)

var placeholderValue = "(?)"
var placeholderKeyValue = "(?,?)"
var schema = map[string][]string{
	"sqlite3": []string{
		`CREATE TABLE IF NOT EXISTS kv (
			key TEXT PRIMARY KEY,
			value TEXT,
			ts TIMESTAMP DEFAULT CURRENT_TIMESTAMP) WITHOUT ROWID`,
	},
}

func init() {
	Factories["sqlite3"] = _new
}

// sqldb the backend SQL DB to persist values
type sqldb struct {
	*sql.DB
	conf Conf
}

// New creates a new sql database
func _new(conf Conf) (DB, error) {
	db, err := sql.Open(conf.Driver, conf.Source)
	if err != nil {
		return nil, err
	}
	for _, v := range schema[conf.Driver] {
		if _, err = db.Exec(v); err != nil {
			db.Close()
			return nil, err
		}
	}
	return &sqldb{DB: db, conf: conf}, nil
}

// Conf returns the configuration
func (d *sqldb) Conf() Conf {
	return d.conf
}

// PutKV put key and value into SQL DB
func (d *sqldb) PutKV(key string, value []byte) error {
	if len(key) == 0 {
		return nil
	}

	stmt, err := d.Prepare("insert into kv(key,value) values (?,?) on conflict(key) do update set value=excluded.value")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(key, value)
	return err
}

// GetKV gets value by key from SQL DB
func (d *sqldb) GetKV(key string) (result KV, err error) {
	rows, err := d.Query("select value from kv where key=?", key)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	if rows.Next() {
		var value []byte
		err = rows.Scan(&value)
		if err != nil {
			return result, err
		}
		result = KV{Key: key, Value: value}
		return result, nil
	}
	return result, nil
}

// Del deletes key and value from SQL DB
func (d *sqldb) DelKV(key string) error {
	_, err := d.Exec("delete from kv where key=?", key)
	return err
}

// ListKV list kvs under the prefix
func (d *sqldb) ListKV(prefix string) (results []KV, err error) {
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	rows, err := d.Query("select key, value from kv where key like ?", prefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		var value []byte
		err = rows.Scan(&key, &value)
		if err != nil {
			return nil, err
		}
		results = append(results, KV{Key: key, Value: value})
	}
	return results, nil
}
