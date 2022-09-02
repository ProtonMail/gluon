package store

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap"
)

type Store interface {
	Get(messageID imap.InternalMessageID) ([]byte, error)
	NewTransaction() Transaction
	Close() error
}

type Transaction interface {
	Set(messageID imap.InternalMessageID, literal []byte) error
	Delete(messageID ...imap.InternalMessageID) error
	Commit() error
	Rollback() error
}

type Builder interface {
	New(dir, userID string, passphrase []byte) (Store, error)
}

func Tx(store Store, fn func(Transaction) error) error {
	_, err := TxResult(store, func(tx Transaction) (struct{}, error) {
		if err := fn(tx); err != nil {
			return struct{}{}, err
		}

		return struct{}{}, nil
	})

	return err
}

func TxResult[T any](store Store, fn func(Transaction) (T, error)) (T, error) {
	tx := store.NewTransaction()

	var errResult T

	result, err := fn(tx)
	if err != nil {
		if te := tx.Rollback(); te != nil {
			return errResult, fmt.Errorf("failed to rollback transaction:%v - original error: %w", te, err)
		}

		return errResult, err
	}

	if err := tx.Commit(); err != nil {
		if te := tx.Rollback(); te != nil {
			return errResult, fmt.Errorf("failed to rollback transaction:%v - original error: %w", te, err)
		}

		return errResult, err
	}

	return result, nil
}
