package db

import "errors"

var ErrNotFound = errors.New("value not found")
var ErrTransactionFailed = errors.New("transaction failed")

func IsErrNotFound(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, ErrNotFound)
}
