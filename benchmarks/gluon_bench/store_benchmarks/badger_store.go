package store_benchmarks

import (
	"crypto/sha256"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/ProtonMail/gluon/store"
	"github.com/google/uuid"
)

type BadgerStoreBuilder struct{}

func (*BadgerStoreBuilder) New(path string) (store.Store, error) {
	encryptionKey := sha256.Sum256([]byte(*flags.UserPassword))
	return store.NewBadgerStore(path, uuid.NewString(), encryptionKey[:])
}

func init() {
	RegisterStoreBuilder("default", &BadgerStoreBuilder{})
}
