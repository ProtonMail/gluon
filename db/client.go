package db

import (
	"context"
	"path/filepath"

	"github.com/ProtonMail/gluon/imap"
)

const ChunkLimit = 1000

type Client interface {
	Init(ctx context.Context, generator imap.UIDValidityGenerator) error
	Read(ctx context.Context, op func(context.Context, ReadOnly) error) error
	Write(ctx context.Context, op func(context.Context, Transaction) error) error
	Close() error
}

type ClientInterface interface {
	New(path string, userID string) (Client, bool, error)
	Delete(path string, userID string) error
}

func GetDeferredDeleteDBPath(dir string) string {
	return filepath.Join(dir, "deferred_delete")
}

func ClientReadType[T any](ctx context.Context, c Client, op func(context.Context, ReadOnly) (T, error)) (T, error) {
	var result T

	err := c.Read(ctx, func(ctx context.Context, read ReadOnly) error {
		var err error

		result, err = op(ctx, read)

		return err
	})

	return result, err
}

func ClientWriteType[T any](ctx context.Context, c Client, op func(context.Context, Transaction) (T, error)) (T, error) {
	var result T

	err := c.Write(ctx, func(ctx context.Context, t Transaction) error {
		var err error

		result, err = op(ctx, t)

		return err
	})

	return result, err
}
