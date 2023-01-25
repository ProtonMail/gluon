package store_test

import (
	"runtime"
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/store"
	"github.com/stretchr/testify/require"
)

func TestOnDiskStore(t *testing.T) {
	store, err := store.NewOnDiskStore(
		t.TempDir(),
		[]byte("pass"),
		store.WithSemaphore(store.NewSemaphore(runtime.NumCPU())),
	)
	require.NoError(t, err)

	testStore(t, store)
	testStoreList(t, store)
}

func testStore(t *testing.T, store store.Store) {
	id1 := imap.NewInternalMessageID()
	id2 := imap.NewInternalMessageID()
	id3 := imap.NewInternalMessageID()

	require.NoError(t, store.Set(id1, []byte("literal1")))
	require.NoError(t, store.Set(id2, []byte("literal2")))
	require.NoError(t, store.Set(id3, []byte("literal3")))

	require.Equal(t, []byte("literal1"), must(store.Get(id1)))
	require.Equal(t, []byte("literal2"), must(store.Get(id2)))
	require.Equal(t, []byte("literal3"), must(store.Get(id3)))

	require.NoError(t, store.Delete(id1, id2, id3))
}

func testStoreList(t *testing.T, store store.Store) {
	id1 := imap.NewInternalMessageID()
	id2 := imap.NewInternalMessageID()
	id3 := imap.NewInternalMessageID()

	require.NoError(t, store.Set(id1, []byte("literal1")))
	require.NoError(t, store.Set(id2, []byte("literal2")))
	require.NoError(t, store.Set(id3, []byte("literal3")))

	list, err := store.List()
	require.NoError(t, err)
	require.ElementsMatch(t, list, []imap.InternalMessageID{id1, id2, id3})

	require.NoError(t, store.Delete(id1, id2, id3))
}

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}

	return val
}
