package connectioncounter_test

import (
	"sync"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/imap/connectioncounter"
	"github.com/stretchr/testify/assert"
)

type mockConnProvider struct {
	openSessions int
}

func (m *mockConnProvider) GetOpenSessionCount() int {
	return m.openSessions
}

type mockObsSender struct {
	mu                                  sync.Mutex
	called                              bool
	lastOpenConns, lastNewlyOpenedConns int
}

func (m *mockObsSender) AddMetrics(_ ...map[string]any) {}

func (m *mockObsSender) AddDistinctMetrics(_ any, _ ...map[string]any) {}

func (m *mockObsSender) AddIMAPConnectionsExceededThresholdMetric(openConns, newlyOpenedConns int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.called = true
	m.lastOpenConns = openConns
	m.lastNewlyOpenedConns = newlyOpenedConns
}

func (m *mockObsSender) WasCalled() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.called
}

func (m *mockObsSender) LastValues() (int, int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastOpenConns, m.lastNewlyOpenedConns
}

func TestRollingCounter_ObservabilityThresholdNotExceeded(t *testing.T) {
	observabilityThreshold := 5
	connectionLimitThreshold := 5
	rc := connectioncounter.NewRollingCounter(
		connectionLimitThreshold,
		observabilityThreshold,
		3,
		100*time.Millisecond,
	)

	mockSender := &mockObsSender{}
	mockProvider := &mockConnProvider{openSessions: 10}

	ctx := t.Context()

	rc.Start(ctx, mockSender, mockProvider)

	rc.NewConnection()
	time.Sleep(100 * time.Millisecond)
	rc.NewConnection()
	time.Sleep(100 * time.Millisecond)
	rc.NewConnection()
	time.Sleep(100 * time.Millisecond)

	// Wait for the metric to be triggered.
	time.Sleep(300 * time.Millisecond)
	assert.False(t, mockSender.WasCalled())
	rc.Stop()
}

func TestRollingCounter_ObservabilityThresholdExceeded(t *testing.T) {
	observabilityThreshold := 3
	connectionLimitThreshold := 5
	rc := connectioncounter.NewRollingCounter(
		connectionLimitThreshold,
		observabilityThreshold,
		3,
		100*time.Millisecond,
	)

	mockSender := &mockObsSender{}
	mockProvider := &mockConnProvider{openSessions: 7}

	ctx := t.Context()

	rc.Start(ctx, mockSender, mockProvider)

	newConnsOpened := observabilityThreshold * 5
	for range newConnsOpened {
		rc.NewConnection()
	}

	time.Sleep(400 * time.Millisecond)

	assert.True(t, mockSender.WasCalled())

	open, newlyOpened := mockSender.LastValues()
	assert.Equal(t, 7, open)
	assert.Equal(t, newConnsOpened, newlyOpened)

	rc.Stop()
}

func TestRollingCounter_ConnectionLimitThresholdExceeded(t *testing.T) {
	observabilityThreshold := 3
	connectionLimitThreshold := 3
	rc := connectioncounter.NewRollingCounter(
		connectionLimitThreshold,
		observabilityThreshold,
		3,
		100*time.Millisecond,
	)

	mockSender := &mockObsSender{}
	mockProvider := &mockConnProvider{openSessions: 7}

	ctx := t.Context()

	rc.Start(ctx, mockSender, mockProvider)

	newConnsOpened := observabilityThreshold * 5
	for range newConnsOpened {
		rc.NewConnection()
	}
	assert.Equal(t, rc.GetRollingCount(), newConnsOpened)
	assert.True(t, rc.OverConnectionLimitThreshold())

	rc.Stop()
}

func TestRollingCounter_BucketRotation(t *testing.T) {
	jitterPeriod := 50 * time.Millisecond
	bucketRotationInterval := 500 * time.Millisecond
	rc := connectioncounter.NewRollingCounter(
		100,
		100,
		3,
		bucketRotationInterval,
	)

	post10Connections := func() {
		for range 10 {
			rc.NewConnection()
		}
	}

	ctx := t.Context()

	mockSender := &mockObsSender{}
	mockProvider := &mockConnProvider{}

	rc.Start(ctx, mockSender, mockProvider)

	// Input to 1st bucket.
	post10Connections()
	assert.Equal(t, 10, rc.GetRollingCount())

	// Input to 2nd bucket.
	time.Sleep(bucketRotationInterval + jitterPeriod)
	post10Connections()
	assert.Equal(t, 20, rc.GetRollingCount())

	// Input to 3rd bucket.
	time.Sleep(bucketRotationInterval + jitterPeriod)
	post10Connections()
	assert.Equal(t, 30, rc.GetRollingCount())

	// Input to 1st bucket.
	time.Sleep(bucketRotationInterval + jitterPeriod)
	post10Connections()
	assert.Equal(t, 30, rc.GetRollingCount())

	// Input to the 2nd bucket.
	time.Sleep(bucketRotationInterval + jitterPeriod)
	post10Connections()
	assert.Equal(t, 30, rc.GetRollingCount())

	// 2nd bucket reset.
	time.Sleep(bucketRotationInterval + jitterPeriod)
	assert.Equal(t, 20, rc.GetRollingCount())
	// 3rd bucket reset.
	time.Sleep(bucketRotationInterval + jitterPeriod)
	assert.Equal(t, 10, rc.GetRollingCount())
	// 1st bucket reset.
	time.Sleep(bucketRotationInterval + jitterPeriod)
	assert.Equal(t, 0, rc.GetRollingCount())

	rc.Stop()
}
