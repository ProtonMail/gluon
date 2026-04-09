package store

import (
	"bytes"
	"sync"
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/stretchr/testify/require"
)

func TestWriteControlledStore(t *testing.T) {
	id1 := imap.NewInternalMessageID()
	id2 := imap.NewInternalMessageID()
	id3 := imap.NewInternalMessageID()

	st, err := NewOnDiskStore(
		t.TempDir(),
		[]byte("pass"),
	)
	require.NoError(t, err)

	st = NewWriteControlledStore(st)

	wg := sync.WaitGroup{}

	for i := range 256 {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			var id imap.InternalMessageID

			switch i % 3 {
			case 0:
				require.NoError(t, st.Set(id1, bytes.NewReader([]byte("literal1")))) //nolint:testifylint
				id = id1
			case 1:
				require.NoError(t, st.Set(id2, bytes.NewReader([]byte("literal2")))) //nolint:testifylint
				id = id2
			case 2:
				require.NoError(t, st.Set(id3, bytes.NewReader([]byte("literal3")))) //nolint:testifylint
				id = id3
			}

			require.NotEmpty(t, id, imap.InternalMessageID{}) //nolint:testifylint

			// It's not guaranteed which version of the literal will be available on disk, but it should be
			// match one of the following
			literal, err := st.Get(id)
			require.NoError(t, err) //nolint:testifylint

			isEqual := bytes.Equal([]byte("literal1"), literal) ||
				bytes.Equal([]byte("literal2"), literal) ||
				bytes.Equal([]byte("literal3"), literal)

			require.True(t, isEqual) //nolint:testifylint
		}(i)
	}

	wg.Wait()
}
