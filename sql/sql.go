// sql package with log
package sql

import (
	"database/sql"
	"errors"

	"github.com/ije/aisling/log"
)

type LogLevel byte

const (
	LL_OFF LogLevel = iota
	LL_ERROR
	LL_DEBUG
)

type DB struct {
	logLevel LogLevel
	logger   *log.Logger
	*sql.DB
}

func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{logLevel: LL_OFF, DB: db}, nil
}

func (db *DB) SetLog(logger *log.Logger, level LogLevel) {
	db.logLevel = level
	db.logger = logger
}

func (db *DB) log(err error, query string, values ...interface{}) {
	if db.logLevel == LL_OFF || db.logger == nil {
		return
	}
	if err != nil {
		db.logger.Errorf("%v >>> %s [%v]", err, query, values)
	} else if db.logLevel >= LL_DEBUG {
		db.logger.Debugf("%s [%v]", query, values)
	}
}

func (db *DB) Begin() (tx *Tx, err error) {
	otx, err := db.DB.Begin()
	if err == nil {
		tx = &Tx{db, otx}
	}
	go db.log(err, "BEGIN")
	return
}

func (db *DB) Exec(query string, v ...interface{}) (ret sql.Result, err error) {
	ret, err = db.DB.Exec(query, v...)
	go db.log(err, query, v...)
	return
}

func (db *DB) Prepare(query string, v ...interface{}) (stmt *sql.Stmt, err error) {
	stmt, err = db.DB.Prepare(query)
	go db.log(err, query, v...)
	return
}

func (db *DB) Query(query string, v ...interface{}) (rows *sql.Rows, err error) {
	rows, err = db.DB.Query(query, v...)
	go db.log(err, query, v...)
	return
}

func (db *DB) QueryRow(query string, v ...interface{}) (row *Row) {
	rows, err := db.DB.Query(query, v...)
	go db.log(err, query, v...)
	return &Row{rows, err}
}

type Row struct {
	rows *sql.Rows
	err  error // deferred error for easy chaining
}

func (r *Row) Scan(dest ...interface{}) (err error) {
	if r.err != nil {
		return r.err
	}

	for _, dp := range dest {
		if _, ok := dp.(*sql.RawBytes); ok {
			return errors.New("sql: RawBytes isn't allowed on Row.Scan")
		}
	}

	if !r.rows.Next() {
		if err = r.rows.Err(); err != nil {
			return
		}
		return sql.ErrNoRows
	}

	if err = r.rows.Scan(dest...); err != nil {
		return
	}

	// Make sure the query can be processed to completion with no errors.
	if err = r.rows.Close(); err != nil {
		return
	}

	return
}

type Tx struct {
	db *DB
	*sql.Tx
}

func (tx *Tx) Exec(query string, v ...interface{}) (ret sql.Result, err error) {
	ret, err = tx.Tx.Exec(query, v...)
	go tx.db.log(err, query, v...)
	return
}

func (tx *Tx) Prepare(query string, v ...interface{}) (stmt *sql.Stmt, err error) {
	stmt, err = tx.Tx.Prepare(query)
	go tx.db.log(err, query, v...)
	return
}

func (tx *Tx) Query(query string, v ...interface{}) (rows *sql.Rows, err error) {
	rows, err = tx.Tx.Query(query, v...)
	go tx.db.log(err, query, v...)
	return
}

func (tx *Tx) QueryRow(query string, v ...interface{}) (row *Row) {
	rows, err := tx.Tx.Query(query, v...)
	go tx.db.log(err, query, v...)
	return &Row{rows, err}
}

func (tx *Tx) Rollback() (err error) {
	err = tx.Tx.Rollback()
	go tx.db.log(err, "ROLLBACK")
	return
}

func (tx *Tx) Commit() (err error) {
	err = tx.Tx.Commit()
	go tx.db.log(err, "COMMIT")
	return
}
