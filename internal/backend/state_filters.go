package backend

import "github.com/ProtonMail/gluon/imap"

type stateFilter interface {
	filter(s *State) bool
}

type allStateFilter struct{}

func (*allStateFilter) filter(s *State) bool {
	return s.snap != nil
}

func newAllStateFilter() stateFilter {
	return &allStateFilter{}
}

type mboxIDStateFilter struct {
	mboxID imap.InternalMailboxID
}

func newMBoxIDStateFilter(mboxID imap.InternalMailboxID) stateFilter {
	return &mboxIDStateFilter{mboxID: mboxID}
}

func (f *mboxIDStateFilter) filter(s *State) bool {
	return s.snap != nil && s.snap.mboxID.InternalID == f.mboxID
}

type messageIDStateFilter struct {
	messageID imap.InternalMessageID
}

func newMessageIDStateFilter(msgID imap.InternalMessageID) stateFilter {
	return &messageIDStateFilter{messageID: msgID}
}

func (f *messageIDStateFilter) filter(s *State) bool {
	return s.snap != nil && s.snap.hasMessage(f.messageID)
}

type messageAndMBoxIDStateFilter struct {
	msgID  imap.InternalMessageID
	mboxID imap.InternalMailboxID
}

func newMessageAndMBoxIDStateFilter(msgID imap.InternalMessageID, mboxID imap.InternalMailboxID) stateFilter {
	return &messageAndMBoxIDStateFilter{msgID: msgID, mboxID: mboxID}
}

func (f *messageAndMBoxIDStateFilter) filter(s *State) bool {
	return s.snap != nil && s.snap.mboxID.InternalID == f.mboxID && s.snap.hasMessage(f.msgID)
}

type anyMessageIDStateFilter struct {
	messageIDs []imap.InternalMessageID
}

func (f *anyMessageIDStateFilter) filter(s *State) bool {
	if s.snap == nil {
		return false
	}

	for _, msgID := range f.messageIDs {
		if s.snap.hasMessage(msgID) {
			return true
		}
	}

	return false
}
