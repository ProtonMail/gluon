package state

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/db/ent"
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
	return s.PushResponder(ctx, tx, NewFetch(u.MessageID, imap.NewFlagSet(u.flag), contexts.IsUID(ctx), contexts.IsSilent(ctx), false, FetchFlagOpAdd))
}

func (u *RemoteAddMessageFlagsStateUpdate) String() string {
	return fmt.Sprintf("RemoteAddMessageFlagsStateUpdate %v flag = %v", u.MessageIDStateFilter.String(), u.flag)
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
	return s.PushResponder(ctx, tx, NewFetch(u.MessageID, imap.NewFlagSet(u.flag), contexts.IsUID(ctx), contexts.IsSilent(ctx), false, FetchFlagOpRem))
}

func (u *RemoteRemoveMessageFlagsStateUpdate) String() string {
	return fmt.Sprintf("RemoteRemoveMessageFlagsStateUpdate %v flag = %v", u.MessageIDStateFilter.String(), u.flag)
}

type RemoteMessageDeletedStateUpdate struct {
	MessageIDStateFilter
	remoteID imap.MessageID
}
