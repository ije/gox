package pgx

import (
	"github.com/jackc/pgx"
	"strings"
)

var ErrNoRows = pgx.ErrNoRows

func IsDupError(err error, constraintName string) bool {
	pgErr, ok := err.(pgx.PgError)
	return ok && pgErr.Code == "23505" && (pgErr.ConstraintName == constraintName || strings.Contains(pgErr.ConstraintName, constraintName))
}
