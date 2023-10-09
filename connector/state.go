package connector

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
)

type IMAPState interface {
	Read(ctx context.Context, f func(context.Context, IMAPStateRead) error) error
	Write(ctx context.Context, f func(context.Context, IMAPStateWrite) error) error
}

func IMAPStateReadType[T any](ctx context.Context, state IMAPState, f func(context.Context, IMAPStateRead) (T, error)) (T, error) {
	var result T

	err := state.Read(ctx, func(ctx context.Context, r IMAPStateRead) error {
		t, err := f(ctx, r)
		result = t

		return err
	})

	return result, err
}

func IMAPStateWriteType[T any](ctx context.Context, state IMAPState, f func(context.Context, IMAPStateWrite) (T, error)) (T, error) {
	var result T

	err := state.Write(ctx, func(ctx context.Context, w IMAPStateWrite) error {
		t, err := f(ctx, w)
		result = t

		return err
	})

	return result, err
}

type IMAPStateRead interface {
	GetSettings(ctx context.Context) (string, bool, error)

	GetMailboxCount(ctx context.Context) (int, error)

	GetMailboxesWithoutAttrib(ctx context.Context) ([]imap.MailboxNoAttrib, error)
}

type IMAPStateWrite interface {
	IMAPStateRead

	CreateMailbox(ctx context.Context, mailbox imap.Mailbox) error

	UpdateMessageFlags(ctx context.Context, id imap.MessageID, flags imap.FlagSet) error

	StoreSettings(ctx context.Context, value string) error

	// PatchMailboxHierarchyWithoutTransforms will change the name of the mailbox, but will not perform any of the required
	// transformation necessary to ensure that new parent or child mailboxes are created as expected by a regular
	// IMAP rename operation.
	PatchMailboxHierarchyWithoutTransforms(ctx context.Context, id imap.MailboxID, newName []string) error
}
