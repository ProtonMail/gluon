package state

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/bradenaw/juniper/xslices"
)

type Update interface {
	// Filter returns true when the state can be passed into A.
	Filter(s *State) bool
	// Apply the update to a given state.
	Apply(cxt context.Context, tx *ent.Tx, s *State) error

	String() string
}

type messageFlagsAddedStateUpdate struct {
	AnyMessageIDStateFilter
	flags   imap.FlagSet
	mboxID  ids.MailboxIDPair
	stateID int
}

func newMessageFlagsAddedStateUpdate(flags imap.FlagSet, mboxID ids.MailboxIDPair, messageIDs []imap.InternalMessageID, stateID int) Update {
	return &messageFlagsAddedStateUpdate{
		flags:                   flags,
		mboxID:                  mboxID,
		AnyMessageIDStateFilter: AnyMessageIDStateFilter{MessageIDs: messageIDs},
		stateID:                 stateID,
	}
}

func (u *messageFlagsAddedStateUpdate) Apply(ctx context.Context, tx *ent.Tx, s *State) error {
	for _, messageID := range u.MessageIDs {
		snapFlags, err := s.snap.getMessageFlags(messageID)
		if err != nil {
			return err
		}

		newFlags := snapFlags.AddFlagSet(u.flags)

		if s.snap.mboxID != u.mboxID {
			newFlags = newFlags.Set(imap.FlagDeleted, snapFlags.Contains(imap.FlagDeleted))
		}

		if err := s.PushResponder(ctx, tx, NewFetch(
			messageID,
			newFlags,
			contexts.IsUID(ctx),
			s.StateID == u.stateID && contexts.IsSilent(ctx),
		)); err != nil {
			return err
		}
	}

	return nil
}

func (u *messageFlagsAddedStateUpdate) String() string {
	return fmt.Sprintf("MessagFlagsAddedStateUpdate: mbox = %v messages = %v flags = %v",
		u.mboxID.InternalID.ShortID(),
		xslices.Map(u.MessageIDs, func(id imap.InternalMessageID) string {
			return id.ShortID()
		}),
		u.flags)
}

// applyMessageFlagsAdded adds the flags to the given messages.
func (state *State) applyMessageFlagsAdded(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, addFlags imap.FlagSet) error {
	if addFlags.Contains(imap.FlagRecent) {
		return fmt.Errorf("the recent flag is read-only")
	}

	client := tx.Client()

	curFlags, err := db.GetMessageFlags(ctx, client, messageIDs)
	if err != nil {
		return err
	}

	delFlags, err := db.GetMessageDeleted(ctx, client, state.snap.mboxID.InternalID, messageIDs)
	if err != nil {
		return err
	}

	if addFlags.Contains(imap.FlagDeleted) {
		if err := db.SetDeletedFlag(ctx, tx, state.snap.mboxID.InternalID, xslices.Filter(messageIDs, func(messageID imap.InternalMessageID) bool {
			return !delFlags[messageID]
		}), true); err != nil {
			return err
		}
	}

	for _, flag := range addFlags.Remove(imap.FlagDeleted).ToSlice() {
		if err := db.AddMessageFlag(ctx, tx, xslices.Filter(messageIDs, func(messageID imap.InternalMessageID) bool {
			return !curFlags[messageID].Contains(flag)
		}), flag); err != nil {
			return err
		}
	}

	if err := state.user.QueueOrApplyStateUpdate(ctx, tx, newMessageFlagsAddedStateUpdate(addFlags, state.snap.mboxID, messageIDs, state.StateID)); err != nil {
		return err
	}

	return nil
}

type messageFlagsRemovedStateUpdate struct {
	AnyMessageIDStateFilter
	flags   imap.FlagSet
	mboxID  ids.MailboxIDPair
	stateID int
}

func NewMessageFlagsRemovedStateUpdate(flags imap.FlagSet, mboxID ids.MailboxIDPair, messageIDs []imap.InternalMessageID, stateID int) Update {
	return &messageFlagsRemovedStateUpdate{
		flags:                   flags,
		mboxID:                  mboxID,
		AnyMessageIDStateFilter: AnyMessageIDStateFilter{MessageIDs: messageIDs},
		stateID:                 stateID,
	}
}

func (u *messageFlagsRemovedStateUpdate) Apply(ctx context.Context, tx *ent.Tx, s *State) error {
	for _, messageID := range u.MessageIDs {
		snapFlags, err := s.snap.getMessageFlags(messageID)
		if err != nil {
			return err
		}

		newFlags := snapFlags.RemoveFlagSet(u.flags)

		if s.snap.mboxID != u.mboxID {
			newFlags = newFlags.Set(imap.FlagDeleted, snapFlags.Contains(imap.FlagDeleted))
		}

		if err := s.PushResponder(ctx, tx, NewFetch(
			messageID,
			newFlags,
			contexts.IsUID(ctx),
			s.StateID == u.stateID && contexts.IsSilent(ctx),
		)); err != nil {
			return err
		}
	}

	return nil
}

func (u *messageFlagsRemovedStateUpdate) String() string {
	return fmt.Sprintf("MessagFlagsRemovedStateUpdate: mbox = %v messages = %v flags = %v",
		u.mboxID.InternalID,
		xslices.Map(u.MessageIDs, func(id imap.InternalMessageID) string {
			return id.ShortID()
		}),
		u.flags)
}

// applyMessageFlagsRemoved removes the flags from the given messages.
func (state *State) applyMessageFlagsRemoved(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, remFlags imap.FlagSet) error {
	if remFlags.Contains(imap.FlagRecent) {
		return fmt.Errorf("the recent flag is read-only")
	}

	client := tx.Client()

	curFlags, err := db.GetMessageFlags(ctx, client, messageIDs)
	if err != nil {
		return err
	}

	delFlags, err := db.GetMessageDeleted(ctx, client, state.snap.mboxID.InternalID, messageIDs)
	if err != nil {
		return err
	}

	if remFlags.Contains(imap.FlagDeleted) {
		if err := db.SetDeletedFlag(ctx, tx, state.snap.mboxID.InternalID, xslices.Filter(messageIDs, func(messageID imap.InternalMessageID) bool {
			return delFlags[messageID]
		}), false); err != nil {
			return err
		}
	}

	for _, flag := range remFlags.Remove(imap.FlagDeleted).ToSlice() {
		if err := db.RemoveMessageFlag(ctx, tx, xslices.Filter(messageIDs, func(messageID imap.InternalMessageID) bool {
			return curFlags[messageID].Contains(flag)
		}), flag); err != nil {
			return err
		}
	}

	if err := state.user.QueueOrApplyStateUpdate(ctx, tx, NewMessageFlagsRemovedStateUpdate(remFlags, state.snap.mboxID, messageIDs, state.StateID)); err != nil {
		return err
	}

	res := make(map[imap.InternalMessageID]imap.FlagSet)

	for _, messageID := range messageIDs {
		res[messageID] = curFlags[messageID].RemoveFlagSet(remFlags)
	}

	return nil
}

type messageFlagsSetStateUpdate struct {
	AnyMessageIDStateFilter
	flags   imap.FlagSet
	mboxID  ids.MailboxIDPair
	stateID int
}

func NewMessageFlagsSetStateUpdate(flags imap.FlagSet, mboxID ids.MailboxIDPair, messageIDs []imap.InternalMessageID, stateID int) Update {
	return &messageFlagsSetStateUpdate{
		flags:                   flags,
		mboxID:                  mboxID,
		AnyMessageIDStateFilter: AnyMessageIDStateFilter{MessageIDs: messageIDs},
		stateID:                 stateID,
	}
}

func (u *messageFlagsSetStateUpdate) Apply(ctx context.Context, tx *ent.Tx, state *State) error {
	for _, messageID := range u.MessageIDs {
		newFlags := u.flags

		if state.snap.mboxID != u.mboxID {
			snapFlags, err := state.snap.getMessageFlags(messageID)
			if err != nil {
				return err
			}

			newFlags = newFlags.Set(imap.FlagDeleted, snapFlags.Contains(imap.FlagDeleted))
		}

		if err := state.PushResponder(ctx, tx, NewFetch(
			messageID,
			newFlags,
			contexts.IsUID(ctx),
			state.StateID == u.stateID && contexts.IsSilent(ctx),
		)); err != nil {
			return err
		}
	}

	return nil
}

func (u *messageFlagsSetStateUpdate) String() string {
	return fmt.Sprintf("MessagFlagsSetStateUpdate: mbox = %v messages = %v flags=%v", u.mboxID.InternalID.ShortID(), u.MessageIDs, u.flags)
}

// applyMessageFlagsSet sets the flags of the given messages.
func (state *State) applyMessageFlagsSet(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, setFlags imap.FlagSet) error {
	if setFlags.Contains(imap.FlagRecent) {
		return fmt.Errorf("the recent flag is read-only")
	}

	if state.snap == nil {
		return nil
	}

	if err := db.SetDeletedFlag(ctx, tx, state.snap.mboxID.InternalID, messageIDs, setFlags.Contains(imap.FlagDeleted)); err != nil {
		return err
	}

	if err := db.SetMessageFlags(ctx, tx, messageIDs, setFlags.Remove(imap.FlagDeleted)); err != nil {
		return err
	}

	if err := state.user.QueueOrApplyStateUpdate(ctx, tx, NewMessageFlagsSetStateUpdate(setFlags, state.snap.mboxID, messageIDs, state.StateID)); err != nil {
		return err
	}

	return nil
}

type mailboxRemoteIDUpdateStateUpdate struct {
	SnapFilter
	remoteID imap.LabelID
}

func NewMailboxRemoteIDUpdateStateUpdate(internalID imap.InternalMailboxID, remoteID imap.LabelID) Update {
	return &mailboxRemoteIDUpdateStateUpdate{
		SnapFilter: NewMBoxIDStateFilter(internalID),
		remoteID:   remoteID,
	}
}

func (u *mailboxRemoteIDUpdateStateUpdate) Apply(ctx context.Context, tx *ent.Tx, s *State) error {
	s.snap.mboxID.RemoteID = u.remoteID

	return nil
}

func (u *mailboxRemoteIDUpdateStateUpdate) String() string {
	return fmt.Sprintf("MailboxRemoteIDUpdateStateUpdate: %v remote = %v", u.SnapFilter, u.remoteID.ShortID())
}
