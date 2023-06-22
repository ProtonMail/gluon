package state

import (
	"context"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/limits"
	"github.com/bradenaw/juniper/xslices"
)

// Shared code for state and connector updates related to mailbox message operations

// MoveMessagesFromMailbox moves messages from one mailbox to the other.
func MoveMessagesFromMailbox(
	ctx context.Context,
	tx db.Transaction,
	mboxFromID, mboxToID imap.InternalMailboxID,
	messageIDPairs []db.MessageIDPair,
	messageIDsInternal []imap.InternalMessageID,
	s *State,
	imapLimits limits.IMAP,
	removeOldMessages bool,
) ([]db.UIDWithFlags, []Update, error) {
	messageCount, uid, err := tx.GetMailboxMessageCountAndUID(ctx, mboxToID)
	if err != nil {
		return nil, nil, err
	}

	if err := imapLimits.CheckMailBoxMessageCount(messageCount, len(messageIDPairs)); err != nil {
		return nil, nil, err
	}

	if err := imapLimits.CheckUIDCount(uid, len(messageIDPairs)); err != nil {
		return nil, nil, err
	}

	if mboxFromID != mboxToID && removeOldMessages {
		if err := tx.RemoveMessagesFromMailbox(ctx, mboxFromID, messageIDsInternal); err != nil {
			return nil, nil, err
		}
	}

	messageUIDs, err := tx.AddMessagesToMailbox(ctx, mboxToID, messageIDPairs)
	if err != nil {
		return nil, nil, err
	}

	stateUpdates := make([]Update, 0, len(messageIDPairs)+1)
	{
		responders := xslices.Map(messageUIDs, func(uid db.UIDWithFlags) *exists {
			return newExists(db.MessageIDPair{
				InternalID: uid.InternalID,
				RemoteID:   uid.RemoteID,
			}, uid.UID, uid.GetFlagSet())
		})
		stateUpdates = append(stateUpdates, newExistsStateUpdateWithExists(mboxToID, responders, s))
	}

	if removeOldMessages {
		for _, messageID := range messageIDsInternal {
			stateUpdates = append(stateUpdates, NewMessageIDAndMailboxIDResponderStateUpdate(messageID, mboxFromID, NewExpunge(messageID)))
		}
	}

	return messageUIDs, stateUpdates, nil
}

// AddMessagesToMailbox adds the messages to the given mailbox.
func AddMessagesToMailbox(ctx context.Context,
	tx db.Transaction,
	mboxID imap.InternalMailboxID,
	messageIDs []db.MessageIDPair,
	s *State,
	imapLimits limits.IMAP) ([]db.UIDWithFlags, Update, error) {
	messageCount, uid, err := tx.GetMailboxMessageCountAndUID(ctx, mboxID)
	if err != nil {
		return nil, nil, err
	}

	if err := imapLimits.CheckMailBoxMessageCount(messageCount, len(messageIDs)); err != nil {
		return nil, nil, err
	}

	if err := imapLimits.CheckUIDCount(uid, len(messageIDs)); err != nil {
		return nil, nil, err
	}

	messageUIDs, err := tx.AddMessagesToMailbox(ctx, mboxID, messageIDs)
	if err != nil {
		return nil, nil, err
	}

	responders := xslices.Map(messageUIDs, func(uid db.UIDWithFlags) *exists {
		return newExists(db.MessageIDPair{
			InternalID: uid.InternalID,
			RemoteID:   uid.RemoteID,
		}, uid.UID, uid.GetFlagSet())
	})

	return messageUIDs, newExistsStateUpdateWithExists(mboxID, responders, s), nil
}

// RemoveMessagesFromMailbox removes the messages from the given mailbox.
func RemoveMessagesFromMailbox(ctx context.Context, tx db.Transaction, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]Update, error) {
	if len(messageIDs) > 0 {
		if err := tx.RemoveMessagesFromMailbox(ctx, mboxID, messageIDs); err != nil {
			return nil, err
		}
	}

	stateUpdates := xslices.Map(messageIDs, func(id imap.InternalMessageID) Update {
		return NewMessageIDAndMailboxIDResponderStateUpdate(id, mboxID, NewExpunge(id))
	})

	return stateUpdates, nil
}
