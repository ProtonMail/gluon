package state

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/bradenaw/juniper/xslices"
)

// Shared code for state and connector updates related to mailbox message operations

// MoveMessagesFromMailbox moves messages from one mailbox to the other.
func MoveMessagesFromMailbox(
	ctx context.Context,
	tx *ent.Tx,
	mboxFromID, mboxToID imap.InternalMailboxID,
	messageIDs []imap.InternalMessageID,
) (map[imap.InternalMessageID]int, []Update, error) {
	if mboxFromID != mboxToID {
		if err := db.RemoveMessagesFromMailbox(ctx, tx, messageIDs, mboxFromID); err != nil {
			return nil, nil, err
		}
	}

	if _, err := db.AddMessagesToMailbox(ctx, tx, messageIDs, mboxToID); err != nil {
		return nil, nil, err
	}

	messageUIDs, err := db.GetMessageUIDs(ctx, tx.Client(), mboxToID, messageIDs)
	if err != nil {
		return nil, nil, err
	}

	stateUpdates := make([]Update, 0, len(messageIDs)+1)
	{
		responders := xslices.Map(messageIDs, func(id imap.InternalMessageID) Responder {
			return NewExists(id, messageUIDs[id])
		})
		stateUpdates = append(stateUpdates, NewMailboxIDResponderStateUpdate(mboxToID, responders...))
	}

	for _, messageID := range messageIDs {
		stateUpdates = append(stateUpdates, NewMessageIDAndMailboxIDResponderStateUpdate(messageID, mboxFromID, NewExpunge(messageID, contexts.IsClose(ctx))))
	}

	return messageUIDs, stateUpdates, nil
}

// AddMessagesToMailbox adds the messages to the given mailbox.
func AddMessagesToMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) (map[imap.InternalMessageID]int, Update, error) {
	messageUIDs, err := db.AddMessagesToMailbox(ctx, tx, messageIDs, mboxID)
	if err != nil {
		return nil, nil, err
	}

	responders := xslices.Map(messageIDs, func(id imap.InternalMessageID) Responder {
		return NewExists(id, messageUIDs[id])
	})

	return messageUIDs, NewMailboxIDResponderStateUpdate(mboxID, responders...), nil
}

// RemoveMessagesFromMailbox removes the messages from the given mailbox.
func RemoveMessagesFromMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]Update, error) {
	if len(messageIDs) > 0 {
		if err := db.RemoveMessagesFromMailbox(ctx, tx, messageIDs, mboxID); err != nil {
			return nil, err
		}
	}

	stateUpdates := xslices.Map(messageIDs, func(id imap.InternalMessageID) Update {
		return NewMessageIDAndMailboxIDResponderStateUpdate(id, mboxID, NewExpunge(id, contexts.IsClose(ctx)))
	})

	return stateUpdates, nil
}
