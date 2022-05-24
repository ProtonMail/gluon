package events

type EventSelect struct {
	SessionID int
	Mailbox   string
}

func (EventSelect) _isEvent() {}
