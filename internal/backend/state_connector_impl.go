package backend

import (
	"context"
	"time"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/state"
)

type stateConnectorImpl struct {
	connector connector.Connector
	metadata  map[string]any
	user      *user
}

func newStateConnectorImpl(u *user) state.Connector {
	return &stateConnectorImpl{
		connector: u.connector,
		metadata:  make(map[string]any),
		user:      u,
	}
}

func (sc *stateConnectorImpl) newDBIMAPWrite(tx db.Transaction) DBIMAPStateWrite {
	return DBIMAPStateWrite{
		DBIMAPStateRead: DBIMAPStateRead{rd: tx},
		tx:              tx,
		user:            sc.user,
	}
}

func (sc *stateConnectorImpl) SetConnMetadataValue(key string, value any) {
	sc.metadata[key] = value
}

func (sc *stateConnectorImpl) ClearConnMetadataValue(key string) {
	delete(sc.metadata, key)
}

func (sc *stateConnectorImpl) ClearAllConnMetadata() {
	sc.metadata = make(map[string]any)
}

func (sc *stateConnectorImpl) CreateMailbox(ctx context.Context, tx db.Transaction, name []string) ([]state.Update, imap.Mailbox, error) {
	ctx = sc.newContextWithMetadata(ctx)

	cache := sc.newDBIMAPWrite(tx)

	mbox, err := sc.connector.CreateMailbox(ctx, &cache, name)
	if err != nil {
		return nil, imap.Mailbox{}, err
	}

	return cache.stateUpdates, mbox, nil
}

func (sc *stateConnectorImpl) UpdateMailbox(ctx context.Context, tx db.Transaction, mboxID imap.MailboxID, newName []string) ([]state.Update, error) {
	ctx = sc.newContextWithMetadata(ctx)

	cache := sc.newDBIMAPWrite(tx)

	if err := sc.connector.UpdateMailboxName(ctx, &cache, mboxID, newName); err != nil {
		return nil, err
	}

	return cache.stateUpdates, nil
}

func (sc *stateConnectorImpl) DeleteMailbox(ctx context.Context, tx db.Transaction, mboxID imap.MailboxID) ([]state.Update, error) {
	ctx = sc.newContextWithMetadata(ctx)

	cache := sc.newDBIMAPWrite(tx)

	if err := sc.connector.DeleteMailbox(ctx, &cache, mboxID); err != nil {
		return nil, err
	}

	return cache.stateUpdates, nil
}

func (sc *stateConnectorImpl) CreateMessage(
	ctx context.Context,
	tx db.Transaction,
	mboxID imap.MailboxID,
	literal []byte,
	flags imap.FlagSet,
	date time.Time,
) ([]state.Update, imap.InternalMessageID, imap.Message, []byte, error) {
	ctx = sc.newContextWithMetadata(ctx)

	cache := sc.newDBIMAPWrite(tx)

	msg, newLiteral, err := sc.connector.CreateMessage(ctx, &cache, mboxID, literal, flags, date)
	if err != nil {
		return nil, imap.InternalMessageID{}, imap.Message{}, nil, err
	}

	return cache.stateUpdates, imap.NewInternalMessageID(), msg, newLiteral, nil
}

func (sc *stateConnectorImpl) GetMessageLiteral(ctx context.Context, id imap.MessageID) ([]byte, error) {
	ctx = sc.newContextWithMetadata(ctx)

	return sc.connector.GetMessageLiteral(ctx, id)
}

func (sc *stateConnectorImpl) AddMessagesToMailbox(
	ctx context.Context,
	tx db.Transaction,
	messageIDs []imap.MessageID,
	mboxID imap.MailboxID,
) ([]state.Update, error) {
	ctx = sc.newContextWithMetadata(ctx)

	cache := sc.newDBIMAPWrite(tx)

	if err := sc.connector.AddMessagesToMailbox(ctx, &cache, messageIDs, mboxID); err != nil {
		return nil, err
	}

	return cache.stateUpdates, nil
}

func (sc *stateConnectorImpl) RemoveMessagesFromMailbox(
	ctx context.Context,
	tx db.Transaction,
	messageIDs []imap.MessageID,
	mboxID imap.MailboxID,
) ([]state.Update, error) {
	ctx = sc.newContextWithMetadata(ctx)

	cache := sc.newDBIMAPWrite(tx)

	if err := sc.connector.RemoveMessagesFromMailbox(ctx, &cache, messageIDs, mboxID); err != nil {
		return nil, err
	}

	return cache.stateUpdates, nil
}

func (sc *stateConnectorImpl) MoveMessagesFromMailbox(
	ctx context.Context,
	tx db.Transaction,
	messageIDs []imap.MessageID,
	mboxFromID imap.MailboxID,
	mboxToID imap.MailboxID,
) ([]state.Update, bool, error) {
	ctx = sc.newContextWithMetadata(ctx)

	cache := sc.newDBIMAPWrite(tx)

	shouldMove, err := sc.connector.MoveMessages(ctx, &cache, messageIDs, mboxFromID, mboxToID)
	if err != nil {
		return nil, false, err
	}

	return cache.stateUpdates, shouldMove, nil
}

func (sc *stateConnectorImpl) SetMessagesSeen(ctx context.Context, tx db.Transaction, messageIDs []imap.MessageID, seen bool) ([]state.Update, error) {
	ctx = sc.newContextWithMetadata(ctx)

	cache := sc.newDBIMAPWrite(tx)

	if err := sc.connector.MarkMessagesSeen(ctx, &cache, messageIDs, seen); err != nil {
		return nil, err
	}

	return cache.stateUpdates, nil
}

func (sc *stateConnectorImpl) SetMessagesFlagged(ctx context.Context,
	tx db.Transaction,
	messageIDs []imap.MessageID, flagged bool) ([]state.Update, error) {
	ctx = sc.newContextWithMetadata(ctx)

	cache := sc.newDBIMAPWrite(tx)

	if err := sc.connector.MarkMessagesFlagged(ctx, &cache, messageIDs, flagged); err != nil {
		return nil, err
	}

	return cache.stateUpdates, nil
}

func (sc *stateConnectorImpl) GetMailboxVisibility(ctx context.Context,
	id imap.MailboxID) imap.MailboxVisibility {
	return sc.connector.GetMailboxVisibility(ctx, id)
}

func (sc *stateConnectorImpl) getMetadataValue(key string) any {
	v, ok := sc.metadata[key]
	if !ok {
		return nil
	}

	return v
}

func (sc *stateConnectorImpl) newContextWithMetadata(ctx context.Context) context.Context {
	if v := sc.getMetadataValue(imap.IMAPIDConnMetadataKey); v != nil {
		switch x := v.(type) {
		case imap.IMAPID:
			ctx = imap.NewContextWithIMAPID(ctx, x)
		}
	}

	return ctx
}
