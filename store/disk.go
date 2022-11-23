package store

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ProtonMail/gluon/imap"
	"github.com/sirupsen/logrus"
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

	enc, err := os.ReadFile(filepath.Join(c.path, messageID.String()))
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
		filepath.Join(c.path, messageID.String()),
		c.gcm.Seal(nonce, nonce, b, nil),
		0o600,
	)
}

func (c *onDiskStore) Delete(messageIDs ...imap.InternalMessageID) error {
	if c.sem != nil {
		c.sem.Lock()
		defer c.sem.Unlock()
	}

	for _, messageID := range messageIDs {
		if err := os.RemoveAll(filepath.Join(c.path, messageID.String())); err != nil {
			return err
		}
	}

	return nil
}

func (c *onDiskStore) List() ([]imap.InternalMessageID, error) {
	if c.sem != nil {
		c.sem.Lock()
		defer c.sem.Unlock()
	}

	var ids []imap.InternalMessageID

	if err := filepath.Walk(c.path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		id, err := imap.InternalMessageIDFromString(info.Name())
		if err != nil {
			logrus.WithError(err).Errorf("Invalid id file in cache: %v", info.Name())
		}

		ids = append(ids, id)

		return nil
	}); err != nil {
		return nil, err
	}

	return ids, nil
}

func (c *onDiskStore) Close() error {
	return nil
}

type OnDiskStoreBuilder struct{}

func (*OnDiskStoreBuilder) New(path, userID string, passphrase []byte) (Store, error) {
	storePath := filepath.Join(path, userID)

	return NewOnDiskStore(storePath, passphrase)
}

func (*OnDiskStoreBuilder) Delete(path, userID string) error {
	storePath := filepath.Join(path, userID)

	return os.RemoveAll(storePath)
}
