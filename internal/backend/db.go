package backend

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"entgo.io/ent/dialect"
	"github.com/ProtonMail/gluon/internal/backend/ent"
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
	d.lock.RLock()
	defer d.lock.Unlock()

	return fn(ctx, d.db)
}

func (d *DB) Write(ctx context.Context, fn func(context.Context, *ent.Tx) error) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	tx, err := d.db.Tx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if v := recover(); v != nil {
			if err := tx.Rollback(); err != nil {
				panic(fmt.Errorf("rolling back while recovering (%v): %w", v, err))
			}

			panic(v)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("rolling back transaction: %w", rerr)
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func getDatabasePath(dataPath, userID string) string {
	return fmt.Sprintf("file:%v?cache=shared&_fk=1", filepath.Join(dataPath, fmt.Sprintf("%v.db", userID)))
}

func NewDB(dataPath, userID string) (*DB, error) {
	dbPath := getDatabasePath(dataPath, userID)
	client, err := ent.Open(dialect.SQLite, dbPath)

	if err != nil {
		return nil, err
	}

	return &DB{db: client}, nil
}
