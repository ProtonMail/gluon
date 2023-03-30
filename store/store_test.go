package store_test

import (
	"bytes"
	"crypto/rand"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/store"
	"github.com/ProtonMail/gluon/store/fallback_v0"
	"github.com/stretchr/testify/require"
)

func TestStore_DecryptFailedOnFilesBiggerThanBlockSize(t *testing.T) {
	store, err := store.NewOnDiskStore(
		t.TempDir(),
		[]byte("pass"),
		store.WithSemaphore(store.NewSemaphore(runtime.NumCPU(), async.NoopPanicHandler{})),
	)
	require.NoError(t, err)

	data := make([]byte, 1024*1204)
	{
		_, err := rand.Read(data) //nolint:gosec
		require.NoError(t, err)
	}

	id := imap.NewInternalMessageID()
	require.NoError(t, store.Set(id, bytes.NewReader(data)))
	read, err := store.Get(id)
	require.NoError(t, err)
	require.True(t, bytes.Equal(read, data))
	require.NoError(t, store.Delete(id))
}

func BenchmarkStoreRead(t *testing.B) {
	store, err := store.NewOnDiskStore(
		t.TempDir(),
		[]byte("pass"),
		store.WithSemaphore(store.NewSemaphore(runtime.NumCPU(), async.NoopPanicHandler{})),
	)
	require.NoError(t, err)

	data := make([]byte, 15*1024*1204)
	{
		_, err := rand.Read(data) //nolint:gosec
		require.NoError(t, err)
	}

	id := imap.NewInternalMessageID()
	require.NoError(t, store.Set(id, bytes.NewReader(data)))

	t.ResetTimer()

	for i := 0; i < t.N; i++ {
		_, err := store.Get(id)
		require.NoError(t, err)
	}
}

func TestStoreReadFailsIfHeaderDoesNotMatch(t *testing.T) {
	storeDir := t.TempDir()
	store, err := store.NewOnDiskStore(
		storeDir,
		[]byte("pass"),
		store.WithSemaphore(store.NewSemaphore(runtime.NumCPU(), async.NoopPanicHandler{})),
	)
	require.NoError(t, err)

	id := imap.NewInternalMessageID()
	// Generate dummy file
	{
		data := make([]byte, 15*1024)
		_, err := rand.Read(data) //nolint:gosec
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(storeDir, id.String()), data, 0o600))
	}

	_, err = store.Get(id)
	require.Error(t, err)
}

func TestStoreFallbackRead(t *testing.T) {
	fallbackStore := fallback_v0.NewOnDiskStoreV0WithCompressor(&fallback_v0.GZipCompressor{})

	storeDir := t.TempDir()

	password := []byte("pass")

	fileContents := []byte("Hello world from gluon store")

	id := imap.NewInternalMessageID()

	{
		// create old store file on disk
		gcm, err := store.NewCipher(password)
		require.NoError(t, err)

		filepath := filepath.Join(storeDir, id.String())

		require.NoError(t, fallbackStore.Write(gcm, filepath, fileContents))
	}

	// Reading file without fallback should fail.
	{
		store, err := store.NewOnDiskStore(
			storeDir,
			[]byte("pass"),
			store.WithSemaphore(store.NewSemaphore(runtime.NumCPU(), async.NoopPanicHandler{})),
		)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, store.Close())
		}()

		_, err = store.Get(id)
		require.Error(t, err)
	}

	//
	{
		store, err := store.NewOnDiskStore(
			storeDir,
			[]byte("pass"),
			store.WithSemaphore(store.NewSemaphore(runtime.NumCPU(), async.NoopPanicHandler{})),
			store.WithFallback(fallbackStore),
		)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, store.Close())
		}()

		b, err := store.Get(id)
		require.NoError(t, err)
		require.Equal(t, fileContents, b)
	}
}

func TestOnDiskStore(t *testing.T) {
	store, err := store.NewOnDiskStore(
		t.TempDir(),
		[]byte("pass"),
		store.WithSemaphore(store.NewSemaphore(runtime.NumCPU(), async.NoopPanicHandler{})),
	)
	require.NoError(t, err)

	testStore(t, store)
	testStoreList(t, store)
}

func testStore(t *testing.T, store store.Store) {
	id1 := imap.NewInternalMessageID()
	id2 := imap.NewInternalMessageID()
	id3 := imap.NewInternalMessageID()

	require.NoError(t, store.Set(id1, bytes.NewReader([]byte("literal1"))))
	require.NoError(t, store.Set(id2, bytes.NewReader([]byte("literal2"))))
	require.NoError(t, store.Set(id3, bytes.NewReader([]byte("literal3"))))

	require.Equal(t, []byte("literal1"), must(store.Get(id1)))
	require.Equal(t, []byte("literal2"), must(store.Get(id2)))
	require.Equal(t, []byte("literal3"), must(store.Get(id3)))

	require.NoError(t, store.Delete(id1, id2, id3))
}

func testStoreList(t *testing.T, store store.Store) {
	id1 := imap.NewInternalMessageID()
	id2 := imap.NewInternalMessageID()
	id3 := imap.NewInternalMessageID()

	require.NoError(t, store.Set(id1, bytes.NewReader([]byte("literal1"))))
	require.NoError(t, store.Set(id2, bytes.NewReader([]byte("literal2"))))
	require.NoError(t, store.Set(id3, bytes.NewReader([]byte("literal3"))))

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
