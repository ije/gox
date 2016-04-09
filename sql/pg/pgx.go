package pg

import (
	"strconv"
	"strings"

	"github.com/jackc/pgx"
)

type Result string

func (ret Result) RowsAffected() int64 {
	s := string(ret)
	index := strings.LastIndex(s, " ")
	if index == -1 {
		return 0
	}
	n, _ := strconv.ParseInt(s[index+1:], 10, 64)
	return n
}

type Row struct {
	*pgx.Row
}

type Rows struct {
	*pgx.Rows
}
