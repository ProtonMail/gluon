package events

type Select struct {
	eventBase

	SessionID int
	Mailbox   string
}
