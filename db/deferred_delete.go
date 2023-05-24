package db

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"os"
	"path/filepath"
)

// DeleteDB will rename all the database files for the given user to a directory within the same folder to avoid
// issues with ent not being able to close the database on demand. The database will be cleaned up on the next
// run on the Gluon server.
func DeleteDB(dir, userID string) error {
	// Rather than deleting the files immediately move them to a directory to be cleaned up later.
	deferredDeletePath := GetDeferredDeleteDBPath(dir)

	if err := os.MkdirAll(deferredDeletePath, 0o700); err != nil {
		return fmt.Errorf("failed to create deferred delete dir: %w", err)
	}

	matchingFiles, err := filepath.Glob(filepath.Join(dir, userID+"*"))
	if err != nil {
		return fmt.Errorf("failed to match db files:%w", err)
	}

	for _, file := range matchingFiles {
		// Use new UUID to avoid conflict with existing files
		if err := os.Rename(file, filepath.Join(deferredDeletePath, uuid.NewString())); err != nil {
			return fmt.Errorf("failed to move db file '%v' :%w", file, err)
		}
	}

	return nil
}

// DeleteDeferredDBFiles deletes all data from previous databases that were scheduled for removal.
func DeleteDeferredDBFiles(dir string) error {
	deferredDeleteDir := GetDeferredDeleteDBPath(dir)
	if err := os.RemoveAll(deferredDeleteDir); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	return nil
}
