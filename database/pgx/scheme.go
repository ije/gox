package pgx

import (
	"time"

	"github.com/jackc/pgx"
)

type Scheme struct {
	Name string
	*ConnPool
}

func (scheme *Scheme) String() string {
	if len(scheme.Name) == 0 {
		return "public"
	}
	return scheme.Name
}

func (scheme *Scheme) Count(table string, where map[string]interface{}, whereFilter func(expressions []string, values []interface{}) ([]string, []interface{}), columnFilter ...string) (count int, err error) {
	if len(table) == 0 {
		return
	}

	whereSql, values := ParseWhere(where, whereFilter, columnFilter...)

	var rows int32
	err = scheme.QueryRow(SQLFormat(`SELECT COUNT(*) AS "rows" FROM "%s"."%s" %s`, scheme.String(), table, whereSql), values...).Scan(&rows)
	if err == pgx.ErrNoRows {
		err = nil
	}

	count = int(rows)
	return
}

func (scheme *Scheme) Delete(table string, where map[string]interface{}, whereFilter func(expressions []string, values []interface{}) ([]string, []interface{}), columnFilter ...string) (affected int, err error) {
	if len(table) == 0 || where == nil || len(where) == 0 {
		return
	}

	whereSql, values := ParseWhere(where, whereFilter, columnFilter...)
	if len(whereSql) > 0 {
		var ret Result
		ret, err = scheme.Exec(SQLFormat(`DELETE FROM "%s"."%s" %s`, scheme.String(), table, whereSql), values...)
		if err != nil {
			return
		}
		affected = int(ret.RowsAffected())
	}

	return
}

func (scheme *Scheme) NewUpdateTask(table string, where map[string]interface{}, updateDelay time.Duration) *DBUpdateTask {
	return &DBUpdateTask{
		Table:       table,
		Where:       where,
		UpdateDelay: updateDelay,
		Scheme:      scheme,
	}
}
