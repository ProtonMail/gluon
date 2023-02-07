package imap

type Mailbox struct {
	ID MailboxID

	Name []string

	Flags, PermanentFlags, Attributes FlagSet
}

const Inbox = "INBOX"

type MailboxVisibility int

const (
	Hidden MailboxVisibility = iota
	Visible
	HiddenIfEmpty
)
