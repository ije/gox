package pg

import (
	"fmt"
	"regexp"
	"strings"
)

func ParseWhere(where map[string]interface{}, filter func(expressions []string, values []interface{}) ([]string, []interface{}), columnFilter ...string) (whereSql string, values []interface{}) {
	if where == nil || len(where) == 0 {
		return
	}

	var expressions []string
	var allowedColumns map[string]struct{}

	if len(columnFilter) > 0 {
		allowedColumns = map[string]struct{}{}
		for _, column := range columnFilter {
			allowedColumns[column] = struct{}{}
		}
	}

	for column, data := range where {
		operator := "="
		column = strings.TrimSpace(column)
		if strings.HasSuffix(column, "%=") {
			operator = " LIKE "
		} else if strings.HasSuffix(column, "!=") || strings.HasSuffix(column, "<>") {
			operator = "!="
		} else if strings.HasSuffix(column, "<") {
			operator = "<"
		} else if strings.HasSuffix(column, "<=") {
			operator = "<="
		} else if strings.HasSuffix(column, ">") {
			operator = ">"
		} else if strings.HasSuffix(column, ">=") {
			operator = ">="
		}

		if column = strings.TrimRight(column, " %!=<>"); len(column) == 0 {
			continue
		}

		if allowedColumns != nil {
			if _, ok := allowedColumns[column]; !ok {
				continue
			}
		}

		if a, ok := data.([]interface{}); ok {
			if l := len(a); l > 0 {
				expressions = append(expressions, Expression(column, l))
				values = append(values, a...)
			}
		} else if a, ok := data.([]int); ok {
			if l := len(a); l > 0 {
				expressions = append(expressions, Expression(column, l))
				for _, i := range a {
					values = append(values, i)
				}
			}
		} else if a, ok := data.([]float64); ok {
			if l := len(a); l > 0 {
				expressions = append(expressions, Expression(column, l))
				for _, i := range a {
					values = append(values, i)
				}
			}
		} else if a, ok := data.([]string); ok {
			if l := len(a); l > 0 {
				expressions = append(expressions, Expression(column, l))
				for _, s := range a {
					values = append(values, s)
				}
			}
		} else {
			expressions = append(expressions, fmt.Sprintf(`"%s"%s?`, column, operator))
			values = append(values, data)
		}
	}

	if filter != nil {
		expressions, values = filter(expressions, values)
	}

	if len(expressions) > 0 {
		whereSql = "WHERE " + strings.Join(expressions, " AND ")
	}
	return
}

func ParseOderBy(orderBy []string, columnFilter ...string) (sql string) {
	if orderBy == nil || len(orderBy) == 0 {
		return
	}

	var allowedColumns map[string]struct{}
	if len(columnFilter) > 0 {
		allowedColumns = map[string]struct{}{}
		for _, column := range columnFilter {
			allowedColumns[column] = struct{}{}
		}
	}

	var orderSqls []string
	for _, column := range orderBy {
		co := "ASC"
		column = strings.TrimPrefix(strings.TrimSpace(column), "+")
		if strings.HasPrefix(column, "-") {
			column = strings.TrimPrefix(column, "-")
			co = "DESC"
		}

		if column = strings.TrimSpace(column); len(column) == 0 {
			continue
		}

		if allowedColumns != nil {
			if _, ok := allowedColumns[column]; !ok {
				continue
			}
		}

		orderSqls = append(orderSqls, fmt.Sprintf(`"%s" %s`, column, co))
	}
	if len(orderSqls) > 0 {
		sql = fmt.Sprintf("ORDER BY %s", strings.Join(orderSqls, ", "))
	}
	return
}

func Expression(column string, values int) string {
	if values == 1 {
		return fmt.Sprintf(`"%s"=?`, column)
	}

	return fmt.Sprintf(`"%s" IN (%s)`, column, QMS(values))
}

func QMS(l int) string {
	qs := make([]string, l)
	for i, _ := range qs {
		qs[i] = "?"
	}
	return strings.Join(qs, ",")
}

func SQLFormat(format string, a ...interface{}) string {
	var i int

	return regexp.MustCompile(`\?`).ReplaceAllStringFunc(fmt.Sprintf(format, a...), func(q string) string {
		i++
		return fmt.Sprintf("$%d", i)
	})
}
