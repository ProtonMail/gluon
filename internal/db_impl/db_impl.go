package db_impl

import (
	"context"
	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3"
)

func NewSQLiteDB(options ...sqlite3.Option) db.ClientInterface {
	return sqlite3.NewBuilder(options...)
}

func TestUpdateDBVersion(ctx context.Context, dbPath, userID string, version int) error {
	return sqlite3.TestUpdateDBVersion(ctx, dbPath, userID, version)
}
