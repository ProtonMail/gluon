package store

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/hash"
	"github.com/dgraph-io/badger/v3"
	"github.com/sirupsen/logrus"
)

type BadgerStore struct {
	db     *badger.DB
	stopCh chan struct{}
	stopWG sync.WaitGroup
}

type badgerTransaction struct {
	tx *badger.Txn
}

func NewBadgerStore(path string, userID string, passphrase []byte) (*BadgerStore, error) {
	db, err := badger.Open(badger.DefaultOptions(filepath.Join(path, userID)).
		WithLogger(logrus.StandardLogger()).
		WithLoggingLevel(badger.ERROR).
		WithEncryptionKey(hash.SHA256(passphrase)).
		WithIndexCacheSize(128 * 1024 * 1024),
	)
	if err != nil {
		return nil, err
	}

	store := &BadgerStore{
		db:     db,
		stopCh: make(chan struct{}),
	}

	store.startGCCollector()

	return store, nil
}

func (b *BadgerStore) startGCCollector() {
	// Garbage collection needs to be run manually by us at some point.
	// See https://dgraph.io/docs/badger/get-started/#garbage-collection for more details.
	b.stopWG.Add(1)

	go func() {
		defer b.stopWG.Done()

		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				{
				again:
					if err := b.db.RunValueLogGC(0.5); err == nil {
						goto again
					}
				}

			case <-b.stopCh:
				return
			}
		}
	}()
}

func (b *BadgerStore) Get(messageID imap.InternalMessageID) ([]byte, error) {
	var data []byte

	if err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(messageID.ToBytes())
		if err != nil {
			return err
		}

		data, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return data, nil
}

func (b *BadgerStore) NewTransaction() Transaction {
	return &badgerTransaction{tx: b.db.NewTransaction(true)}
}

func (b *badgerTransaction) Set(messageID imap.InternalMessageID, literal []byte) error {
	return b.tx.Set(messageID.ToBytes(), literal)
}

func (b *badgerTransaction) Delete(messageID ...imap.InternalMessageID) error {
	for _, v := range messageID {
		if err := b.tx.Delete(v.ToBytes()); err != nil {
			return err
		}
	}

	return nil
}

func (b *badgerTransaction) Commit() error {
	return b.tx.Commit()
}

func (b *badgerTransaction) Rollback() error {
	b.tx.Discard()

	return nil
}

func (b *BadgerStore) Close() error {
	close(b.stopCh)

	b.stopWG.Wait()

	return b.db.Close()
}

type BadgerStoreBuilder struct{}

func (*BadgerStoreBuilder) New(directory, userID string, encryptionPassphrase []byte) (Store, error) {
	return NewBadgerStore(directory, userID, encryptionPassphrase)
}

func (*BadgerStoreBuilder) Delete(directory, userID string) error {
	return os.RemoveAll(filepath.Join(directory, userID))
}
