package sqlite3

import (
	"context"
	"database/sql"

	"github.com/sirupsen/logrus"
)

// Collection of wrappers to help with tracing and debugging of SQL queries and statements.

// QueryWrapper is a wrapper around go's sql.DB and sql.Tx types so we can override the calls with trackers (e.g.:
// DebugQueryWrapper).
type QueryWrapper interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareStatement(ctx context.Context, query string) (StmtWrapper, error)
}

// StmtWrapper is a wrapper around go's sql.Stmt type so we can override the calls with trackers (e.g.:
// DebugStmtWrapper).
type StmtWrapper interface {
	QueryContext(ctx context.Context, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, args ...any) *sql.Row
	ExecContext(ctx context.Context, args ...any) (sql.Result, error)
	Close() error
}

type DBWrapper struct {
	db *sql.DB
}

func (d DBWrapper) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

func (d DBWrapper) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

func (d DBWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

func (d DBWrapper) PrepareStatement(ctx context.Context, query string) (StmtWrapper, error) {
	return d.db.PrepareContext(ctx, query)
}

type TXWrapper struct {
	tx *sql.Tx
}

func (t TXWrapper) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t TXWrapper) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t TXWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t TXWrapper) PrepareStatement(ctx context.Context, query string) (StmtWrapper, error) {
	return t.tx.PrepareContext(ctx, query)
}

type DebugQueryWrapper struct {
	qw    QueryWrapper
	entry *logrus.Entry
}

func (d DebugQueryWrapper) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	d.entry.Debugf("query=%v args=%v", query, args)

	return d.qw.QueryContext(ctx, query, args...)
}

func (d DebugQueryWrapper) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	d.entry.Debugf("query=%v args=%v", query, args)

	return d.qw.QueryRowContext(ctx, query, args...)
}

func (d DebugQueryWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	d.entry.Debugf("Exec=%v args=%v", query, args)

	return d.qw.ExecContext(ctx, query, args...)
}

func (d DebugQueryWrapper) PrepareStatement(ctx context.Context, query string) (StmtWrapper, error) {
	stmt, err := d.qw.PrepareStatement(ctx, query)
	if err != nil {
		return nil, err
	}

	return &DebugStmtWrapper{
		sw:    stmt,
		entry: d.entry,
		query: query,
	}, nil
}

type DebugStmtWrapper struct {
	sw    StmtWrapper
	entry *logrus.Entry
	query string
}

func (d DebugStmtWrapper) QueryContext(ctx context.Context, args ...any) (*sql.Rows, error) {
	d.entry.Debugf("query=%v args=%v", d.query, args)

	return d.sw.QueryContext(ctx, args...)
}

func (d DebugStmtWrapper) QueryRowContext(ctx context.Context, args ...any) *sql.Row {
	d.entry.Debugf("query=%v args=%v", d.query, args)

	return d.sw.QueryRowContext(ctx, args...)
}

func (d DebugStmtWrapper) ExecContext(ctx context.Context, args ...any) (sql.Result, error) {
	d.entry.Debugf("query=%v args=%v", d.query, args)

	return d.sw.ExecContext(ctx, args...)
}

func (d DebugStmtWrapper) Close() error {
	return d.sw.Close()
}
