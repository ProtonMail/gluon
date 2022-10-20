package imap

type Mailbox struct {
	ID MailboxID

	Name []string

	Flags, PermanentFlags, Attributes FlagSet
}

const Inbox = "INBOX"
