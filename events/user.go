package events

type EventUserAdded struct {
	UserID   string
	Username string
}

func (EventUserAdded) _isEvent() {}

type EventUserRemoved struct {
	UserID string
}

func (EventUserRemoved) _isEvent() {}
