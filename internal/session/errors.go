package session

import "errors"

var (
	ErrCreateInbox = errors.New("cannot create INBOX")
	ErrDeleteInbox = errors.New("cannot delete INBOX")
	ErrReadOnly    = errors.New("the mailbox is read-only")

	ErrTLSUnavailable       = errors.New("TLS is unavailable")
	ErrNotAuthenticated     = errors.New("session is not authenticated")
	ErrAlreadyAuthenticated = errors.New("session is already authenticated")

	ErrTooManyInvalid = errors.New("too many invalid IMAP commands")

	ErrNotImplemented = errors.New("not implemented")
)
