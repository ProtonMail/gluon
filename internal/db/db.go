package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/ProtonMail/gluon/internal/utils"
	"github.com/google/uuid"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"entgo.io/ent/dialect"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/reporter"
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
	d.lock.Lock()
	defer d.lock.Unlock()

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

func getDatabaseConn(dir, userID, path string) string {
	return fmt.Sprintf("file:%v?cache=shared&_fk=1&_journal=WAL", path)
}

func getDatabasePath(dir, userID string) string {
	return filepath.Join(dir, fmt.Sprintf("%v.db", userID))
}

func GetDeferredDeleteDBPath(dir string) string {
	return filepath.Join(dir, "deferred_delete")
}

// NewDB creates a new database instance.
// If the database does not exist, it will be created and the second return value will be true.
func NewDB(dir, userID string) (*DB, bool, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, false, err
	}

	path := getDatabasePath(dir, userID)

	// Check if the database already exists.
	exists, err := pathExists(path)
	if err != nil {
		return nil, false, err
	}

	client, err := ent.Open(dialect.SQLite, getDatabaseConn(dir, userID, path))
	if err != nil {
		return nil, false, err
	}

	return &DB{db: client}, !exists, nil
}

// DeleteDB will rename all the database files for the given user to a directory within the same folder to avoid
// issues with ent not being able to close the database on demand. The database will be cleaned up on the next
// run on the Gluon server.
func DeleteDB(dir, userID string) error {
	// Rather than deleting the files immediately move them to a directory to be cleaned up later.
	deferredDeletePath := GetDeferredDeleteDBPath(dir)

	if err := os.MkdirAll(deferredDeletePath, 0o700); err != nil {
		return fmt.Errorf("failed to create deferred delete dir: %w", err)
	}

	matchingFiles, err := filepath.Glob(filepath.Join(dir, userID+"*"))
	if err != nil {
		return fmt.Errorf("failed to match db files:%w", err)
	}

	for _, file := range matchingFiles {
		// Use new UUID to avoid conflict with existing files
		if err := os.Rename(file, filepath.Join(deferredDeletePath, uuid.NewString())); err != nil {
			return fmt.Errorf("failed to move db file '%v' :%w", file, err)
		}
	}

	return nil
}

// DeleteDeferredDBFiles deletes all data from previous databases that were scheduled for removal.
func DeleteDeferredDBFiles(dir string) error {
	deferredDeleteDir := GetDeferredDeleteDBPath(dir)
	if err := os.RemoveAll(deferredDeleteDir); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	return nil
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
