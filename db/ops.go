package db

type ReadOnly interface {
	MailboxReadOps
	MessageReadOps
	SubscriptionReadOps
}

type Transaction interface {
	ReadOnly
	MailboxWriteOps
	MessageWriteOps
	SubscriptionWriteOps
}
