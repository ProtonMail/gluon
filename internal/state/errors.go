package state

import "errors"

var (
	ErrNoSuchMessage = errors.New("no such message")
	ErrNoSuchMailbox = errors.New("no such mailbox")

	ErrExistingMailbox     = errors.New("a mailbox with that name already exists")
	ErrAlreadySubscribed   = errors.New("already subscribed to this mailbox")
	ErrAlreadyUnsubscribed = errors.New("not subscribed to this mailbox")
	ErrSessionNotSelected  = errors.New("session is not selected")
)
