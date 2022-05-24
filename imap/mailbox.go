package imap

type Mailbox struct {
	ID string

	Name []string

	Flags, PermanentFlags, Attributes FlagSet
}

const Inbox = "INBOX"
