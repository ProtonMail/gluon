package tests

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/store"
	"github.com/stretchr/testify/require"
)

type TestStoreBuilder struct {
	builder store.Builder
	store   store.Store
}

func newTestStoreBuilder() *TestStoreBuilder {
	return &TestStoreBuilder{builder: &store.OnDiskStoreBuilder{}}
}

func (t *TestStoreBuilder) New(dir, userID string, passphrase []byte) (store.Store, error) {
	st, err := t.builder.New(dir, userID, passphrase)
	if err != nil {
		return nil, err
	}

	testStoreBuilderTestIDs := []imap.InternalMessageID{
		imap.NewInternalMessageID(), imap.NewInternalMessageID(), imap.NewInternalMessageID(), imap.NewInternalMessageID(),
	}

	for _, id := range testStoreBuilderTestIDs {
		if err := st.Set(id, []byte{0xD, 0xE, 0xA, 0xD, 0xB, 0xE, 0xE, 0xF}); err != nil {
			panic("failed to store test data in store")
		}
	}

	t.store = st

	return st, nil
}

func (t *TestStoreBuilder) Delete(dir, userID string) error {
	return t.builder.Delete(dir, userID)
}

func TestStoreCleanupOnStartup(t *testing.T) {
	// Add a bunch of random ids to the store and see if they are cleaned up on startup as they are not in
	// the database.
	testStore := newTestStoreBuilder()
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withStoreBuilder(testStore)), func(connection *testConnection, session *testSession) {
		idsInStore, err := testStore.store.List()
		require.NoError(t, err)
		require.Empty(t, idsInStore)
	})
}
