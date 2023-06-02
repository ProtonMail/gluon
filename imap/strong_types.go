package imap

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

type MailboxID string

type MessageID string

func (l MailboxID) ShortID() string {
	return ShortID(string(l))
}

func (m MessageID) ShortID() string {
	return ShortID(string(m))
}

type InternalMessageID struct {
	uuid.UUID
}

type InternalMailboxID uint64

func (i InternalMailboxID) ShortID() string {
	return strconv.FormatUint(uint64(i), 10)
}

func (i InternalMessageID) ShortID() string {
	return ShortID(i.String())
}

func (i InternalMailboxID) String() string {
	return strconv.FormatUint(uint64(i), 10)
}

func (i InternalMessageID) String() string {
	return i.UUID.String()
}

func NewInternalMessageID() InternalMessageID {
	return InternalMessageID{UUID: uuid.New()}
}

func InternalMessageIDFromString(id string) (InternalMessageID, error) {
	num, err := uuid.Parse(id)
	if err != nil {
		return InternalMessageID{}, fmt.Errorf("invalid message id:%w", err)
	}

	return InternalMessageID{UUID: num}, nil
}

type UID uint32

func (u UID) Add(v uint32) UID {
	return UID(uint32(u) + v)
}

type SeqID uint32
