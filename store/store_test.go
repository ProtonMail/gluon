package store_test

import (
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/ProtonMail/gluon/imap"
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
		})
	}
}

func TestInMemoryStore(t *testing.T) {
	testStore(t, store.NewInMemoryStore())
}

func testStore(t *testing.T, store store.Store) {
	require.NoError(t, store.Set("messageID1", []byte("literal1")))
	require.NoError(t, store.Set("messageID2", []byte("literal2")))
	require.NoError(t, store.Set("messageID3", []byte("literal3")))

	require.Equal(t, []byte("literal1"), must(store.Get("messageID1")))
	require.Equal(t, []byte("literal2"), must(store.Get("messageID2")))
	require.Equal(t, []byte("literal3"), must(store.Get("messageID3")))

	var wg sync.WaitGroup

	for i := 0; i < 1<<5; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			for j := 0; j < 1<<5; j++ {
				message := imap.InternalMessageID(fmt.Sprintf("message(%v)(%v)", i, j))
				literal := fmt.Sprintf("literal(%v)(%v)", i, j)

				require.NoError(t, store.Set(message, []byte(literal)))
				require.Equal(t, []byte(literal), must(store.Get(message)))
			}
		}(i)
	}

	wg.Wait()
}

func must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}

	return val
}
