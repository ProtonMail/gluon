package imap

type Update interface {
	Waiter

	String() string

	_isUpdate()
}

type updateBase struct{}

func (updateBase) _isUpdate() {}
