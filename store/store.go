package store

import "github.com/ProtonMail/gluon/imap"

type Store interface {
	Get(messageID imap.InternalMessageID) ([]byte, error)
	Set(messageID imap.InternalMessageID, literal []byte) error
	Delete(messageID ...imap.InternalMessageID) error
	Close() error
	List() ([]imap.InternalMessageID, error)
}

type Builder interface {
	New(dir, userID string, passphrase []byte) (Store, error)
	Delete(dir, userID string) error
}
