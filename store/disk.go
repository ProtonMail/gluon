package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"os"
	"path/filepath"

	"github.com/ProtonMail/gluon/imap"
)

type onDiskStore struct {
	path string
	gcm  cipher.AEAD
	cmp  Compressor
	sem  *Semaphore
}

func NewOnDiskStore(path string, pass []byte, opt ...Option) (Store, error) {
	if err := os.MkdirAll(path, 0o700); err != nil {
		return nil, err
	}

	aes, err := aes.NewCipher(hash(pass))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, err
	}

	store := &onDiskStore{
		path: path,
		gcm:  gcm,
	}

	for _, opt := range opt {
		opt.config(store)
	}

	return store, nil
}

func (c *onDiskStore) Get(messageID imap.InternalMessageID) ([]byte, error) {
	if c.sem != nil {
		c.sem.Lock()
		defer c.sem.Unlock()
	}

	enc, err := os.ReadFile(filepath.Join(c.path, hashString(string(messageID))))
	if err != nil {
		return nil, err
	}

	b, err := c.gcm.Open(nil, enc[:c.gcm.NonceSize()], enc[c.gcm.NonceSize():], nil)
	if err != nil {
		return nil, err
	}

	if c.cmp != nil {
		dec, err := c.cmp.Decompress(b)
		if err != nil {
			return nil, err
		}

		b = dec
	}

	return b, nil
}

func (c *onDiskStore) Set(messageID imap.InternalMessageID, b []byte) error {
	if c.sem != nil {
		c.sem.Lock()
		defer c.sem.Unlock()
	}

	nonce := make([]byte, c.gcm.NonceSize())

	if _, err := rand.Read(nonce); err != nil {
		return err
	}

	if c.cmp != nil {
		enc, err := c.cmp.Compress(b)
		if err != nil {
			return err
		}

		b = enc
	}

	return os.WriteFile(
		filepath.Join(c.path, hashString(string(messageID))),
		c.gcm.Seal(nonce, nonce, b, nil),
		0o600,
	)
}

func (c *onDiskStore) Delete(ids ...imap.InternalMessageID) error {
	if c.sem != nil {
		c.sem.Lock()
		defer c.sem.Unlock()
	}

	for _, id := range ids {
		if err := os.RemoveAll(filepath.Join(c.path, hashString(string(id)))); err != nil {
			return err
		}
	}

	return nil
}

func (c *onDiskStore) Close() error {
	return nil
}

type OnDiskStoreBuilder struct{}

func (*OnDiskStoreBuilder) New(path, userID string, userPassword []byte) (Store, error) {
	storePath := filepath.Join(path, userID)

	return NewOnDiskStore(storePath, userPassword)
}
