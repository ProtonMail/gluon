package imap

type Mailbox struct {
	ID MailboxID

	Name []string

	Flags, PermanentFlags, Attributes FlagSet
}

type MailboxNoAttrib struct {
	ID MailboxID

	Name []string
}

const Inbox = "INBOX"

type MailboxVisibility int

const (
	Hidden MailboxVisibility = iota
	Visible
	HiddenIfEmpty
)

type MailboxData struct {
	InternalID InternalMailboxID
	RemoteID   string
	GluonName  string
	BridgeName []string
}
