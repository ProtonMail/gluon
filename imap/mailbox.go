package imap

type Mailbox struct {
	ID LabelID

	Name []string

	Flags, PermanentFlags, Attributes FlagSet
}

const Inbox = "INBOX"
