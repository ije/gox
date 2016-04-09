package pgx

import (
	"time"

	"github.com/jackc/pgx"
)

type Instance struct {
	DefaultScheme string
	*ConnPool
}

func (instance *Instance) Scheme() string {
	if len(instance.DefaultScheme) == 0 {
		return "public"
	}
	return instance.DefaultScheme
}

func (instance *Instance) Count(table string, where map[string]interface{}, whereFilter func(expressions []string, values []interface{}) ([]string, []interface{}), columnFilter ...string) (count int, err error) {
	if len(table) == 0 {
		return
	}

	whereSql, values := ParseWhere(where, whereFilter, columnFilter...)

	var rows int32
	err = instance.QueryRow(SQLFormat(`SELECT COUNT(*) AS "rows" FROM "%s"."%s" %s`, instance.Scheme(), table, whereSql), values...).Scan(&rows)
	if err == pgx.ErrNoRows {
		err = nil
	}

	count = int(rows)
	return
}

func (instance *Instance) Delete(table string, where map[string]interface{}, whereFilter func(expressions []string, values []interface{}) ([]string, []interface{}), columnFilter ...string) (affected int, err error) {
	if len(table) == 0 || where == nil || len(where) == 0 {
		return
	}

	whereSql, values := ParseWhere(where, whereFilter, columnFilter...)
	if len(whereSql) > 0 {
		var ret Result
		ret, err = instance.Exec(SQLFormat(`DELETE FROM "%s"."%s" %s`, instance.Scheme(), table, whereSql), values...)
		if err != nil {
			return
		}
		affected = int(ret.RowsAffected())
	}

	return
}

func (instance *Instance) NewUpdateTask(table string, where map[string]interface{}, updateDelay time.Duration) *DBUpdateTask {
	return &DBUpdateTask{
		Table:       table,
		Where:       where,
		UpdateDelay: updateDelay,
		Instance:    instance,
	}
}
