package store

type Store interface {
	Get(messageID string) ([]byte, error)
	Set(messageID string, literal []byte) error
	Update(oldID, newID string) error
	Delete(messageID ...string) error
	Close() error
}

type Builder interface {
	New(dir, userID string, passphrase []byte) (Store, error)
}
