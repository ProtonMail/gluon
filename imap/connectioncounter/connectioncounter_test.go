package connectioncounter_test

import (
	"context"
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

func (m *mockObsSender) AddMetrics(_ ...map[string]interface{}) {}

func (m *mockObsSender) AddDistinctMetrics(_ interface{}, _ ...map[string]interface{}) {}

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

func TestRollingCounter_ThresholdNotExceeded(t *testing.T) {
	rc := connectioncounter.NewRollingCounter(
		5,
		3,
		100*time.Millisecond,
	)

	mockSender := &mockObsSender{}
	mockProvider := &mockConnProvider{openSessions: 10}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

func TestRollingCounter_ThresholdExceeded(t *testing.T) {
	newConnThreshold := 3
	rc := connectioncounter.NewRollingCounter(
		newConnThreshold,
		3,
		100*time.Millisecond,
	)

	mockSender := &mockObsSender{}
	mockProvider := &mockConnProvider{openSessions: 7}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rc.Start(ctx, mockSender, mockProvider)

	newConnsOpened := newConnThreshold * 5
	for i := 0; i < newConnsOpened; i++ {
		rc.NewConnection()
	}

	time.Sleep(400 * time.Millisecond)

	assert.True(t, mockSender.WasCalled())

	open, newlyOpened := mockSender.LastValues()
	assert.Equal(t, 7, open)
	assert.Equal(t, newConnsOpened, newlyOpened)

	rc.Stop()
}

func TestRollingCounter_BucketRotation(t *testing.T) {
	jitterPeriod := 50 * time.Millisecond
	bucketRotationInterval := 500 * time.Millisecond
	rc := connectioncounter.NewRollingCounter(
		100,
		3,
		bucketRotationInterval,
	)

	post10Connections := func() {
		for i := 0; i < 10; i++ {
			rc.NewConnection()
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
