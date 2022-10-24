package ids

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db/ent"
)

type MailboxIDPair struct {
	InternalID imap.InternalMailboxID
	RemoteID   imap.MailboxID
}

func (m *MailboxIDPair) String() string {
	return fmt.Sprintf("%v::%v", m.InternalID, m.RemoteID)
}

type MessageIDPair struct {
	InternalID imap.InternalMessageID
	RemoteID   imap.MessageID
}

func (m *MessageIDPair) String() string {
	return fmt.Sprintf("%v::%v", m.InternalID, m.RemoteID)
}

func NewMailboxIDPair(mbox *ent.Mailbox) MailboxIDPair {
	return MailboxIDPair{
		InternalID: mbox.ID,
		RemoteID:   mbox.RemoteID,
	}
}

func NewMailboxIDPairWithoutRemote(internalID imap.InternalMailboxID) MailboxIDPair {
	return MailboxIDPair{
		InternalID: internalID,
		RemoteID:   "",
	}
}

func NewMessageIDPair(msg *ent.Message) MessageIDPair {
	return MessageIDPair{
		InternalID: msg.ID,
		RemoteID:   msg.RemoteID,
	}
}

func SplitMessageIDPairSlice(s []MessageIDPair) ([]imap.InternalMessageID, []imap.MessageID) {
	l := len(s)

	internalMessageIDs := make([]imap.InternalMessageID, 0, l)
	remoteMessageIDs := make([]imap.MessageID, 0, l)

	for _, v := range s {
		internalMessageIDs = append(internalMessageIDs, v.InternalID)
		remoteMessageIDs = append(remoteMessageIDs, v.RemoteID)
	}

	return internalMessageIDs, remoteMessageIDs
}

func SplitMailboxIDPairSlice(s []MailboxIDPair) ([]imap.InternalMailboxID, []imap.MailboxID) {
	l := len(s)

	internalMailboxIDs := make([]imap.InternalMailboxID, 0, l)
	mailboxIDs := make([]imap.MailboxID, 0, l)

	for _, v := range s {
		internalMailboxIDs = append(internalMailboxIDs, v.InternalID)
		mailboxIDs = append(mailboxIDs, v.RemoteID)
	}

	return internalMailboxIDs, mailboxIDs
}
