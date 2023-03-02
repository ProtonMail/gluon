package state

import "errors"

var (
	ErrNoSuchMessage = errors.New("no such message")
	ErrNoSuchMailbox = errors.New("no such mailbox")

	ErrExistingMailbox     = errors.New("a mailbox with that name already exists")
	ErrAlreadySubscribed   = errors.New("already subscribed to this mailbox")
	ErrAlreadyUnsubscribed = errors.New("not subscribed to this mailbox")
	ErrSessionNotSelected  = errors.New("session is not selected")

	ErrOperationNotAllowed            = errors.New("operation not allowed")
	ErrMailboxNameBeginsWithSeparator = errors.New("invalid mailbox name: begins with hierarchy separator")
	ErrMailboxNameAdjacentSeparator   = errors.New("invalid mailbox name: has adjacent hierarchy separators")
)

func IsStateError(err error) bool {
	return errors.Is(err, ErrNoSuchMailbox) ||
		errors.Is(err, ErrNoSuchMessage) ||
		errors.Is(err, ErrExistingMailbox) ||
		errors.Is(err, ErrAlreadySubscribed) ||
		errors.Is(err, ErrAlreadyUnsubscribed) ||
		errors.Is(err, ErrSessionNotSelected) ||
		errors.Is(err, ErrOperationNotAllowed) ||
		errors.Is(err, ErrMailboxNameBeginsWithSeparator) ||
		errors.Is(err, ErrMailboxNameAdjacentSeparator)
}
