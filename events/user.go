package events

import "github.com/ProtonMail/gluon/imap"

type UserAdded struct {
	eventBase

	UserID string

	Counts map[imap.MailboxID]int
}

type UserRemoved struct {
	eventBase

	UserID string
}
