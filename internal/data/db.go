package data

import (
	"errors"
	"io/fs"
	"os"
)

// pathExists returns whether the given file exists.
//
// nolint: unused
func pathExists(path string) (bool, error) {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
