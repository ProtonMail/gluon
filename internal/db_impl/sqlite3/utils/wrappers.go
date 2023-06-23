package utils

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
	DB *sql.DB
}

func (d DBWrapper) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.DB.QueryContext(ctx, query, args...)
}

func (d DBWrapper) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.DB.QueryRowContext(ctx, query, args...)
}

func (d DBWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.DB.ExecContext(ctx, query, args...)
}

func (d DBWrapper) PrepareStatement(ctx context.Context, query string) (StmtWrapper, error) {
	return d.DB.PrepareContext(ctx, query)
}

type TXWrapper struct {
	TX *sql.Tx
}

func (t TXWrapper) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.TX.QueryContext(ctx, query, args...)
}

func (t TXWrapper) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.TX.QueryRowContext(ctx, query, args...)
}

func (t TXWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.TX.ExecContext(ctx, query, args...)
}

func (t TXWrapper) PrepareStatement(ctx context.Context, query string) (StmtWrapper, error) {
	return t.TX.PrepareContext(ctx, query)
}

type DebugQueryWrapper struct {
	QW    QueryWrapper
	Entry *logrus.Entry
}

func (d DebugQueryWrapper) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	d.Entry.Debugf("query=%v args=%v", query, args)

	return d.QW.QueryContext(ctx, query, args...)
}

func (d DebugQueryWrapper) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	d.Entry.Debugf("query=%v args=%v", query, args)

	return d.QW.QueryRowContext(ctx, query, args...)
}

func (d DebugQueryWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	d.Entry.Debugf("Exec=%v args=%v", query, args)

	return d.QW.ExecContext(ctx, query, args...)
}

func (d DebugQueryWrapper) PrepareStatement(ctx context.Context, query string) (StmtWrapper, error) {
	d.Entry.Debugf("Prepare Statement=%v ", query)

	stmt, err := d.QW.PrepareStatement(ctx, query)
	if err != nil {
		return nil, err
	}

	return &DebugStmtWrapper{
		sw:    stmt,
		entry: d.Entry,
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
