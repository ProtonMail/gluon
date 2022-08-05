package store

type Store interface {
	Get(messageID string) ([]byte, error)
	Set(messageID string, literal []byte) error
	Update(oldID, newID string) error
	Delete(messageID ...string) error
	Close() error
}

type StoreBuilder interface {
	New(directory, userID, encryptionPassphrase string) (Store, error)
}
