package state

import (
	"context"
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/bradenaw/juniper/xslices"
)

type Update interface {
	// Filter returns true when the state can be passed into A.
	Filter(s *State) bool
	// Apply the update to a given state.
	Apply(cxt context.Context, tx db.Transaction, s *State) error

	String() string
}

type messageFlagsComboStateUpdate struct {
	AllStateFilter
	updates []Update
}

func newMessageFlagsComboStateUpdate() *messageFlagsComboStateUpdate {
	return &messageFlagsComboStateUpdate{}
}

func (u *messageFlagsComboStateUpdate) Apply(ctx context.Context, tx db.Transaction, s *State) error {
	for _, v := range u.updates {
		if err := v.Apply(ctx, tx, s); err != nil {
			return err
		}
	}

	return nil
}

func (u *messageFlagsComboStateUpdate) addUpdate(update Update) {
	u.updates = append(u.updates, update)
}

func (u *messageFlagsComboStateUpdate) String() string {
	return fmt.Sprintf("messageFlagsComboStateUpdate: %v", u.updates)
}

type messageFlagsAddedStateUpdate struct {
	AllStateFilter
	messageIDs []imap.InternalMessageID
	flags      imap.FlagSet
	mboxID     db.MailboxIDPair
	stateID    StateID
}

func newMessageFlagsAddedStateUpdate(flags imap.FlagSet, mboxID db.MailboxIDPair, messageIDs []imap.InternalMessageID, stateID StateID) Update {
	return &messageFlagsAddedStateUpdate{
		flags:      flags,
		mboxID:     mboxID,
		messageIDs: messageIDs,
		stateID:    stateID,
	}
}

func (u *messageFlagsAddedStateUpdate) Apply(ctx context.Context, tx db.Transaction, s *State) error {
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
func (state *State) applyMessageFlagsAdded(ctx context.Context,
	tx db.Transaction,
	messageIDs []imap.InternalMessageID,
	addFlags imap.FlagSet) ([]Update, error) {
	if addFlags.ContainsUnchecked(imap.FlagRecentLowerCase) {
		return nil, fmt.Errorf("the recent flag is read-only")
	}

	var allUpdates []Update

	// Since DB state can be more up to date then the flag state we should only emit add flag updates for values
	// that actually changed.

	curFlags, err := tx.GetMessagesFlags(ctx, messageIDs)
	if err != nil {
		return nil, err
	}

	// If setting messages as seen, only set those messages that aren't currently seen.
	if addFlags.ContainsUnchecked(imap.FlagSeenLowerCase) {
		var messagesToApply []imap.MessageID

		for _, msg := range curFlags {
			if !msg.FlagSet.ContainsUnchecked(imap.FlagSeenLowerCase) && !ids.IsRecoveredRemoteMessageID(msg.RemoteID) {
				messagesToApply = append(messagesToApply, msg.RemoteID)
			}
		}

		if len(messagesToApply) != 0 {
			updates, err := state.user.GetRemote().SetMessagesSeen(ctx, tx, messagesToApply, true)
			if err != nil {
				return nil, err
			}

			allUpdates = append(allUpdates, updates...)
		}
	}

	// If setting messages as flagged, only set those messages that aren't currently flagged.
	if addFlags.ContainsUnchecked(imap.FlagFlaggedLowerCase) {
		var messagesToApply []imap.MessageID

		for _, msg := range curFlags {
			if !msg.FlagSet.ContainsUnchecked(imap.FlagFlaggedLowerCase) && !ids.IsRecoveredRemoteMessageID(msg.RemoteID) {
				messagesToApply = append(messagesToApply, msg.RemoteID)
			}
		}

		if len(messagesToApply) != 0 {
			updates, err := state.user.GetRemote().SetMessagesFlagged(ctx, tx, messagesToApply, true)
			if err != nil {
				return nil, err
			}

			allUpdates = append(allUpdates, updates...)
		}
	}

	flagStateUpdate := newMessageFlagsComboStateUpdate()

	if addFlags.ContainsUnchecked(imap.FlagDeletedLowerCase) {
		if err := tx.SetMailboxMessagesDeletedFlag(ctx, state.snap.mboxID.InternalID, messageIDs, true); err != nil {
			return nil, err
		}

		flagStateUpdate.addUpdate(newMessageFlagsAddedStateUpdate(imap.NewFlagSet(imap.FlagDeleted), state.snap.mboxID, messageIDs, state.StateID))
	}

	remainingFlags := addFlags.Remove(imap.FlagDeleted)
	for _, flag := range remainingFlags {
		flagLowerCase := strings.ToLower(flag)

		messagesToFlag := make([]imap.InternalMessageID, 0, len(messageIDs)/2)

		for _, v := range curFlags {
			if !v.FlagSet.ContainsUnchecked(flagLowerCase) {
				messagesToFlag = append(messagesToFlag, v.ID)
			}
		}

		if err := tx.AddFlagToMessages(ctx, messagesToFlag, flag); err != nil {
			return nil, err
		}

		flagStateUpdate.addUpdate(newMessageFlagsAddedStateUpdate(remainingFlags, state.snap.mboxID, messagesToFlag, state.StateID))
	}

	return append(allUpdates, flagStateUpdate), nil
}

type messageFlagsRemovedStateUpdate struct {
	AllStateFilter
	messageIDs []imap.InternalMessageID
	flags      imap.FlagSet
	mboxID     db.MailboxIDPair
	stateID    StateID
}

func NewMessageFlagsRemovedStateUpdate(flags imap.FlagSet, mboxID db.MailboxIDPair, messageIDs []imap.InternalMessageID, stateID StateID) Update {
	return &messageFlagsRemovedStateUpdate{
		flags:      flags,
		mboxID:     mboxID,
		messageIDs: messageIDs,
		stateID:    stateID,
	}
}

func (u *messageFlagsRemovedStateUpdate) Apply(ctx context.Context, tx db.Transaction, s *State) error {
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
func (state *State) applyMessageFlagsRemoved(ctx context.Context,
	tx db.Transaction,
	messageIDs []imap.InternalMessageID,
	remFlags imap.FlagSet) ([]Update, error) {
	if remFlags.ContainsUnchecked(imap.FlagRecentLowerCase) {
		return nil, fmt.Errorf("the recent flag is read-only")
	}

	var allUpdates []Update

	curFlags, err := tx.GetMessagesFlags(ctx, messageIDs)
	if err != nil {
		return nil, err
	}
	// If setting messages as unseen, only set those messages that are currently seen.
	if remFlags.ContainsUnchecked(imap.FlagSeenLowerCase) {
		var messagesToApply []imap.MessageID

		for _, msg := range curFlags {
			if msg.FlagSet.ContainsUnchecked(imap.FlagSeenLowerCase) && !ids.IsRecoveredRemoteMessageID(msg.RemoteID) {
				messagesToApply = append(messagesToApply, msg.RemoteID)
			}
		}

		if len(messagesToApply) != 0 {
			updates, err := state.user.GetRemote().SetMessagesSeen(ctx, tx, messagesToApply, false)
			if err != nil {
				return nil, err
			}

			allUpdates = append(allUpdates, updates...)
		}
	}

	// If setting messages as unflagged, only set those messages that are currently flagged.
	if remFlags.ContainsUnchecked(imap.FlagFlaggedLowerCase) {
		var messagesToApply []imap.MessageID

		for _, msg := range curFlags {
			if msg.FlagSet.ContainsUnchecked(imap.FlagFlaggedLowerCase) && !ids.IsRecoveredRemoteMessageID(msg.RemoteID) {
				messagesToApply = append(messagesToApply, msg.RemoteID)
			}
		}

		if len(messagesToApply) != 0 {
			updates, err := state.user.GetRemote().SetMessagesFlagged(ctx, tx, messagesToApply, false)
			if err != nil {
				return nil, err
			}

			allUpdates = append(allUpdates, updates...)
		}
	}

	flagStateUpdate := newMessageFlagsComboStateUpdate()

	if remFlags.ContainsUnchecked(imap.FlagDeletedLowerCase) {
		if err := tx.SetMailboxMessagesDeletedFlag(ctx, state.snap.mboxID.InternalID, messageIDs, false); err != nil {
			return nil, err
		}

		flagStateUpdate.addUpdate(NewMessageFlagsRemovedStateUpdate(imap.NewFlagSet(imap.FlagDeleted), state.snap.mboxID, messageIDs, state.StateID))
	}

	remainingFlags := remFlags.Remove(imap.FlagDeleted)
	for _, flag := range remainingFlags {
		flagLowerCase := strings.ToLower(flag)

		messagesToFlag := make([]imap.InternalMessageID, 0, len(messageIDs)/2)

		for _, v := range curFlags {
			if v.FlagSet.ContainsUnchecked(flagLowerCase) {
				messagesToFlag = append(messagesToFlag, v.ID)
			}
		}

		if err := tx.RemoveFlagFromMessages(ctx, messagesToFlag, flag); err != nil {
			return nil, err
		}

		flagStateUpdate.addUpdate(NewMessageFlagsRemovedStateUpdate(remainingFlags, state.snap.mboxID, messagesToFlag, state.StateID))
	}

	return append(allUpdates, flagStateUpdate), nil
}

type messageFlagsSetStateUpdate struct {
	AllStateFilter
	messageIDs []imap.InternalMessageID
	flags      imap.FlagSet
	mboxID     db.MailboxIDPair
	stateID    StateID
}

func NewMessageFlagsSetStateUpdate(flags imap.FlagSet, mboxID db.MailboxIDPair, messageIDs []imap.InternalMessageID, stateID StateID) Update {
	return &messageFlagsSetStateUpdate{
		flags:      flags,
		mboxID:     mboxID,
		messageIDs: messageIDs,
		stateID:    stateID,
	}
}

func (u *messageFlagsSetStateUpdate) Apply(ctx context.Context, tx db.Transaction, state *State) error {
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
func (state *State) applyMessageFlagsSet(ctx context.Context,
	tx db.Transaction,
	messageIDs []imap.InternalMessageID,
	setFlags imap.FlagSet) ([]Update, error) {
	if setFlags.ContainsUnchecked(imap.FlagRecentLowerCase) {
		return nil, fmt.Errorf("the recent flag is read-only")
	}

	if state.snap == nil {
		return nil, nil
	}

	curFlags, err := tx.GetMessagesFlags(ctx, messageIDs)
	if err != nil {
		return nil, err
	}

	// If setting messages as seen, only set those messages that aren't currently seen, and vice versa.
	setSeen := map[bool][]imap.MessageID{true: {}, false: {}}

	for _, msg := range curFlags {
		if seen := setFlags.ContainsUnchecked(imap.FlagSeenLowerCase); seen != msg.FlagSet.ContainsUnchecked(imap.FlagSeenLowerCase) && !ids.IsRecoveredRemoteMessageID(msg.RemoteID) {
			setSeen[seen] = append(setSeen[seen], msg.RemoteID)
		}
	}

	var allUpdates []Update

	for seen, messageIDs := range setSeen {
		updates, err := state.user.GetRemote().SetMessagesSeen(ctx, tx, messageIDs, seen)
		if err != nil {
			return nil, err
		}

		allUpdates = append(allUpdates, updates...)
	}

	// If setting messages as flagged, only set those messages that aren't currently flagged, and vice versa.
	setFlagged := map[bool][]imap.MessageID{true: {}, false: {}}

	for _, msg := range curFlags {
		if flagged := setFlags.ContainsUnchecked(imap.FlagFlaggedLowerCase); flagged != msg.FlagSet.ContainsUnchecked(imap.FlagFlaggedLowerCase) && !ids.IsRecoveredRemoteMessageID(msg.RemoteID) {
			setFlagged[flagged] = append(setFlagged[flagged], msg.RemoteID)
		}
	}

	for flagged, messageIDs := range setFlagged {
		updates, err := state.user.GetRemote().SetMessagesFlagged(ctx, tx, messageIDs, flagged)
		if err != nil {
			return nil, err
		}

		allUpdates = append(allUpdates, updates...)
	}

	if err := tx.SetMailboxMessagesDeletedFlag(ctx, state.snap.mboxID.InternalID, messageIDs, setFlags.Contains(imap.FlagDeleted)); err != nil {
		return nil, err
	}

	remainingFlags := setFlags.Remove(imap.FlagDeleted)
	if remainingFlags.Len() != 0 {
		if err := tx.SetFlagsOnMessages(ctx, messageIDs, remainingFlags); err != nil {
			return nil, err
		}
	}

	return append(allUpdates, NewMessageFlagsSetStateUpdate(setFlags, state.snap.mboxID, messageIDs, state.StateID)), nil
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

func (u *mailboxRemoteIDUpdateStateUpdate) Apply(ctx context.Context, tx db.Transaction, s *State) error {
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

func (u *mailboxDeletedStateUpdate) Apply(ctx context.Context, tx db.Transaction, s *State) error {
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

func (u *uidValidityBumpedStateUpdate) Apply(ctx context.Context, tx db.Transaction, s *State) error {
	s.markInvalid()

	return nil
}

func (u *uidValidityBumpedStateUpdate) String() string {
	return "UIDValidityBumpedStateUpdate"
}
