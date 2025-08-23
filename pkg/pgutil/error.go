package pgutil

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func IsUniqueViolationError(err error) bool {
	var pqErr *pgconn.PgError
	ok := errors.As(err, &pqErr)
	return ok && pqErr.Code == "23505"
}
