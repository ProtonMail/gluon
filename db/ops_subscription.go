package db

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
)

type SubscriptionReadOps interface {
	GetDeletedSubscriptionSet(ctx context.Context) (map[imap.MailboxID]*DeletedSubscription, error)
}

type SubscriptionWriteOps interface {
	AddDeletedSubscription(ctx context.Context, mboxName string, mboxID imap.MailboxID) error
	RemoveDeletedSubscriptionWithName(ctx context.Context, mboxName string) (int, error)
}
