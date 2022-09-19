package store_benchmarks

import (
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/store"
	"github.com/google/uuid"
)

type BadgerStoreBuilder struct{}

func (*BadgerStoreBuilder) New(path string) (store.Store, error) {
	return store.NewBadgerStore(path, uuid.NewString(), []byte(*flags.UserPassword))
}

func init() {
	RegisterStoreBuilder("default", &BadgerStoreBuilder{})
}
