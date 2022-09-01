package state

import (
	"fmt"

	"github.com/ProtonMail/gluon/imap"
)

type SnapFilter interface {
	Filter(s *State) bool
	String() string
}

type AllStateFilter struct{}

func (*AllStateFilter) Filter(s *State) bool {
	return s.snap != nil
}

func (*AllStateFilter) String() string {
	return "AllStates"
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

func (f *MBoxIDStateFilter) String() string {
	return fmt.Sprintf("mbox = %v", f.MboxID.ShortID())
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

func (f *MessageIDStateFilter) String() string {
	return fmt.Sprintf("message = %v", f.MessageID.ShortID())
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

func (f *MessageAndMBoxIDStateFilter) String() string {
	return fmt.Sprintf("mbox = %v message = %v", f.MBoxID.ShortID(), f.MessageID.ShortID())
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
