package state

import (
	"context"
	"errors"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/db/ent"
	errors2 "github.com/ProtonMail/gluon/internal/errors"
	"github.com/ProtonMail/gluon/internal/ids"
)

type RemoteAddMessageFlagsStateUpdate struct {
	MessageIDStateFilter
	flag string
}

func NewRemoteAddMessageFlagsStateUpdate(messageID imap.InternalMessageID, flag string) Update {
	return &RemoteAddMessageFlagsStateUpdate{
		MessageIDStateFilter: MessageIDStateFilter{MessageID: messageID},
		flag:                 flag,
	}
}

func (u *RemoteAddMessageFlagsStateUpdate) Apply(ctx context.Context, tx *ent.Tx, s *State) error {
	snapFlags, err := s.snap.getMessageFlags(u.MessageID)
	if err != nil {
		return err
	}

	return s.PushResponder(ctx, tx, NewFetch(u.MessageID, snapFlags.Add(u.flag), contexts.IsUID(ctx), contexts.IsSilent(ctx)))
}

type RemoteRemoveMessageFlagsStateUpdate struct {
	MessageIDStateFilter
	flag string
}

func NewRemoteRemoveMessageFlagsStateUpdate(messageID imap.InternalMessageID, flag string) Update {
	return &RemoteRemoveMessageFlagsStateUpdate{
		MessageIDStateFilter: MessageIDStateFilter{MessageID: messageID},
		flag:                 flag,
	}
}

func (u *RemoteRemoveMessageFlagsStateUpdate) Apply(ctx context.Context, tx *ent.Tx, s *State) error {
	snapFlags, err := s.snap.getMessageFlags(u.MessageID)
	if err != nil {
		if errors.Is(err, errors2.ErrNoSuchMessage) {
			return nil
		}

		return err
	}

	return s.PushResponder(ctx, tx, NewFetch(u.MessageID, snapFlags.Remove(u.flag), contexts.IsUID(ctx), contexts.IsSilent(ctx)))
}

type RemoteMessageDeletedStateUpdate struct {
	MessageIDStateFilter
	remoteID imap.MessageID
}

func NewRemoteMessageDeletedStateUpdate(messageID imap.InternalMessageID, remoteID imap.MessageID) Update {
	return &RemoteMessageDeletedStateUpdate{
		MessageIDStateFilter: MessageIDStateFilter{MessageID: messageID},
		remoteID:             remoteID,
	}
}

func (u *RemoteMessageDeletedStateUpdate) Apply(ctx context.Context, tx *ent.Tx, s *State) error {
	return s.actionRemoveMessagesFromMailbox(ctx, tx, []ids.MessageIDPair{{
		InternalID: u.MessageID,
		RemoteID:   u.remoteID,
	}}, s.snap.mboxID)
}
