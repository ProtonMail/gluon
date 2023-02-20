package events

type LoginFailed struct {
	eventBase

	SessionID int
	Username  string
}
