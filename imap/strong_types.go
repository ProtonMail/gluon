package imap

import (
	"github.com/ProtonMail/gluon/internal/utils"
)

type LabelID string

type MessageID string

func (l *LabelID) ShortID() string {
	return utils.ShortID(string(*l))
}

func (m *MessageID) ShortID() string {
	return utils.ShortID(string(*m))
}

type InternalMessageID string

type InternalMailboxID string

func (i *InternalMailboxID) ShortID() string {
	return utils.ShortID(string(*i))
}

func (i *InternalMessageID) ShortID() string {
	return utils.ShortID(string(*i))
}
