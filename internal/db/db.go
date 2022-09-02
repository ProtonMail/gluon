package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"entgo.io/ent/dialect"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/store"
)

type DB struct {
	db   *ent.Client
	lock sync.RWMutex
}

func (d *DB) Init(ctx context.Context) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.db.Schema.Create(ctx)
}

func (d *DB) Read(ctx context.Context, fn func(context.Context, *ent.Client) error) error {
	_, err := ReadResult(ctx, d, func(ctx context.Context, client *ent.Client) (struct{}, error) {
		return struct{}{}, fn(ctx, client)
	})

	return err
}

func (d *DB) Write(ctx context.Context, fn func(context.Context, *ent.Tx) error) error {
	_, err := WriteResult(ctx, d, func(ctx context.Context, tx *ent.Tx) (struct{}, error) {
		return struct{}{}, fn(ctx, tx)
	})

	return err
}

func (d *DB) Close() error {
	return d.db.Close()
}

func ReadResult[T any](ctx context.Context, db *DB, fn func(context.Context, *ent.Client) (T, error)) (T, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	return fn(ctx, db.db)
}

func WriteResult[T any](ctx context.Context, db *DB, fn func(context.Context, *ent.Tx) (T, error)) (T, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	var failResult T

	tx, err := db.db.Tx(ctx)
	if err != nil {
		return failResult, err
	}

	defer func() {
		if v := recover(); v != nil {
			if err := tx.Rollback(); err != nil {
				panic(fmt.Errorf("rolling back while recovering (%v): %w", v, err))
			}

			panic(v)
		}
	}()

	result, err := fn(ctx, tx)
	if err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return failResult, fmt.Errorf("rolling back transaction: %w", rerr)
		}

		return failResult, err
	}

	if err := tx.Commit(); err != nil {
		reporter.MessageWithContext(ctx,
			"Failed to commit database transaction",
			reporter.Context{"error": err},
		)

		return failResult, fmt.Errorf("committing transaction: %w", err)
	}

	return result, nil
}

func getDatabasePath(dir, userID string) string {
	return fmt.Sprintf("file:%v?cache=shared&_fk=1", filepath.Join(dir, fmt.Sprintf("%v.db", userID)))
}

func NewDB(dir, userID string) (*DB, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}

	client, err := ent.Open(dialect.SQLite, getDatabasePath(dir, userID))
	if err != nil {
		return nil, err
	}

	return &DB{db: client}, nil
}

// WriteAndStore is the same as WriteStoreAndResult.
func WriteAndStore(ctx context.Context, db *DB, st store.Store, fn func(context.Context, *ent.Tx, store.Transaction) error) error {
	return db.Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
		return store.Tx(st, func(transaction store.Transaction) error {
			storeTxErr := fn(ctx, tx, transaction)

			if storeTxErr != nil {
				reporter.MessageWithContext(ctx,
					"Failed to commit storage transaction",
					reporter.Context{"error": storeTxErr},
				)
			}

			return storeTxErr
		})
	})
}

// WriteAndStoreResult wraps the two transactions from the SQL and storage databases. The store transaction is wrapped by
// the sql transaction. It is more important to guarantee that the SQL db is consistent, and we accept some unnecessary
// changes in the storage db, we can always recover from these more easily.
func WriteAndStoreResult[T any](ctx context.Context, db *DB, st store.Store, fn func(context.Context, *ent.Tx, store.Transaction) (T, error)) (T, error) {
	return WriteResult(ctx, db, func(ctx context.Context, tx *ent.Tx) (T, error) {
		val, storeTxErr := store.TxResult(st, func(transaction store.Transaction) (T, error) {
			return fn(ctx, tx, transaction)
		})

		if storeTxErr != nil {
			reporter.MessageWithContext(ctx,
				"Failed to commit storage transaction",
				reporter.Context{"error": storeTxErr},
			)
		}

		return val, storeTxErr
	})
}
