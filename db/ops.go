package db

type ReadOnly interface {
	MailboxReadOps
	MessageReadOps
	SubscriptionReadOps
}

type Transaction interface {
	MailboxWriteOps
	MessageWriteOps
	SubscriptionWriteOps
}
