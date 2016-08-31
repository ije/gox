package pgx

import (
	"github.com/jackc/pgx"
	"strings"
)

var ErrNoRows = pgx.ErrNoRows

func IsDBError(err error) bool {
	_, ok := err.(pgx.PgError)
	return ok
}

func IsDupError(err error, constraintName string) bool {
	pgErr, ok := err.(pgx.PgError)
	return ok && pgErr.Code == "23505" && (pgErr.ConstraintName == constraintName || strings.Contains(pgErr.ConstraintName, constraintName))
}
