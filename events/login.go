package events

type Login struct {
	eventBase

	SessionID int
	UserID    string
}
