package pgx

import (
	"time"

	"github.com/jackc/pgx"
)

type Schema struct {
	Name string
	*ConnPool
}

func (schema *Schema) String() string {
	if len(schema.Name) == 0 {
		return "public"
	}
	return schema.Name
}

func (schema *Schema) Count(table string, where map[string]interface{}, whereFilter func(expressions []string, values []interface{}) ([]string, []interface{}), columnFilter ...string) (count int, err error) {
	if len(table) == 0 {
		return
	}

	whereSql, values := ParseWhere(where, whereFilter, columnFilter...)

	var rows int32
	err = schema.QueryRow(SQLFormat(`SELECT COUNT(*) AS "rows" FROM "%s"."%s" %s`, schema.String(), table, whereSql), values...).Scan(&rows)
	if err == pgx.ErrNoRows {
		err = nil
	}

	count = int(rows)
	return
}

func (schema *Schema) Delete(table string, where map[string]interface{}, whereFilter func(expressions []string, values []interface{}) ([]string, []interface{}), columnFilter ...string) (affected int, err error) {
	if len(table) == 0 || where == nil || len(where) == 0 {
		return
	}

	whereSql, values := ParseWhere(where, whereFilter, columnFilter...)
	if len(whereSql) > 0 {
		var ret Result
		ret, err = schema.Exec(SQLFormat(`DELETE FROM "%s"."%s" %s`, schema.String(), table, whereSql), values...)
		if err != nil {
			return
		}
		affected = int(ret.RowsAffected())
	}

	return
}

func (schema *Schema) NewUpdateTask(table string, where map[string]interface{}, updateDelay time.Duration) *DBUpdateTask {
	return &DBUpdateTask{
		Table:       table,
		Where:       where,
		UpdateDelay: updateDelay,
		Schema:      schema,
	}
}

func (schema *Schema) Backup() (sql []byte, err error) {
	// SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname='public'
	// SELECT column_name FROM information_schema.columns WHERE table_schema='public' AND table_name ='posts'
	return
}
