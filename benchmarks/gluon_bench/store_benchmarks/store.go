package store_benchmarks

import (
	"path/filepath"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/store"
	"github.com/google/uuid"
)

type OnDiskStoreBuilder struct{}

func (*OnDiskStoreBuilder) New(path string) (store.Store, error) {
	return store.NewOnDiskStore(filepath.Join(path, uuid.NewString()), []byte(*flags.UserPassword))
}

func init() {
	RegisterStoreBuilder("default", &OnDiskStoreBuilder{})
}
