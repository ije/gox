package pgx

import (
	"github.com/jackc/pgx"
)

type ConnPoolConfig struct {
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

	pool = &ConnPool{xPool}
	return
}

func (pool *ConnPool) Scheme(name string) *Scheme {
	return &Scheme{Name: name, ConnPool: pool}
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
