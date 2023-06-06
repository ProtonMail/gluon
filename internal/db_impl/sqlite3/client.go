package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/internal/utils"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

type Client struct {
	db    *sql.DB
	lock  sync.RWMutex
	debug bool
	trace bool
}

func (c *Client) Init(ctx context.Context) error {
	if _, err := c.db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable db pragma: %w", err)
	}

	if _, err := c.db.ExecContext(ctx, "PRAGMA journal_mode = WAL"); err != nil {
		return fmt.Errorf("failed to enable db pragma: %w", err)
	}

	if _, err := c.db.ExecContext(ctx, "PRAGMA journal_mode = WAL"); err != nil {
		return fmt.Errorf("failed to enable db pragma: %w", err)
	}

	return c.wrapTx(ctx, func(ctx context.Context, tx *sql.Tx, entry *logrus.Entry) error {
		entry.Debugf("Running database migrations")
		return RunMigrations(ctx, TXWrapper{tx: tx})
	})
}

func (c *Client) Read(ctx context.Context, op func(context.Context, db.ReadOnly) error) error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	rdID := uuid.NewString()

	if c.debug {
		logrus.Debugf("Begin Read %v", rdID)
		defer logrus.Debugf("End Read %v", rdID)
	}

	entry := logrus.WithField("rd", rdID)

	var qw QueryWrapper = &DBWrapper{
		db: c.db,
	}

	if c.debug {
		qw = &DebugQueryWrapper{
			entry: entry,
			qw:    qw,
		}
	}

	var ops db.ReadOnly = &readOps{qw: qw}

	if c.trace {
		ops = &ReadTracer{rd: ops, entry: entry}
	}

	if err := op(ctx, ops); err != nil {
		return err
	}

	return nil
}

func (c *Client) Write(ctx context.Context, op func(context.Context, db.Transaction) error) error {
	return c.wrapTx(ctx, func(ctx context.Context, tx *sql.Tx, entry *logrus.Entry) error {

		var qw QueryWrapper = &TXWrapper{
			tx: tx,
		}

		if c.debug {
			qw = &DebugQueryWrapper{
				qw:    qw,
				entry: entry,
			}
		}

		var transaction db.Transaction = &writeOps{
			readOps: readOps{
				qw: qw,
			},
			qw: qw,
		}

		if c.trace {
			transaction = &WriteTracer{tx: transaction, ReadTracer: ReadTracer{rd: transaction, entry: entry}}
		}

		return op(ctx, transaction)
	})
}

func (c *Client) wrapTx(ctx context.Context, op func(context.Context, *sql.Tx, *logrus.Entry) error) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	var entry *logrus.Entry

	if c.debug {
		entry = logrus.WithField("tx", uuid.NewString())
	} else {
		entry = logrus.WithField("tx", "tx")
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if c.debug {
		entry.Debugf("Begin Transaction")
	}

	defer func() {
		if v := recover(); v != nil {
			if c.debug {
				entry.Debugf("Panic during Transaction")
			}

			if err := tx.Rollback(); err != nil {
				panic(fmt.Errorf("rolling back while recovering (%v): %w", v, err))
			}

			panic(v)
		}
	}()

	if err := op(ctx, tx, entry); err != nil {
		if c.debug {
			entry.Debugf("Rolling back Transaction")
		}

		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("rolling back transaction: %w", rerr)
		}

		return err
	}

	if err := tx.Commit(); err != nil {
		if !errors.Is(err, context.Canceled) {
			reporter.MessageWithContext(ctx,
				"Failed to commit database transaction",
				reporter.Context{"error": err, "type": utils.ErrCause(err)},
			)
		}

		if c.debug {
			entry.Debugf("Failed to commit Transaction")
		}

		return fmt.Errorf("%v: %w", err, db.ErrTransactionFailed)
	}

	if c.debug {
		entry.Debugf("Transaction Committed")
	}

	return nil
}

func (c *Client) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.db.Close()
}

type Builder struct {
	debug bool
	trace bool
}

type Option interface {
	apply(builder *Builder)
}

type dbDebugOption struct{}

func (dbDebugOption) apply(builder *Builder) {
	builder.debug = true
}

type dbTraceOption struct{}

func (dbTraceOption) apply(builder *Builder) {
	builder.trace = true
}

// Trace enables db interface call tracing. Name of the called functions will be written to trace log.
func Trace() Option {
	return &dbTraceOption{}
}

// Debug enables logging of the SQL queries and their values. Written to debug log.
func Debug() Option {
	return &dbDebugOption{}
}

func NewBuilder(options ...Option) db.ClientInterface {
	builder := &Builder{
		debug: false,
		trace: false,
	}

	for _, opt := range options {
		opt.apply(builder)
	}

	return builder
}

func (b Builder) New(dir string, userID string) (db.Client, bool, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, false, err
	}

	path := getDatabasePath(dir, userID)

	// Check if the database already exists.
	exists, err := pathExists(path)
	if err != nil {
		return nil, false, err
	}

	client, err := sql.Open("sqlite3", getDatabaseConn(dir, userID, path))
	if err != nil {
		return nil, false, err
	}

	return &Client{db: client, debug: b.debug, trace: b.trace}, !exists, nil
}

func (Builder) Delete(dir string, userID string) error {
	return db.DeleteDB(dir, userID)
}

func getDatabasePath(dir, userID string) string {
	return filepath.Join(dir, fmt.Sprintf("%v.db", userID))
}

func pathExists(path string) (bool, error) {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func getDatabaseConn(dir, userID, path string) string {
	return fmt.Sprintf("file:%v?cache=shared&_fk=1&_journal=WAL", path)
}
