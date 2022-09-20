package imap

import (
	"encoding/binary"
	"fmt"
	"github.com/ProtonMail/gluon/internal/utils"
	"strconv"
)

type LabelID string

type MessageID string

func (l LabelID) ShortID() string {
	return utils.ShortID(string(l))
}

func (m MessageID) ShortID() string {
	return utils.ShortID(string(m))
}

type InternalMessageID uint64

type InternalMailboxID uint64

func (i InternalMailboxID) ShortID() string {
	return strconv.FormatUint(uint64(i), 10)
}

func (i InternalMessageID) ShortID() string {
	return strconv.FormatUint(uint64(i), 10)
}

func (i InternalMailboxID) String() string {
	return strconv.FormatUint(uint64(i), 10)
}

func (i InternalMessageID) String() string {
	return strconv.FormatUint(uint64(i), 10)
}

func InternalMessageIDFromString(id string) (InternalMessageID, error) {
	num, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid message id:%w", err)
	}

	return InternalMessageID(num), nil
}

func (i InternalMessageID) ToBytes() []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(i))

	return bytes
}

type UID uint32

func (u UID) Add(v uint32) UID {
	return UID(uint32(u) + v)
}

type SeqID uint32
