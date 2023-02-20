package db

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/db/ent/deletedsubscription"
)

func AddDeletedSubscription(ctx context.Context, tx *ent.Tx, mboxName string, mboxID imap.MailboxID) error {
	count, err := tx.DeletedSubscription.Update().Where(deletedsubscription.NameEqualFold(mboxName)).SetRemoteID(mboxID).Save(ctx)
	if err != nil {
		return err
	}

	if count == 0 {
		if _, err := tx.DeletedSubscription.Create().SetRemoteID(mboxID).SetName(mboxName).Save(ctx); err != nil {
			return err
		}
	}

	return nil
}

func RemoveDeletedSubscriptionWithName(ctx context.Context, tx *ent.Tx, mboxName string) (int, error) {
	return tx.DeletedSubscription.Delete().Where(deletedsubscription.NameEqualFold(mboxName)).Exec(ctx)
}

func GetDeletedSubscriptionSet(ctx context.Context, client *ent.Client) (map[imap.MailboxID]*ent.DeletedSubscription, error) {
	const QueryLimit = 16000

	subscriptions := make(map[imap.MailboxID]*ent.DeletedSubscription)

	for i := 0; ; i += QueryLimit {
		result, err := client.DeletedSubscription.Query().
			Limit(QueryLimit).
			Offset(i).
			All(ctx)
		if err != nil {
			return nil, err
		}

		resultLen := len(result)
		if resultLen == 0 {
			break
		}

		for _, v := range result {
			subscriptions[v.RemoteID] = v
		}
	}

	return subscriptions, nil
}
