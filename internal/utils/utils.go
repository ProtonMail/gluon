package utils

import (
	"errors"
	"github.com/google/uuid"
)

// NewRandomUserID return a new random user ID. For debugging purposes, the ID starts with the 'user-' prefix.
func NewRandomUserID() string {
	return "usr-" + uuid.NewString()
}

// NewRandomMailboxID return a new random mailbox ID. For debugging purposes, the ID starts with the 'lbl-' prefix.
func NewRandomMailboxID() string {
	return "lbl-" + uuid.NewString()
}

// NewRandomMessageID return a new random message ID. For debugging purposes, the ID starts with the 'message-' prefix.
func NewRandomMessageID() string {
	return "msg-" + uuid.NewString()
}

// ShortID return a string containing a short version of the given ID. Use only for debug display.
func ShortID(id string) string {
	const l = 12

	if len(id) < l {
		return id
	}

	return id[0:l] + "..."
}

// ErrCause returns the cause of the error, the inner-most error in the wrapped chain.
func ErrCause(err error) error {
	cause := err

	for errors.Unwrap(cause) != nil {
		cause = errors.Unwrap(cause)
	}

	return cause
}
