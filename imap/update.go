package imap

type Update interface {
	Waiter

	String() string

	_isUpdate()
}
