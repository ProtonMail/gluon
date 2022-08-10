package store

import (
	"path/filepath"
	"sync"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/sirupsen/logrus"
)

type BadgerStore struct {
	db       *badger.DB
	gcExitCh chan struct{}
	wg       sync.WaitGroup
}

func logrusLevelToBadgerLevel(options badger.Options) badger.Options {
	switch logrus.GetLevel() {
	case logrus.InfoLevel:
		return options.WithLoggingLevel(badger.INFO)
	case logrus.TraceLevel:
		fallthrough
	case logrus.DebugLevel:
		return options.WithLoggingLevel(badger.DEBUG)
	case logrus.WarnLevel:
		return options.WithLoggingLevel(badger.WARNING)
	case logrus.FatalLevel:
		fallthrough
	case logrus.PanicLevel:
		fallthrough
	case logrus.ErrorLevel:
		return options.WithLoggingLevel(badger.ERROR)
	default:
		return options.WithLoggingLevel(badger.ERROR)
	}
}

func NewBadgerStore(path string, userID string, encryptionPassphrase []byte) (*BadgerStore, error) {
	directory := filepath.Join(path, userID)
	db, err := badger.Open(logrusLevelToBadgerLevel(badger.DefaultOptions(directory)).
		WithLogger(logrus.StandardLogger()).
		WithEncryptionKey(encryptionPassphrase).
		WithIndexCacheSize(128 * 1024 * 1024))

	if err != nil {
		return nil, nil
	}

	store := &BadgerStore{db: db, gcExitCh: make(chan struct{})}

	store.startGCCollector()

	return store, nil
}

func (b *BadgerStore) startGCCollector() {
	// Garbage collection needs to be run manually by us at some point.
	// See https://dgraph.io/docs/badger/get-started/#garbage-collection for more details.
	b.wg.Add(1)

	go func() {
		defer b.wg.Done()

		gcRun := time.After(5 * time.Minute)

		select {
		case <-gcRun:
			{
			again:
				err := b.db.RunValueLogGC(0.7)
				if err == nil {
					goto again
				}
			}
		case <-b.gcExitCh:
			return
		}
	}()
}

func (b *BadgerStore) Get(messageID string) ([]byte, error) {
	var data []byte

	if err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(messageID))
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

func (b *BadgerStore) Set(messageID string, literal []byte) error {
	return b.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(messageID), literal)
	})
}

func (b *BadgerStore) Update(oldID, newID string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		oldIDBytes := []byte(oldID)
		newIDBytes := []byte(newID)

		item, err := txn.Get(oldIDBytes)
		if err != nil {
			return err
		}

		buffer := make([]byte, item.ValueSize())
		buffer, err = item.ValueCopy(buffer)
		if err != nil {
			return err
		}

		if err := txn.Set(newIDBytes, buffer); err != nil {
			return err
		}

		return txn.Delete(oldIDBytes)
	})
}

func (b *BadgerStore) Delete(messageID ...string) error {
	return b.db.Update(func(txn *badger.Txn) error {
		for _, v := range messageID {
			if err := txn.Delete([]byte(v)); err != nil {
				return err
			}
		}

		return nil
	})
}

func (b *BadgerStore) Close() error {
	close(b.gcExitCh)
	b.wg.Wait()

	return b.db.Close()
}

type BadgerStoreBuilder struct{}

func (*BadgerStoreBuilder) New(directory, userID, encryptionPassphrase string) (Store, error) {
	return NewBadgerStore(directory, userID, []byte(encryptionPassphrase))
}
