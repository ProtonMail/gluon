package tests

import (
	"errors"
	"github.com/ProtonMail/gluon/db"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestAccountRemovalMovesDBToDeferredDeleteFolder(t *testing.T) {
	dataDir := t.TempDir()
	dbDir := t.TempDir()

	dbDeleteDirPath := db.GetDeferredDeleteDBPath(dbDir)

	// Deferred delete folder does not exist.
	empty, err := isEmptyFolder(dbDeleteDirPath)
	require.NoError(t, err)
	require.True(t, empty)

	runServer(t, defaultServerOptions(t, withDataDir(dataDir), withDatabaseDir(dbDir)), func(session *testSession) {
		userID := session.removeAccount(t, "user")

		// Remove account's DB was moved to this folder.
		empty, err := isEmptyFolder(dbDeleteDirPath)
		require.NoError(t, err)
		require.False(t, empty)

		// DB dir should not have user id files anymore.
		matchingFiles, err := filepath.Glob(filepath.Join(dbDir, userID+"*"))
		require.NoError(t, err)
		require.Empty(t, matchingFiles)
	})

	// Database will be deleted on server startup.
	runServer(t, defaultServerOptions(t, withDataDir(dataDir), withDatabaseDir(dbDir)), func(session *testSession) {
		empty, err := isEmptyFolder(dbDeleteDirPath)
		require.NoError(t, err)
		require.True(t, empty)
	})
}

func isEmptyFolder(path string) (bool, error) {
	matchingFiles, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return true, nil
		}

		return false, err
	}

	return len(matchingFiles) == 0, nil
}
