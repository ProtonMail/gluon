package events

type EventSelect struct {
	eventBase

	SessionID int
	Mailbox   string
}
