package store_test

import (
	"github.com/ProtonMail/gluon/imap"
	"runtime"
	"testing"

	"github.com/ProtonMail/gluon/store"
	"github.com/stretchr/testify/require"
)

func TestOnDiskStore(t *testing.T) {
	for name, cmp := range map[string]store.Compressor{
		"noop": nil,
		"gzip": &store.GZipCompressor{},
		"zlib": &store.ZLibCompressor{},
	} {
		t.Run(name, func(t *testing.T) {
			store, err := store.NewOnDiskStore(
				t.TempDir(),
				[]byte("pass"),
				store.WithCompressor(cmp),
				store.WithSemaphore(store.NewSemaphore(runtime.NumCPU())),
			)
			require.NoError(t, err)

			testStore(t, store)
			testStoreList(t, store)
		})
	}
}

func testStore(t *testing.T, store store.Store) {
	require.NoError(t, store.Set(1, []byte("literal1")))
	require.NoError(t, store.Set(2, []byte("literal2")))
	require.NoError(t, store.Set(3, []byte("literal3")))

	require.Equal(t, []byte("literal1"), must(store.Get(1)))
	require.Equal(t, []byte("literal2"), must(store.Get(2)))
	require.Equal(t, []byte("literal3"), must(store.Get(3)))
}

func testStoreList(t *testing.T, store store.Store) {
	require.NoError(t, store.Set(1, []byte("literal1")))
	require.NoError(t, store.Set(2, []byte("literal2")))
	require.NoError(t, store.Set(3, []byte("literal3")))

	list, err := store.List()
	require.NoError(t, err)
	require.ElementsMatch(t, list, []imap.InternalMessageID{1, 2, 3})
}

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}

	return val
}
