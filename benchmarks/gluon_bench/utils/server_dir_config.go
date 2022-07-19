package utils

import (
	"os"
)

// ServerDirConfig controls the directory where the server data will be generated.
type ServerDirConfig interface {
	Get() (string, error)
}

// PersistentServerDirConfig always returns a known path.
type PersistentServerDirConfig struct {
	path string
}

func NewPersistentServerDirConfig(path string) *PersistentServerDirConfig {
	return &PersistentServerDirConfig{path: path}
}

func (p *PersistentServerDirConfig) Get() (string, error) {
	if err := os.MkdirAll(p.path, 0o777); err != nil {
		return "", err
	}

	return p.path, nil
}

// TmpServerDirConfig returns a temporary path that is generated on first use.
type TmpServerDirConfig struct {
	path string
}

func (t *TmpServerDirConfig) Get() (string, error) {
	if len(t.path) == 0 {
		path, err := os.MkdirTemp("", "gluon-bench-*")
		if err != nil {
			return "", err
		}

		t.path = path
	}

	return t.path, nil
}
