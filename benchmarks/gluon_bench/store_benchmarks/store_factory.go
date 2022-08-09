package store_benchmarks

import (
	"fmt"

	"github.com/ProtonMail/gluon/store"
)

type StoreBuilder interface {
	New(path string) (store.Store, error)
}

type storeFactory struct {
	builders map[string]StoreBuilder
}

func newStoreFactory() *storeFactory {
	return &storeFactory{builders: make(map[string]StoreBuilder)}
}

func (sf *storeFactory) Register(name string, builder StoreBuilder) error {
	if _, ok := sf.builders[name]; ok {
		return fmt.Errorf("builder already exists")
	}

	sf.builders[name] = builder

	return nil
}

func (sf *storeFactory) New(name, path string) (store.Store, error) {
	builder, ok := sf.builders[name]

	if !ok {
		return nil, fmt.Errorf("no such builder exists")
	}

	return builder.New(path)
}

var storeFactoryInstance = newStoreFactory()

func RegisterStoreBuilder(name string, storeBuilder StoreBuilder) {
	if err := storeFactoryInstance.Register(name, storeBuilder); err != nil {
		panic(err)
	}
}

func NewStore(name, path string) (store.Store, error) {
	return storeFactoryInstance.New(name, path)
}
