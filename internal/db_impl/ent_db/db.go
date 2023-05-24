package ent_db

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"entgo.io/ent/dialect"
	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/internal/db_impl/ent_db/internal"
	"github.com/ProtonMail/gluon/internal/utils"
	"github.com/ProtonMail/gluon/reporter"
)

type DB struct {
	db   *internal.Client
	lock sync.RWMutex
}

func (d *DB) Init(ctx context.Context) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.db.Schema.Create(ctx)
}

func (d *DB) ReadEnt(ctx context.Context, fn func(context.Context, *internal.Client) error) error {
	_, err := ReadResult(ctx, d, func(ctx context.Context, client *internal.Client) (struct{}, error) {
		return struct{}{}, fn(ctx, client)
	})

	return err
}

func (d *DB) WriteEnt(ctx context.Context, fn func(context.Context, *internal.Tx) error) error {
	_, err := WriteResult(ctx, d, func(ctx context.Context, tx *internal.Tx) (struct{}, error) {
		return struct{}{}, fn(ctx, tx)
	})

	return err
}

func (d *DB) Read(ctx context.Context, fn func(context.Context, db.ReadOnly) error) error {
	return d.ReadEnt(ctx, func(ctx context.Context, client *internal.Client) error {
		rd := newOpsReadFromClient(client)
		return fn(ctx, rd)
	})
}

func (d *DB) Write(ctx context.Context, fn func(context.Context, db.Transaction) error) error {
	return d.WriteEnt(ctx, func(ctx context.Context, tx *internal.Tx) error {
		op := newEntOpsWrite(tx)
		return fn(ctx, op)
	})
}

func (d *DB) Close() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.db.Close()
}

func ReadResult[T any](ctx context.Context, db *DB, fn func(context.Context, *internal.Client) (T, error)) (T, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	return fn(ctx, db.db)
}

func WriteResult[T any](ctx context.Context, db *DB, fn func(context.Context, *internal.Tx) (T, error)) (T, error) {
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
		if !errors.Is(err, context.Canceled) {
			reporter.MessageWithContext(ctx,
				"Failed to commit database transaction",
				reporter.Context{"error": err, "type": utils.ErrCause(err)},
			)
		}

		return failResult, fmt.Errorf("committing transaction: %w", err)
	}

	return result, nil
}

type EntDBBuilder struct{}

func NewEntDBBuilder() db.ClientInterface {
	return &EntDBBuilder{}
}

func (EntDBBuilder) New(dir string, userID string) (db.Client, bool, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, false, err
	}

	path := getDatabasePath(dir, userID)

	// Check if the database already exists.
	exists, err := pathExists(path)
	if err != nil {
		return nil, false, err
	}

	client, err := internal.Open(dialect.SQLite, getDatabaseConn(dir, userID, path))
	if err != nil {
		return nil, false, err
	}

	return &DB{db: client}, !exists, nil
}

func (EntDBBuilder) Delete(dir string, userID string) error {
	return db.DeleteDB(dir, userID)
}

func getDatabaseConn(dir, userID, path string) string {
	return fmt.Sprintf("file:%v?cache=shared&_fk=1&_journal=WAL", path)
}

// pathExists returns whether the given file exists.
func pathExists(path string) (bool, error) {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func getDatabasePath(dir, userID string) string {
	return filepath.Join(dir, fmt.Sprintf("%v.db", userID))
}

func NewEntDB() db.ClientInterface {
	return &EntDBBuilder{}
}

func wrapEntError(err error) error {
	if err == nil {
		return nil
	}

	if internal.IsNotFound(err) {
		return fmt.Errorf("%v (%w)", err, db.ErrNotFound)
	}

	return err
}

func wrapEntErrFn(fn func() error) error {
	return wrapEntError(fn())
}

func wrapEntErrFnTyped[T any](fn func() (T, error)) (T, error) {
	val, err := fn()
	return val, wrapEntError(err)
}
