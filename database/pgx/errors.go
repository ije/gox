package pgx

import (
	"github.com/jackc/pgx"
)

var ErrNoRows = pgx.ErrNoRows

func IsDupError(err error) bool {
	pgErr, ok := err.(pgx.PgError)
	return ok && pgErr.Code == "23505"
}
