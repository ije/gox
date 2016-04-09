package pgx

import (
	"github.com/jackc/pgx"
)

var ErrNoRows = pgx.ErrNoRows

func IsDupError(err error, table string, column string) bool {
	pgErr, ok := err.(pgx.PgError)
	return ok && pgErr.Code == "23505" && pgErr.TableName == table && pgErr.ColumnName == column
}
