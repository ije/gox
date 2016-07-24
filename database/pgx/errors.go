package pgx

import (
	"github.com/jackc/pgx"
	"strings"
)

var ErrNoRows = pgx.ErrNoRows

func IsDupError(err error, cloumn string) bool {
	pgErr, ok := err.(pgx.PgError)
	return ok && pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, cloumn)
}
