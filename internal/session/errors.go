package session

import (
	"context"
	"errors"
	"net"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/rfc822"
)

var (
	ErrCreateInbox = errors.New("cannot create INBOX")
	ErrDeleteInbox = errors.New("cannot delete INBOX")
	ErrReadOnly    = errors.New("the mailbox is read-only")

	ErrTLSUnavailable       = errors.New("TLS is unavailable")
	ErrNotAuthenticated     = errors.New("session is not authenticated")
	ErrAlreadyAuthenticated = errors.New("session is already authenticated")

	ErrNotImplemented = errors.New("not implemented")
)

func shouldReportIMAPCommandError(err error) bool {
	var netErr *net.OpError

	switch {
	case errors.Is(err, ErrCreateInbox) || errors.Is(err, ErrDeleteInbox) || errors.Is(err, ErrReadOnly):
		return false
	case state.IsStateError(err):
		return false
	case errors.Is(err, connector.ErrOperationNotAllowed):
		return false
	case errors.Is(err, context.Canceled):
		return false
	case errors.As(err, &netErr):
		return false
	case errors.Is(err, rfc822.ErrNoSuchPart):
		return false
	case errors.Is(err, state.ErrKnownRecoveredMessage):
		return false
	}

	return true
}
