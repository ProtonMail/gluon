package store_benchmarks

import (
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/store"
)

type DefaultStoreBuilder struct{}

func (*DefaultStoreBuilder) New(path string) (store.Store, error) {
	return store.NewOnDiskStore(path, []byte(*flags.UserPassword))
}

func init() {
	RegisterStoreBuilder("default", &DefaultStoreBuilder{})
}
