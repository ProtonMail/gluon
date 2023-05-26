package db

import (
	"fmt"
	"time"

	"github.com/ProtonMail/gluon/imap"
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

func NewMailboxIDPair(mbox *Mailbox) MailboxIDPair {
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

func NewMessageIDPair(msg *Message) MessageIDPair {
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

type MailboxFlag struct {
	ID    int
	Value string
}

type MailboxAttr struct {
	ID    int
	Value string
}

type Mailbox struct {
	ID             imap.InternalMailboxID
	RemoteID       imap.MailboxID
	Name           string
	UIDNext        imap.UID
	UIDValidity    imap.UID
	Subscribed     bool
	Flags          []*MailboxFlag
	PermanentFlags []*MailboxFlag
	Attributes     []*MailboxAttr
}

type MessageFlag struct {
	ID    int
	Value string
}

type Message struct {
	ID            imap.InternalMessageID
	RemoteID      imap.MessageID
	Date          time.Time
	Size          int
	Body          string
	BodyStructure string
	Envelope      string
	Deleted       bool
	Flags         []*MessageFlag
	UIDs          []*UID
}

type UID struct {
	UID     imap.UID
	Deleted bool
	Recent  bool
}

type DeletedSubscription struct {
	Name     string
	RemoteID imap.MailboxID
}
