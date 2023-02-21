// Package gluon implements an IMAP4rev1 (+ extensions) mailserver.
package gluon

import (
	"errors"

	"github.com/ProtonMail/gluon/internal/state"
)

// IsNoSuchMessage returns true if the error is ErrNoSuchMessage.
func IsNoSuchMessage(err error) bool {
	return errors.Is(err, state.ErrNoSuchMessage)
}

// IsNoSuchMailbox returns true if the error is ErrNoSuchMailbox.
func IsNoSuchMailbox(err error) bool {
	return errors.Is(err, state.ErrNoSuchMailbox)
}
