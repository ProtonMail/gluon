package tests

import (
	"context"
	"github.com/ProtonMail/gluon/internal/db_impl"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFailedMigrationRestsDatabase(t *testing.T) {
	dbDir := t.TempDir()
	serverOptions := defaultServerOptions(t, withDatabaseDir(dbDir))

	var userID string

	runServer(t, serverOptions, func(session *testSession) {
		userID = session.userIDs["user"]
	})

	require.NoError(t, db_impl.TestUpdateDBVersion(context.Background(), dbDir, userID, 99999))

	runServer(t, serverOptions, func(session *testSession) {})
}
