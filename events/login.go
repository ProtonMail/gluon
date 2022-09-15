package events

type EventLogin struct {
	eventBase

	SessionID int
	UserID    string
}
