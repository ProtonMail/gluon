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

	mbox, err := sc.connector.CreateLabel(ctx, name)
	if err != nil {
		return imap.Mailbox{}, err
	}

	return mbox, nil
}

func (sc *stateConnectorImpl) UpdateMailbox(ctx context.Context, mboxID imap.LabelID, oldName, newName []string) error {
	ctx = sc.newContextWithMetadata(ctx)

	return sc.connector.UpdateLabel(ctx, mboxID, newName)
}

func (sc *stateConnectorImpl) DeleteMailbox(ctx context.Context, mboxID imap.LabelID) error {
	ctx = sc.newContextWithMetadata(ctx)

	return sc.connector.DeleteLabel(ctx, mboxID)
}

func (sc *stateConnectorImpl) CreateMessage(
	ctx context.Context,
	mboxID imap.LabelID,
	literal []byte,
	parsedMessage *imap.ParsedMessage,
	flags imap.FlagSet,
	date time.Time,
) (imap.InternalMessageID, imap.Message, error) {
	ctx = sc.newContextWithMetadata(ctx)

	msg, err := sc.connector.CreateMessage(ctx, mboxID, literal, parsedMessage, flags, date)
	if err != nil {
		return 0, imap.Message{}, err
	}

	return sc.user.nextMessageID(), msg, nil
}

func (sc *stateConnectorImpl) AddMessagesToMailbox(
	ctx context.Context,
	messageIDs []imap.MessageID,
	mboxID imap.LabelID,
) error {
	ctx = sc.newContextWithMetadata(ctx)

	if err := sc.connector.LabelMessages(ctx, messageIDs, mboxID); err != nil {
		return sc.refresh(ctx, messageIDs, mboxID)
	}

	return nil
}

func (sc *stateConnectorImpl) RemoveMessagesFromMailbox(
	ctx context.Context,
	messageIDs []imap.MessageID,
	mboxID imap.LabelID,
) error {
	ctx = sc.newContextWithMetadata(ctx)

	if err := sc.connector.UnlabelMessages(ctx, messageIDs, mboxID); err != nil {
		return sc.refresh(ctx, messageIDs, mboxID)
	}

	return nil
}

func (sc *stateConnectorImpl) MoveMessagesFromMailbox(
	ctx context.Context,
	messageIDs []imap.MessageID,
	mboxFromID imap.LabelID,
	mboxToID imap.LabelID,
) error {
	ctx = sc.newContextWithMetadata(ctx)

	if err := sc.connector.MoveMessages(ctx, messageIDs, mboxFromID, mboxToID); err != nil {
		return sc.refresh(ctx, messageIDs, mboxFromID)
	}

	return nil
}

func (sc *stateConnectorImpl) SetMessagesSeen(ctx context.Context, messageIDs []imap.MessageID, seen bool) error {
	ctx = sc.newContextWithMetadata(ctx)

	if err := sc.connector.MarkMessagesSeen(ctx, messageIDs, seen); err != nil {
		return sc.refresh(ctx, messageIDs)
	}

	return nil
}

func (sc *stateConnectorImpl) SetMessagesFlagged(ctx context.Context, messageIDs []imap.MessageID, flagged bool) error {
	ctx = sc.newContextWithMetadata(ctx)

	if err := sc.connector.MarkMessagesFlagged(ctx, messageIDs, flagged); err != nil {
		return sc.refresh(ctx, messageIDs)
	}

	return nil
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

func (sc *stateConnectorImpl) refresh(ctx context.Context, messageIDs []imap.MessageID, mboxIDs ...imap.LabelID) error {
	for _, messageID := range messageIDs {
		message, mboxIDs, err := sc.connector.GetMessage(ctx, messageID)
		if err != nil {
			return err
		}

		sc.user.updateInjector.send(imap.NewMessageLabelsUpdated(
			message.ID,
			mboxIDs,
			message.Flags.ContainsUnchecked(imap.FlagSeenLowerCase),
			message.Flags.ContainsUnchecked(imap.FlagFlaggedLowerCase),
		), true)
	}

	for _, mboxID := range mboxIDs {
		mailbox, err := sc.connector.GetLabel(ctx, mboxID)
		if err != nil {
			return err
		}

		sc.user.updateInjector.send(imap.NewMailboxUpdated(
			mailbox.ID,
			mailbox.Name,
		), true)
	}

	return nil
}
