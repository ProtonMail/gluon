package db

import (
	"context"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/db/ent/subscription"
	"github.com/ProtonMail/gluon/internal/ids"
)

func AddSubscription(ctx context.Context, tx *ent.Tx, mboxName string, mboxID ids.MailboxIDPair) error {
	count, err := tx.Subscription.Update().Where(subscription.NameEqualFold(mboxName)).SetMailboxID(mboxID.InternalID).SetRemoteID(mboxID.RemoteID).Save(ctx)
	if err != nil {
		return err
	}

	if count == 0 {
		if _, err := tx.Subscription.Create().SetMailboxID(mboxID.InternalID).SetRemoteID(mboxID.RemoteID).SetName(mboxName).Save(ctx); err != nil {
			return err
		}
	}

	return nil
}

func DeleteSubscriptionWithName(ctx context.Context, tx *ent.Tx, mboxName string) (int, error) {
	return tx.Subscription.Delete().Where(subscription.NameEqualFold(mboxName)).Exec(ctx)
}

func DeleteSubscriptionWithMboxID(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID) (int, error) {
	return tx.Subscription.Delete().Where(subscription.MailboxID(mboxID)).Exec(ctx)
}

func GetSubscription(ctx context.Context, client *ent.Client, mboxName string) (*ent.Subscription, error) {
	return client.Subscription.Query().Where(subscription.NameEqualFold(mboxName)).Only(ctx)
}

func GetSubscriptionWithMBoxID(ctx context.Context, client *ent.Client, mboxID imap.InternalMailboxID) (*ent.Subscription, error) {
	return client.Subscription.Query().Where(subscription.MailboxID(mboxID)).Only(ctx)
}

func GetSubscriptionSet(ctx context.Context, client *ent.Client) (map[imap.InternalMailboxID]*ent.Subscription, error) {
	const QueryLimit = 16000

	subscriptions := make(map[imap.InternalMailboxID]*ent.Subscription)

	for i := 0; ; i += QueryLimit {
		result, err := client.Subscription.Query().
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
			subscriptions[v.MailboxID] = v
		}
	}

	return subscriptions, nil
}
