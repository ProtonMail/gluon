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
	GetMailboxCount(ctx context.Context) (int, error)
}

type IMAPStateWrite interface {
	IMAPStateRead

	CreateMailbox(ctx context.Context, mailbox imap.Mailbox) error

	UpdateMessageFlags(ctx context.Context, id imap.MessageID, flags imap.FlagSet) error
}
