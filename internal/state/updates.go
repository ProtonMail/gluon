package state

import (
	"context"
	"fmt"
	"strings"

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
	AllStateFilter
	messageIDs []imap.InternalMessageID
	flags      imap.FlagSet
	mboxID     ids.MailboxIDPair
	stateID    StateID
}

func newMessageFlagsAddedStateUpdate(flags imap.FlagSet, mboxID ids.MailboxIDPair, messageIDs []imap.InternalMessageID, stateID StateID) Update {
	return &messageFlagsAddedStateUpdate{
		flags:      flags,
		mboxID:     mboxID,
		messageIDs: messageIDs,
		stateID:    stateID,
	}
}

func (u *messageFlagsAddedStateUpdate) Apply(ctx context.Context, tx *ent.Tx, s *State) error {
	for _, messageID := range u.messageIDs {
		newFlags := u.flags

		if err := s.PushResponder(ctx, tx, NewFetch(
			messageID,
			newFlags,
			contexts.IsUID(ctx),
			s.StateID == u.stateID && contexts.IsSilent(ctx),
			s.snap.mboxID != u.mboxID,
			FetchFlagOpAdd,
		)); err != nil {
			return err
		}
	}

	return nil
}

func (u *messageFlagsAddedStateUpdate) String() string {
	return fmt.Sprintf("MessagFlagsAddedStateUpdate: mbox = %v messages = %v flags = %v",
		u.mboxID.InternalID.ShortID(),
		xslices.Map(u.messageIDs, func(id imap.InternalMessageID) string {
			return id.ShortID()
		}),
		u.flags)
}

// applyMessageFlagsAdded adds the flags to the given messages.
func (state *State) applyMessageFlagsAdded(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, addFlags imap.FlagSet) error {
	if addFlags.ContainsUnchecked(imap.FlagRecentLowerCase) {
		return fmt.Errorf("the recent flag is read-only")
	}

	client := tx.Client()

	curFlags, err := db.GetMessageFlags(ctx, client, messageIDs)
	if err != nil {
		return err
	}

	if addFlags.ContainsUnchecked(imap.FlagDeletedLowerCase) {
		if err := db.SetDeletedFlag(ctx, tx, state.snap.mboxID.InternalID, messageIDs, true); err != nil {
			return err
		}
	}

	for _, flag := range addFlags.Remove(imap.FlagDeleted).ToSlice() {
		flagLowerCase := strings.ToLower(flag)

		var messagesToFlag []imap.InternalMessageID

		for _, v := range curFlags {
			if !v.FlagSet.ContainsUnchecked(flagLowerCase) {
				messagesToFlag = append(messagesToFlag, v.ID)
			}
		}

		if err := db.AddMessageFlag(ctx, tx, messagesToFlag, flag); err != nil {
			return err
		}
	}

	if err := state.user.QueueOrApplyStateUpdate(ctx, tx, newMessageFlagsAddedStateUpdate(addFlags, state.snap.mboxID, messageIDs, state.StateID)); err != nil {
		return err
	}

	return nil
}

type messageFlagsRemovedStateUpdate struct {
	AllStateFilter
	messageIDs []imap.InternalMessageID
	flags      imap.FlagSet
	mboxID     ids.MailboxIDPair
	stateID    StateID
}

func NewMessageFlagsRemovedStateUpdate(flags imap.FlagSet, mboxID ids.MailboxIDPair, messageIDs []imap.InternalMessageID, stateID StateID) Update {
	return &messageFlagsRemovedStateUpdate{
		flags:      flags,
		mboxID:     mboxID,
		messageIDs: messageIDs,
		stateID:    stateID,
	}
}

func (u *messageFlagsRemovedStateUpdate) Apply(ctx context.Context, tx *ent.Tx, s *State) error {
	for _, messageID := range u.messageIDs {
		newFlags := u.flags

		if err := s.PushResponder(ctx, tx, NewFetch(
			messageID,
			newFlags,
			contexts.IsUID(ctx),
			s.StateID == u.stateID && contexts.IsSilent(ctx),
			s.snap.mboxID != u.mboxID,
			FetchFlagOpRem,
		)); err != nil {
			return err
		}
	}

	return nil
}

func (u *messageFlagsRemovedStateUpdate) String() string {
	return fmt.Sprintf("MessagFlagsRemovedStateUpdate: mbox = %v messages = %v flags = %v",
		u.mboxID.InternalID,
		xslices.Map(u.messageIDs, func(id imap.InternalMessageID) string {
			return id.ShortID()
		}),
		u.flags)
}

// applyMessageFlagsRemoved removes the flags from the given messages.
func (state *State) applyMessageFlagsRemoved(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, remFlags imap.FlagSet) error {
	if remFlags.ContainsUnchecked(imap.FlagRecentLowerCase) {
		return fmt.Errorf("the recent flag is read-only")
	}

	client := tx.Client()

	curFlags, err := db.GetMessageFlags(ctx, client, messageIDs)
	if err != nil {
		return err
	}

	if remFlags.ContainsUnchecked(imap.FlagDeletedLowerCase) {
		if err := db.SetDeletedFlag(ctx, tx, state.snap.mboxID.InternalID, messageIDs, false); err != nil {
			return err
		}
	}

	for _, flag := range remFlags.Remove(imap.FlagDeleted).ToSlice() {
		flagLowerCase := strings.ToLower(flag)

		var messagesToFlag []imap.InternalMessageID

		for _, v := range curFlags {
			if v.FlagSet.ContainsUnchecked(flagLowerCase) {
				messagesToFlag = append(messagesToFlag, v.ID)
			}
		}

		if err := db.RemoveMessageFlag(ctx, tx, messagesToFlag, flag); err != nil {
			return err
		}
	}

	if err := state.user.QueueOrApplyStateUpdate(ctx, tx, NewMessageFlagsRemovedStateUpdate(remFlags, state.snap.mboxID, messageIDs, state.StateID)); err != nil {
		return err
	}

	return nil
}

type messageFlagsSetStateUpdate struct {
	AllStateFilter
	messageIDs []imap.InternalMessageID
	flags      imap.FlagSet
	mboxID     ids.MailboxIDPair
	stateID    StateID
}

func NewMessageFlagsSetStateUpdate(flags imap.FlagSet, mboxID ids.MailboxIDPair, messageIDs []imap.InternalMessageID, stateID StateID) Update {
	return &messageFlagsSetStateUpdate{
		flags:      flags,
		mboxID:     mboxID,
		messageIDs: messageIDs,
		stateID:    stateID,
	}
}

func (u *messageFlagsSetStateUpdate) Apply(ctx context.Context, tx *ent.Tx, state *State) error {
	for _, messageID := range u.messageIDs {
		newFlags := u.flags

		if err := state.PushResponder(ctx, tx, NewFetch(
			messageID,
			newFlags,
			contexts.IsUID(ctx),
			state.StateID == u.stateID && contexts.IsSilent(ctx),
			state.snap.mboxID != u.mboxID,
			FetchFlagOpSet,
		)); err != nil {
			return err
		}
	}

	return nil
}

func (u *messageFlagsSetStateUpdate) String() string {
	return fmt.Sprintf("MessageFlagsSetStateUpdate: mbox = %v messages = %v flags=%v",
		u.mboxID.InternalID.ShortID(),
		xslices.Map(u.messageIDs, func(id imap.InternalMessageID) string {
			return id.ShortID()
		}),
		u.flags)
}

// applyMessageFlagsSet sets the flags of the given messages.
func (state *State) applyMessageFlagsSet(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, setFlags imap.FlagSet) error {
	if setFlags.ContainsUnchecked(imap.FlagRecentLowerCase) {
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
	remoteID imap.MailboxID
}

func NewMailboxRemoteIDUpdateStateUpdate(internalID imap.InternalMailboxID, remoteID imap.MailboxID) Update {
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
	return fmt.Sprintf("MailboxRemoteIDUpdateStateUpdate: %v remote = %v", u.SnapFilter.String(), u.remoteID.ShortID())
}

type mailboxDeletedStateUpdate struct {
	MBoxIDStateFilter
}

func NewMailboxDeletedStateUpdate(mboxID imap.InternalMailboxID) Update {
	return &mailboxDeletedStateUpdate{MBoxIDStateFilter: MBoxIDStateFilter{MboxID: mboxID}}
}

func (u *mailboxDeletedStateUpdate) Apply(ctx context.Context, tx *ent.Tx, s *State) error {
	s.markInvalid()

	return nil
}

func (u *mailboxDeletedStateUpdate) String() string {
	return fmt.Sprintf("MailboxDeletedStateUpdate: %v", u.MBoxIDStateFilter.String())
}

type uidValidityBumpedStateUpdate struct {
	AllStateFilter
}

func NewUIDValidityBumpedStateUpdate() Update {
	return &uidValidityBumpedStateUpdate{}
}

func (u *uidValidityBumpedStateUpdate) Apply(ctx context.Context, tx *ent.Tx, s *State) error {
	s.markInvalid()

	return nil
}

func (u *uidValidityBumpedStateUpdate) String() string {
	return "UIDValidityBumpedStateUpdate"
}
