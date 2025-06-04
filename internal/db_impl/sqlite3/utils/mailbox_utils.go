package utils

import (
	"context"
	"errors"
	"strings"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/mattn/go-sqlite3"
)

type dbFetcher interface {
	GetMailboxByRemoteID(ctx context.Context, mboxID imap.MailboxID) (*db.Mailbox, error)
	GetMailboxByName(ctx context.Context, name string) (*db.Mailbox, error)
}

func appendFieldToTable(tableName, field string) string {
	return tableName + "." + field
}

func handleUniqueConstraintConflict(
	ctx context.Context,
	fetch func(context.Context) (*db.Mailbox, error),
	field string,
	remoteID imap.MailboxID, name string,
) error {
	uniqueLabelConstraintError := &db.UniqueLabelConstraintError{
		Field:         field,
		AttemptedID:   string(remoteID),
		AttemptedName: name,
	}

	mbox, err := fetch(ctx)
	if err != nil {
		uniqueLabelConstraintError.OriginalErr = err
		return uniqueLabelConstraintError
	}

	uniqueLabelConstraintError.Existing = mbox
	return uniqueLabelConstraintError
}

func MapLabelsUniqueConstraintError(
	ctx context.Context,
	dbFetcher dbFetcher,
	tableName, remoteIDField, nameField string,
	remoteID imap.MailboxID, name string,
	err error,
) error {
	var sqliteErr sqlite3.Error
	if !errors.As(err, &sqliteErr) || sqliteErr.ExtendedCode != sqlite3.ErrConstraintUnique {
		return err
	}

	msg := sqliteErr.Error()
	tableRemoteIDField := appendFieldToTable(tableName, remoteIDField)
	tableNameField := appendFieldToTable(tableName, nameField)

	switch {
	case strings.Contains(msg, tableRemoteIDField):
		return handleUniqueConstraintConflict(ctx,
			func(ctx context.Context) (*db.Mailbox, error) {
				return dbFetcher.GetMailboxByRemoteID(ctx, remoteID)
			},
			tableRemoteIDField, remoteID, name)

	case strings.Contains(msg, tableNameField):
		return handleUniqueConstraintConflict(ctx,
			func(ctx context.Context) (*db.Mailbox, error) {
				return dbFetcher.GetMailboxByName(ctx, name)
			},
			tableNameField, remoteID, name)
	default:
		return err
	}
}
