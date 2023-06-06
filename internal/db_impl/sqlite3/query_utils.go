package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/db"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
)

// Collection of SQL utilities to process SQL Rows and to convert SQL errors to db.Errors.

type RowScanner interface {
	Scan(args ...any) error
}

func MapStmtRowsFn[T any](ctx context.Context, qw StmtWrapper, m func(RowScanner) (T, error), args ...any) ([]T, error) {
	rows, err := qw.QueryContext(ctx, args...)
	if err != nil {
		return nil, mapSQLError(err)
	}

	return mapSQLRowsFn(rows, m)
}

func MapStmtRows[T any](ctx context.Context, qw StmtWrapper, args ...any) ([]T, error) {
	return MapStmtRowsFn(ctx, qw, func(scanner RowScanner) (T, error) {
		var v T

		err := scanner.Scan(&v)

		return v, err
	}, args...)
}

func MapStmtRowFn[T any](ctx context.Context, qw StmtWrapper, m func(RowScanner) (T, error), args ...any) (T, error) {
	rows := qw.QueryRowContext(ctx, args...)

	return mapSQLRowFn(rows, m)
}

func MapStmtRow[T any](ctx context.Context, qw StmtWrapper, args ...any) (T, error) {
	return MapStmtRowFn(ctx, qw, func(scanner RowScanner) (T, error) {
		var v T

		err := scanner.Scan(&v)

		return v, err
	}, args...)
}

func MapQueryRowsFn[T any](ctx context.Context, qw QueryWrapper, query string, m func(RowScanner) (T, error), args ...any) ([]T, error) {
	rows, err := qw.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, mapSQLError(err)
	}

	return mapSQLRowsFn(rows, m)
}

func MapQueryRows[T any](ctx context.Context, qw QueryWrapper, query string, args ...any) ([]T, error) {
	return MapQueryRowsFn(ctx, qw, query, func(scanner RowScanner) (T, error) {
		var v T

		err := scanner.Scan(&v)

		return v, err
	}, args...)
}

func MapQueryRowFn[T any](ctx context.Context, qw QueryWrapper, query string, m func(RowScanner) (T, error), args ...any) (T, error) {
	row := qw.QueryRowContext(ctx, query, args...)

	return mapSQLRowFn(row, m)
}

func MapQueryRow[T any](ctx context.Context, qw QueryWrapper, query string, args ...any) (T, error) {
	return MapQueryRowFn(ctx, qw, query, func(scanner RowScanner) (T, error) {
		var v T

		err := scanner.Scan(&v)

		return v, err
	}, args...)
}

func ExecQueryAndCheckUpdatedNotZero(ctx context.Context, wrapper QueryWrapper, query string, args ...any) error {
	updated, err := ExecQuery(ctx, wrapper, query, args...)
	if err != nil {
		return err
	}

	if updated == 0 {
		return fmt.Errorf("no values changed")
	}

	return nil
}

func ExecStmtAndCheckUpdatedNotZero(ctx context.Context, wrapper StmtWrapper, args ...any) error {
	updated, err := ExecStmt(ctx, wrapper, args...)
	if err != nil {
		return err
	}

	if updated == 0 {
		return fmt.Errorf("no values changed")
	}

	return nil
}

func ExecQuery(ctx context.Context, wrapper QueryWrapper, query string, args ...any) (int, error) {
	r, err := wrapper.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := r.RowsAffected()
	if err != nil {
		panic("affected rows is unsupported")
	}

	return int(affected), nil
}

func ExecStmt(ctx context.Context, wrapper StmtWrapper, args ...any) (int, error) {
	r, err := wrapper.ExecContext(ctx, args...)
	if err != nil {
		return 0, err
	}

	affected, err := r.RowsAffected()
	if err != nil {
		panic("affected rows is unsupported")
	}

	return int(affected), nil
}

func GenSQLIn(count int) string {
	if count <= 0 {
		panic("count can't be less or equal to 0")
	}

	if count == 1 {
		return "?"
	}

	return strings.Repeat("?,", count-1) + "?"
}

func MapSliceToAny[T any](v []T) []any {
	return xslices.Map(v, func(t T) any {
		return t
	})
}

func QueryExists(ctx context.Context, qw QueryWrapper, query string, args ...any) (bool, error) {
	if _, err := MapQueryRow[int](ctx, qw, query, args...); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func WrapStmtClose(st StmtWrapper) {
	if err := st.Close(); err != nil {
		logrus.WithError(err).Error("Failed to close statement")
	}
}

func mapSQLError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return db.ErrNotFound
	}

	return err
}

func mapSQLRowsFn[T any](rows *sql.Rows, m func(RowScanner) (T, error)) ([]T, error) {
	defer func() { _ = rows.Close() }()

	var result []T

	for rows.Next() {
		val, err := m(rows)
		if err != nil {
			return nil, err
		}

		result = append(result, val)
	}

	return result, nil
}

func mapSQLRowFn[T any](row *sql.Row, m func(scanner RowScanner) (T, error)) (T, error) {
	v, err := m(row)

	return v, mapSQLError(err)
}
