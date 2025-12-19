package cmdwatcher_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/session/cmdwatcher"
	"github.com/ProtonMail/gluon/internal/unleash"
	"github.com/ProtonMail/gluon/internal/unleash/featureflags"
	"github.com/stretchr/testify/assert"
)

const defaultTestInterval = 40 * time.Millisecond

type mockProgressSender struct {
	lock        sync.Mutex
	called      bool
	callCount   int
	errToReturn error
}

func (m *mockProgressSender) SendProgress(msg string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.called = true
	m.callCount++
	return m.errToReturn
}

func (m *mockProgressSender) WasCalled() bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.called
}

func (m *mockProgressSender) ResetCalled() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.called = false
}

func (m *mockProgressSender) GetCallCount() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.callCount
}

func (m *mockProgressSender) SetErrToReturn(err error) {
	m.errToReturn = err
}

type mockFeatureFlagProvider struct {
	flags map[string]bool
}

func newMockFeatureFlagProvider() *mockFeatureFlagProvider {
	return &mockFeatureFlagProvider{
		flags: make(map[string]bool),
	}
}

func (m *mockFeatureFlagProvider) GetFlagValue(key string) bool {
	return m.flags[key]
}

func (m *mockFeatureFlagProvider) SetFlag(key string, value bool) {
	m.flags[key] = value
}

type mockPanicHandler struct{}

func (m *mockPanicHandler) HandlePanic(r interface{}) {}

func createTestService(testInterval time.Duration, featureFlags map[string]bool) (*cmdwatcher.Service, *mockProgressSender) {
	ctx := context.Background()
	sender := &mockProgressSender{}
	flagProvider := newMockFeatureFlagProvider()

	for key, val := range featureFlags {
		flagProvider.SetFlag(key, val)
	}

	service := cmdwatcher.NewAndRun(
		ctx,
		1,
		&mockPanicHandler{},
		flagProvider,
		sender.SendProgress,
		cmdwatcher.WithProgressInterval(testInterval),
	)

	return service, sender
}

func createImapID(name string) imap.IMAPID {
	return imap.IMAPID{
		Name:    name,
		Version: "1.0",
		OS:      "test-os",
	}
}

func TestCheckAndReportProgress_GlobalDisabled(t *testing.T) {
	service, sender := createTestService(defaultTestInterval, map[string]bool{
		featureflags.CommandWatcherGlobalDisabled: true,
	})

	imapID := createImapID("Thunderbird")
	service.TrackedWithImapID(imapID)("TEST_CMD")

	time.Sleep(45 * time.Millisecond)
	assert.False(t, sender.WasCalled())
}

func TestCheckAndReportProgress_NonThunderbirdDisabled_WithThunderbird(t *testing.T) {
	tests := []struct {
		name     string
		imapName string
	}{
		{"lowercase", "thunderbird"},
		{"uppercase", "THUNDERBIRD"},
		{"mixed case", "ThUnDeRbIrD"},
		{"with prefix", "Mozilla Thunderbird"},
		{"with version", "Thunderbird 115.0"},
		{"fork", "my-thunderbird-fork"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, sender := createTestService(defaultTestInterval, map[string]bool{
				featureflags.CommandWatcherNonThunderbirdDisabled: true,
			})

			imapID := createImapID(tt.imapName)
			service.TrackedWithImapID(imapID)("TEST_CMD")

			time.Sleep(45 * time.Millisecond)

			assert.True(t, sender.WasCalled())
		})
	}
}

func TestCheckAndReportProgress_NonThunderbirdDisabled_WithoutThunderbird(t *testing.T) {
	tests := []struct {
		name     string
		imapName string
	}{
		{"outlook", "Outlook"},
		{"apple mail", "Apple Mail"},
		{"gmail", "Gmail"},
		{"empty", ""},
		{"generic", "IMAP Client"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, sender := createTestService(defaultTestInterval, map[string]bool{
				featureflags.CommandWatcherNonThunderbirdDisabled: true,
			})

			imapID := createImapID(tt.imapName)
			service.TrackedWithImapID(imapID)("TEST_CMD")

			time.Sleep(45 * time.Millisecond)

			assert.False(t, sender.WasCalled())
		})
	}
}

func TestCheckAndReportProgress_BothFlagsDisabled(t *testing.T) {
	tests := []struct {
		name     string
		imapName string
	}{
		{"test 1", "test 1 client"},
		{"thunderbird", "Thunderbird"},
		{"outlook", "Outlook"},
		{"apple mail", "Apple Mail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, sender := createTestService(defaultTestInterval, map[string]bool{
				featureflags.CommandWatcherGlobalDisabled:         false,
				featureflags.CommandWatcherNonThunderbirdDisabled: false,
			})

			imapID := createImapID(tt.imapName)
			service.TrackedWithImapID(imapID)("TEST_CMD")

			time.Sleep(45 * time.Millisecond)

			assert.True(t, sender.WasCalled())
		})
	}
}

func TestTrackCommand_BasicTracking(t *testing.T) {
	service, sender := createTestService(defaultTestInterval, map[string]bool{})

	imapID := createImapID("TEST")
	cleanup := service.TrackedWithImapID(imapID)("TEST_CMD")

	// Track command, should be sent.
	time.Sleep(45 * time.Millisecond)
	assert.True(t, sender.WasCalled())

	sender.ResetCalled()
	assert.False(t, sender.WasCalled())
	time.Sleep(45 * time.Millisecond)
	assert.True(t, sender.WasCalled())

	// Cleanup and wait, a new message should not be sent.
	cleanup()
	sender.ResetCalled()

	time.Sleep(45 * time.Millisecond)
	assert.False(t, sender.WasCalled())
}

func TestTrackCommand_CleanupFunction(t *testing.T) {
	service, sender := createTestService(defaultTestInterval, map[string]bool{})

	imapID := createImapID("Test")
	cleanup := service.TrackedWithImapID(imapID)("TEST_CMD")

	time.Sleep(45 * time.Millisecond)
	callCount := sender.GetCallCount()
	assert.Equal(t, 1, callCount)

	time.Sleep(45 * time.Millisecond)
	callCount = sender.GetCallCount()
	assert.Equal(t, 2, callCount)

	time.Sleep(45 * time.Millisecond)
	callCount = sender.GetCallCount()
	assert.Equal(t, 3, callCount)

	cleanup()

	time.Sleep(45 * time.Millisecond)
	callCount = sender.GetCallCount()
	assert.Equal(t, 3, callCount)

	time.Sleep(45 * time.Millisecond)
	callCount = sender.GetCallCount()
	assert.Equal(t, 3, callCount)
}

func TestCheckAndReportProgress_NoCommand(t *testing.T) {
	_, sender := createTestService(defaultTestInterval, map[string]bool{})

	time.Sleep(90 * time.Millisecond)
	assert.False(t, sender.WasCalled())
}

func TestCheckAndReportProgress_CtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sender := &mockProgressSender{}
	flagProvider := &unleash.NullFeatureFlagProvider{}

	service := cmdwatcher.NewAndRun(
		ctx,
		1,
		&mockPanicHandler{},
		flagProvider,
		sender.SendProgress,
		cmdwatcher.WithProgressInterval(defaultTestInterval),
	)

	cancel()

	imapID := createImapID("TestClient")
	cleanup := service.TrackedWithImapID(imapID)("FETCH 1:*")
	defer cleanup()

	time.Sleep(90 * time.Millisecond)

	assert.False(t, sender.WasCalled())
}

func TestCheckAndReportProgress_CommandNotOldEnough(t *testing.T) {
	service, sender := createTestService(defaultTestInterval, map[string]bool{})

	imapID := createImapID("TestClient")
	cleanup := service.TrackedWithImapID(imapID)("FETCH 1:*")
	defer cleanup()

	time.Sleep(30 * time.Millisecond)
	assert.False(t, sender.WasCalled())
}

func TestCheckAndReportProgress_SendProgressError(t *testing.T) {
	service, sender := createTestService(defaultTestInterval, map[string]bool{})

	sender.SetErrToReturn(errors.New("test error"))
	imapID := createImapID("TestClient")

	cleanup := service.TrackedWithImapID(imapID)("FETCH 1:*")
	time.Sleep(45 * time.Millisecond)
	assert.True(t, sender.WasCalled())

	time.Sleep(45 * time.Millisecond)
	assert.Equal(t, 2, sender.GetCallCount())

	time.Sleep(45 * time.Millisecond)
	assert.Equal(t, 3, sender.GetCallCount())

	cleanup()
	time.Sleep(45 * time.Millisecond)
	assert.Equal(t, 3, sender.GetCallCount())
}
