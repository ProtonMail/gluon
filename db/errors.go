package db

import "errors"

var ErrNotFound = errors.New("value not found")
var ErrTransactionFailed = errors.New("transaction failed")
var ErrMigrationFailed = errors.New("database migration failed")
var ErrInvalidDatabaseVersion = errors.New("invalid database version")

func IsErrNotFound(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, ErrNotFound)
}
