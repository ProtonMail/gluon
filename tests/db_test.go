package tests

import (
	"context"
	"github.com/ProtonMail/gluon/internal/db_impl"
	"os"
	"testing"

	"github.com/ProtonMail/gluon/db"
	"github.com/stretchr/testify/require"
)

func dbCheckUserMessageCount(s *testSession, user string, expectedCount int) {
	err := s.withUserDB(user, func(ent db.Client, ctx context.Context) {
		val, err := db.ClientReadType(ctx, ent, func(ctx context.Context, only db.ReadOnly) (int, error) {
			return only.GetTotalMessageCount(ctx)
		})
		require.NoError(s.tb, err)
		require.Equal(s.tb, expectedCount, val)
	})
	require.NoError(s.tb, err)
}

func TestRunEntThenSqlite3(t *testing.T) {
	if _, ok := os.LookupEnv("GLUON_TEST_FORCE_ENT_DB"); ok {
		t.Skip("Does not make sense to run this test under these conditions")
	}

	dataDir := t.TempDir()

	dbDir := t.TempDir()

	// Run once with Ent DB.
	runServer(t, defaultServerOptions(t, withDatabase(db_impl.NewEntDB()), withDatabaseDir(dbDir), withDataDir(dataDir)), func(session *testSession) {

	})

	// Run once with SQLite DB.
	runServer(t, defaultServerOptions(t, withDatabase(db_impl.NewSQLiteDB()), withDatabaseDir(dbDir), withDataDir(dataDir)), func(session *testSession) {
	})

	// Run second time with SQLite DB.
	runServer(t, defaultServerOptions(t, withDatabase(db_impl.NewSQLiteDB()), withDatabaseDir(dbDir), withDataDir(dataDir)), func(session *testSession) {
	})
}
