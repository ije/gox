package pgx

import (
	"github.com/jackc/pgx"
)

type Tx struct {
	*pgx.Tx
}

func (tx *Tx) Exec(sql string, args ...interface{}) (ret Result, err error) {
	commandTag, err := tx.Tx.Exec(sql, args...)
	if err != nil {
		return
	}
	ret = Result(commandTag)
	return
}

func (tx *Tx) Query(sql string, args ...interface{}) (rows *Rows, err error) {
	rs, err := tx.Tx.Query(sql, args...)
	if err != nil {
		return
	}

	rows = &Rows{rs}
	return
}

func (tx *Tx) QueryRow(sql string, args ...interface{}) *Row {
	return &Row{tx.Tx.QueryRow(sql, args...)}
}
