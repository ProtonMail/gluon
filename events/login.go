package events

type EventLogin struct {
	SessionID int
	UserID    string
}

func (EventLogin) _isEvent() {}
