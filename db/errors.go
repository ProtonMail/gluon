package db

import (
	"errors"
	"fmt"

	"github.com/ProtonMail/gluon/internal/hash"
)

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

type UniqueLabelConstraintError struct {
	Field         string
	Existing      *Mailbox
	AttemptedID   string
	AttemptedName string
	OriginalErr   error
}

func (e *UniqueLabelConstraintError) Error() string {
	if e.Existing != nil {
		return fmt.Sprintf(
			"sql: unique constraint on %s: existing mailbox with remote ID %s and name %q, attempted remote ID %s and name %q",
			e.Field, e.Existing.RemoteID, hash.HashToString(e.Existing.Name), e.AttemptedID, hash.HashToString(e.AttemptedName),
		)
	}

	return fmt.Sprintf("sql: unique constraint on %s, attempted remote ID %s and name %q, error: %v",
		e.Field,
		e.AttemptedID,
		hash.HashToString(e.AttemptedName),
		e.OriginalErr)
}

func IsUniqueLabelConstraintError(err error) bool {
	if err == nil {
		return false
	}

	var target *UniqueLabelConstraintError
	return errors.As(err, &target)
}
