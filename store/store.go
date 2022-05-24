package store

type Store interface {
	Get(messageID string) ([]byte, error)
	Set(messageID string, literal []byte) error
	Update(oldID, newID string) error
}
