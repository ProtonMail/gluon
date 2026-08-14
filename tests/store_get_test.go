package tests

import (
	"testing"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/store"
	"github.com/stretchr/testify/require"
)

// Store mock with an artificial delay before every Get.
// The goal is to deterministically widen the window which a lock held around the call would block concurrent DB access.
type slowGetStore struct {
	store.Store
	delay time.Duration
}

func (s *slowGetStore) Get(id imap.InternalMessageID) ([]byte, error) {
	time.Sleep(s.delay)
	return s.Store.Get(id)
}

type slowGetStoreBuilder struct {
	delay time.Duration
}

func (b *slowGetStoreBuilder) New(dir, userID string, passphrase []byte) (store.Store, error) {
	inner, err := (&store.OnDiskStoreBuilder{}).New(dir, userID, passphrase)
	if err != nil {
		return nil, err
	}

	return &slowGetStore{Store: inner, delay: b.delay}, nil
}

func (b *slowGetStoreBuilder) Delete(dir, userID string) error {
	return (&store.OnDiskStoreBuilder{}).Delete(dir, userID)
}

// TestConcurrentFetchDuringMessageUpdateDoesNotStall tests for a lock-order hazard in applyMessageUpdated.
// Reading on-disk literal must never happen while the per-user DB write lock is held. The reason being every concurrent
// FETCH on the account is stalled for as long as the read will take.
func TestConcurrentFetchDuringMessageUpdateDoesNotStall(t *testing.T) {
	const storeGetDelay = 500 * time.Millisecond

	options := defaultServerOptions(t, withStoreBuilder(&slowGetStoreBuilder{delay: storeGetDelay}))

	runOneToOneTestWithAuth(t, options, func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox1"})
		messageID := s.messageCreated("user", mailboxID, buildStructuralTestLiteral("outerA", "innerA", "attachment"), time.Now())

		c.C(`A001 SELECT mbox1`).OK(`A001`)

		// Mailboxes-only update path, reads the on-disk literal, should hit the artificially slowed store.Get
		updateDone := make(chan struct{})

		go func() {
			defer close(updateDone)

			s.messageUpdatedWithID(
				"user",
				messageID,
				mailboxID,
				buildStructuralTestLiteral("outerB", "innerB", "attachment"),
				time.Now(),
			)
		}()

		time.Sleep(storeGetDelay / 5)

		fetchStart := time.Now()
		withTag(func(tag string) {
			c.Cf(`%v FETCH 1 (FLAGS)`, tag).OK(tag)
		})
		fetchElapsed := time.Since(fetchStart)

		require.Less(t, fetchElapsed, storeGetDelay/2,
			"FETCH took %v while a concurrent MessageUpdated was reading the store", fetchElapsed)

		select {
		case <-updateDone:
		case <-time.After(5 * time.Second):
			t.Fatal("MessageUpdated did not complete in time")
		}
	})
}
