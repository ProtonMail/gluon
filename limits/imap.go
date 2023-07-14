package limits

import (
	"errors"
	"fmt"
	"math"
	"math/bits"

	"github.com/ProtonMail/gluon/imap"
)

// IMAP contains configurable upper limits that can be enforced by the Gluon server.
type IMAP struct {
	maxMailboxCount           int64
	maxMessageCountPerMailbox int64
	maxUIDValidity            int64
	maxUID                    int64
}

func (i IMAP) CheckMailBoxCount(mailboxCount int) error {
	if int64(mailboxCount) >= i.maxMailboxCount {
		return ErrMaxMailboxCountReached
	}

	return nil
}

func (i IMAP) CheckMailBoxMessageCount(existingCount int, newCount int) error {
	nextMessageCount := int64(existingCount) + int64(newCount)

	if nextMessageCount > i.maxMessageCountPerMailbox || nextMessageCount < int64(existingCount) {
		return ErrMaxMailboxMessageCountReached
	}

	return nil
}

func (i IMAP) CheckUIDCount(existingUID imap.UID, newCount int) error {
	nextUIDCount := int64(existingUID) + int64(newCount)

	if nextUIDCount > i.maxUID || nextUIDCount < int64(existingUID) {
		return ErrMaxUIDReached
	}

	return nil
}

func (i IMAP) CheckUIDValidity(uid imap.UID) error {
	if int64(uid) >= i.maxUIDValidity {
		return ErrMaxUIDValidityReached
	}

	return nil
}

func DefaultLimits() IMAP {
	var maxInt int64
	if bits.UintSize == 64 {
		maxInt = math.MaxUint32
	} else {
		maxInt = math.MaxInt32
	}

	return IMAP{
		maxMailboxCount:           maxInt,
		maxMessageCountPerMailbox: maxInt,
		maxUIDValidity:            maxInt,
		maxUID:                    maxInt,
	}
}

func NewIMAPLimits(maxMailboxCount uint32, maxMessageCount uint32, maxUID imap.UID, maxUIDValidity imap.UID) IMAP {
	return IMAP{
		maxMailboxCount:           int64(maxMailboxCount),
		maxMessageCountPerMailbox: int64(maxMessageCount),
		maxUIDValidity:            int64(maxUIDValidity),
		maxUID:                    int64(maxUID),
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
