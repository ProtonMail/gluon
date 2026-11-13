package tests

import (
	"testing"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/store"
)

// Mock Store which panics on Every Get.
// Used to simulate a panic in the apply call.
type panicOnGetStore struct {
	store.Store
}

func (s *panicOnGetStore) Get(imap.InternalMessageID) ([]byte, error) {
	panic("synthetic panic injected by test")
}

type panicOnGetStoreBuilder struct{}

func (b *panicOnGetStoreBuilder) New(dir, userID string, passphrase []byte) (store.Store, error) {
	inner, err := (&store.OnDiskStoreBuilder{}).New(dir, userID, passphrase)
	if err != nil {
		return nil, err
	}

	return &panicOnGetStore{Store: inner}, nil
}

func (b *panicOnGetStoreBuilder) Delete(dir, userID string) error {
	return (&store.OnDiskStoreBuilder{}).Delete(dir, userID)
}

// TestApplyRecoversFromPanicAndKeepsProcessingUpdates tests for applys() panic safety.
// A panic in apply() must still call update.Done() so anyone waiting on Wait/WaitContext is unblocked.
func TestApplyRecoversFromPanicAndKeepsProcessingUpdates(t *testing.T) {
	options := defaultServerOptions(t, withStoreBuilder(&panicOnGetStoreBuilder{}))

	runOneToOneTestWithAuth(t, options, func(c *testConnection, s *testSession) {
		mailboxID := s.mailboxCreated("user", []string{"mbox1"})
		messageID := s.messageCreated("user", mailboxID, buildStructuralTestLiteral("outerA", "innerA", "attachment"), time.Now())

		s.setUpdatesAllowedToFail("user", true)

		updateDone := make(chan struct{})

		go func() {
			defer close(updateDone)

			s.messageUpdatedWithID(
				"user", messageID, mailboxID,
				buildStructuralTestLiteral("outerB", "innerB", "attachment"),
				time.Now(),
			)
		}()

		select {
		case <-updateDone:
		case <-time.After(5 * time.Second):
			t.Fatal("MessageUpdated did not complete after a panic - update.Done() was potentially skipped")
		}

		// per-user update goroutine should have survived; a later unrelated update must be still applicable.
		otherDone := make(chan struct{})

		go func() {
			defer close(otherDone)

			s.mailboxCreated("user", []string{"mbox2"})
		}()

		select {
		case <-otherDone:
		case <-time.After(5 * time.Second):
			t.Fatal("A later update did not apply after a panic - the per-user update goroutine potentially died")
		}

		c.C(`A001 STATUS mbox2 (MESSAGES)`)
		c.S(`* STATUS "mbox2" (MESSAGES 0)`)
		c.OK(`A001`)
	})
}
