package state

import "github.com/ProtonMail/gluon/imap"

type SnapFilter interface {
	Filter(s *State) bool
}

type AllStateFilter struct{}

func (*AllStateFilter) Filter(s *State) bool {
	return s.snap != nil
}

func NewAllStateFilter() SnapFilter {
	return &AllStateFilter{}
}

type MBoxIDStateFilter struct {
	MboxID imap.InternalMailboxID
}

func NewMBoxIDStateFilter(mboxID imap.InternalMailboxID) SnapFilter {
	return &MBoxIDStateFilter{MboxID: mboxID}
}

func (f *MBoxIDStateFilter) Filter(s *State) bool {
	return s.snap != nil && s.snap.mboxID.InternalID == f.MboxID
}

type MessageIDStateFilter struct {
	MessageID imap.InternalMessageID
}

func NewMessageIDStateFilter(msgID imap.InternalMessageID) SnapFilter {
	return &MessageIDStateFilter{MessageID: msgID}
}

func (f *MessageIDStateFilter) Filter(s *State) bool {
	return s.snap != nil && s.snap.hasMessage(f.MessageID)
}

type MessageAndMBoxIDStateFilter struct {
	MessageID imap.InternalMessageID
	MBoxID    imap.InternalMailboxID
}

func NewMessageAndMBoxIDStateFilter(msgID imap.InternalMessageID, mboxID imap.InternalMailboxID) SnapFilter {
	return &MessageAndMBoxIDStateFilter{MessageID: msgID, MBoxID: mboxID}
}

func (f *MessageAndMBoxIDStateFilter) Filter(s *State) bool {
	return s.snap != nil && s.snap.mboxID.InternalID == f.MBoxID && s.snap.hasMessage(f.MessageID)
}

type AnyMessageIDStateFilter struct {
	MessageIDs []imap.InternalMessageID
}

func (f *AnyMessageIDStateFilter) Filter(s *State) bool {
	if s.snap == nil {
		return false
	}

	for _, msgID := range f.MessageIDs {
		if s.snap.hasMessage(msgID) {
			return true
		}
	}

	return false
}
