package pg

import (
	"github.com/jackc/pgx"
)

type ConnPoolConfig struct {
	DefaultScheme  string
	Host           string // host (e.g. localhost) or path to unix domain socket directory (e.g. /private/tmp)
	Port           uint16 // default: 5432
	Database       string
	User           string // default: OS user name
	Password       string
	Logger         pgx.Logger
	LogLevel       int
	MaxConnections int                   // max simultaneous connections to use, default 5, must be at least 2
	AfterConnect   func(*pgx.Conn) error // function to call on every new connection
}

type ConnPool struct {
	DefaultScheme string
	*pgx.ConnPool
}

func NewConnPool(config ConnPoolConfig) (pool *ConnPool, err error) {
	xPool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     config.Host,
			Port:     config.Port,
			Database: config.Database,
			User:     config.User,
			Password: config.Password,
			Logger:   config.Logger,
			LogLevel: config.LogLevel,
		},
		MaxConnections: config.MaxConnections,
		AfterConnect:   config.AfterConnect,
	})
	if err != nil {
		return
	}

	pool = &ConnPool{config.DefaultScheme, xPool}
	return
}

func (pool *ConnPool) Scheme() string {
	if len(pool.DefaultScheme) == 0 {
		return "public"
	}
	return pool.DefaultScheme
}

func (pool *ConnPool) Count(table string, where map[string]interface{}, whereFilter func(expressions []string, values []interface{}) ([]string, []interface{}), columnFilter ...string) (count int, err error) {
	if len(table) == 0 {
		return
	}

	whereSql, values := ParseWhere(where, whereFilter, columnFilter...)

	var rows int32
	err = pool.QueryRow(SQLFormat(`SELECT COUNT(*) AS "rows" FROM "%s"."%s" %s`, pool.Scheme(), table, whereSql), values...).Scan(&rows)
	if err == pgx.ErrNoRows {
		err = nil
	}

	count = int(rows)
	return
}

func (pool *ConnPool) Delete(table string, where map[string]interface{}, whereFilter func(expressions []string, values []interface{}) ([]string, []interface{}), columnFilter ...string) (affected int, err error) {
	if len(table) == 0 || where == nil || len(where) == 0 {
		return
	}

	whereSql, values := ParseWhere(where, whereFilter, columnFilter...)
	if len(whereSql) > 0 {
		var ret Result
		ret, err = pool.Exec(SQLFormat(`DELETE FROM "%s"."%s" %s`, pool.Scheme(), table, whereSql), values...)
		if err != nil {
			return
		}
		affected = int(ret.RowsAffected())
	}

	return
}

func (pool *ConnPool) Begin() (tx *Tx, err error) {
	xTx, err := pool.ConnPool.Begin()
	if err != nil {
		return
	}
	tx = &Tx{xTx}
	return
}

func (pool *ConnPool) Exec(sql string, args ...interface{}) (ret Result, err error) {
	commandTag, err := pool.ConnPool.Exec(sql, args...)
	if err != nil {
		return
	}
	ret = Result(commandTag)
	return
}

func (pool *ConnPool) Query(sql string, args ...interface{}) (rows *Rows, err error) {
	xRows, err := pool.ConnPool.Query(sql, args...)
	if err != nil {
		return
	}
	rows = &Rows{xRows}
	return
}

func (pool *ConnPool) QueryRow(sql string, args ...interface{}) *Row {
	return &Row{pool.ConnPool.QueryRow(sql, args...)}
}
