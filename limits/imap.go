package limits

import (
	"errors"
	"fmt"
	"math"

	"github.com/ProtonMail/gluon/imap"
)

// IMAP contains configurable upper limits that can be enforced by the Gluon server.
type IMAP struct {
	maxMailboxCount           int
	maxMessageCountPerMailbox int
	maxUIDValidity            int
	maxUID                    int
}

func (i IMAP) CheckMailBoxCount(mailboxCount int) error {
	if mailboxCount >= i.maxMailboxCount {
		return ErrMaxMailboxCountReached
	}

	return nil
}

func (i IMAP) CheckMailBoxMessageCount(existingCount int, newCount int) error {
	nextMessageCount := existingCount + newCount

	if nextMessageCount > i.maxMessageCountPerMailbox || nextMessageCount < existingCount {
		return ErrMaxMailboxMessageCountReached
	}

	return nil
}

func (i IMAP) CheckUIDCount(existingUID imap.UID, newCount int) error {
	nextUIDCount := int(existingUID) + newCount

	if nextUIDCount > i.maxUID || nextUIDCount < int(existingUID) {
		return ErrMaxUIDReached
	}

	return nil
}

func (i IMAP) CheckUIDValidity(uid imap.UID) error {
	if int(uid) >= i.maxUIDValidity {
		return ErrMaxUIDValidityReached
	}

	return nil
}

func DefaultLimits() IMAP {
	return IMAP{
		maxMailboxCount:           math.MaxUint32,
		maxMessageCountPerMailbox: math.MaxUint32,
		maxUIDValidity:            math.MaxUint32,
		maxUID:                    math.MaxUint32,
	}
}

func NewIMAPLimits(maxMailboxCount uint32, maxMessageCount uint32, maxUID imap.UID, maxUIDValidity imap.UID) IMAP {
	return IMAP{
		maxMailboxCount:           int(maxMailboxCount),
		maxMessageCountPerMailbox: int(maxMessageCount),
		maxUIDValidity:            int(maxUIDValidity),
		maxUID:                    int(maxUID),
	}
}

var ErrMaxMailboxCountReached = fmt.Errorf("max mailbox count reached")
var ErrMaxMailboxMessageCountReached = fmt.Errorf("max mailbox message count reached")
var ErrMaxUIDReached = fmt.Errorf("max UID value reached")
var ErrMaxUIDValidityReached = fmt.Errorf("max UIDValidity value reached")

func IsIMAPLimitErr(err error) bool {
	return errors.Is(err, ErrMaxUIDValidityReached) ||
		errors.Is(err, ErrMaxMailboxCountReached) ||
		errors.Is(err, ErrMaxUIDReached) ||
		errors.Is(err, ErrMaxMailboxMessageCountReached)
}
