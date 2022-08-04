package benchmark

import (
	"os"
)

// BenchDirConfig controls the directory where the benchmark data will be generated.
type BenchDirConfig interface {
	Get() (string, error)
}

// FixedBenchDirConfig always returns a known path.
type FixedBenchDirConfig struct {
	path string
}

func NewFixedBenchDirConfig(path string) *FixedBenchDirConfig {
	return &FixedBenchDirConfig{path: path}
}

func (p *FixedBenchDirConfig) Get() (string, error) {
	if err := os.MkdirAll(p.path, 0o777); err != nil {
		return "", err
	}

	return p.path, nil
}

// TmpBenchDirConfig returns a temporary path that is generated on first use.
type TmpBenchDirConfig struct {
	path string
}

func (t *TmpBenchDirConfig) Get() (string, error) {
	if len(t.path) == 0 {
		path, err := os.MkdirTemp("", "gluon-bench-*")
		if err != nil {
			return "", err
		}

		t.path = path
	}

	return t.path, nil
}
