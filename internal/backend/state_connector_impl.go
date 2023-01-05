package backend

import (
	"context"
	"time"

	"github.com/ProtonMail/gluon/connector"
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

func (sc *stateConnectorImpl) SetConnMetadataValue(key string, value any) {
	sc.metadata[key] = value
}

func (sc *stateConnectorImpl) ClearConnMetadataValue(key string) {
	delete(sc.metadata, key)
}

func (sc *stateConnectorImpl) ClearAllConnMetadata() {
	sc.metadata = make(map[string]any)
}

func (sc *stateConnectorImpl) CreateMailbox(ctx context.Context, name []string) (imap.Mailbox, error) {
	ctx = sc.newContextWithMetadata(ctx)

	mbox, err := sc.connector.CreateMailbox(ctx, name)
	if err != nil {
		return imap.Mailbox{}, err
	}

	return mbox, nil
}

func (sc *stateConnectorImpl) UpdateMailbox(ctx context.Context, mboxID imap.MailboxID, newName []string) error {
	ctx = sc.newContextWithMetadata(ctx)

	return sc.connector.UpdateMailboxName(ctx, mboxID, newName)
}

func (sc *stateConnectorImpl) DeleteMailbox(ctx context.Context, mboxID imap.MailboxID) error {
	ctx = sc.newContextWithMetadata(ctx)

	return sc.connector.DeleteMailbox(ctx, mboxID)
}

func (sc *stateConnectorImpl) CreateMessage(
	ctx context.Context,
	mboxID imap.MailboxID,
	literal []byte,
	flags imap.FlagSet,
	date time.Time,
) (imap.InternalMessageID, imap.Message, []byte, error) {
	ctx = sc.newContextWithMetadata(ctx)

	msg, newLiteral, err := sc.connector.CreateMessage(ctx, mboxID, literal, flags, date)
	if err != nil {
		return imap.InternalMessageID{}, imap.Message{}, nil, err
	}

	return imap.NewInternalMessageID(), msg, newLiteral, nil
}

func (sc *stateConnectorImpl) AddMessagesToMailbox(
	ctx context.Context,
	messageIDs []imap.MessageID,
	mboxID imap.MailboxID,
) error {
	ctx = sc.newContextWithMetadata(ctx)

	return sc.connector.AddMessagesToMailbox(ctx, messageIDs, mboxID)
}

func (sc *stateConnectorImpl) RemoveMessagesFromMailbox(
	ctx context.Context,
	messageIDs []imap.MessageID,
	mboxID imap.MailboxID,
) error {
	ctx = sc.newContextWithMetadata(ctx)

	return sc.connector.RemoveMessagesFromMailbox(ctx, messageIDs, mboxID)
}

func (sc *stateConnectorImpl) MoveMessagesFromMailbox(
	ctx context.Context,
	messageIDs []imap.MessageID,
	mboxFromID imap.MailboxID,
	mboxToID imap.MailboxID,
) (bool, error) {
	ctx = sc.newContextWithMetadata(ctx)

	return sc.connector.MoveMessages(ctx, messageIDs, mboxFromID, mboxToID)
}

func (sc *stateConnectorImpl) SetMessagesSeen(ctx context.Context, messageIDs []imap.MessageID, seen bool) error {
	ctx = sc.newContextWithMetadata(ctx)

	return sc.connector.MarkMessagesSeen(ctx, messageIDs, seen)
}

func (sc *stateConnectorImpl) SetMessagesFlagged(ctx context.Context, messageIDs []imap.MessageID, flagged bool) error {
	ctx = sc.newContextWithMetadata(ctx)

	return sc.connector.MarkMessagesFlagged(ctx, messageIDs, flagged)
}

func (sc *stateConnectorImpl) SetUIDValidity(uidValidity imap.UID) error {
	return sc.connector.SetUIDValidity(uidValidity)
}

func (sc *stateConnectorImpl) IsMailboxVisible(ctx context.Context, id imap.MailboxID) bool {
	return sc.connector.IsMailboxVisible(ctx, id)
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
