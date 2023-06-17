package pgxerrors

import (
	"errors"

	"github.com/jackc/pgerrcode"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func getCodeUniqueViolation() string {
	return pgerrcode.UniqueViolation
}

func isEqualSQLErrorToCode(sqlErr error, errCode string) (isEqual bool, err error) {
	if sqlErr != nil {
		var pgErr *pgconn.PgError
		if errors.As(sqlErr, &pgErr) {
			isEqual = pgErr.SQLState() == errCode
			return
		}
	}
	return
}

func IsUniqueViolation(sqlErr error) (isOk bool, err error) {
	codeErr := getCodeUniqueViolation()
	return isEqualSQLErrorToCode(sqlErr, codeErr)
}

func IsErrTxCommitRollback(sqlErr error) (isOk bool, err error) {
	return errors.Is(sqlErr, pgx.ErrTxCommitRollback), nil
}
