package connectioncounter

import (
	"context"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/observability"
	"github.com/sirupsen/logrus"
)

type openConnectionProvider interface {
	GetOpenSessionCount() int
}

type RollingCounter struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	log *logrus.Entry

	newConnectionThreshold int

	numberOfBuckets int
	buckets         []int
	current         int

	bucketLock sync.Mutex

	bucketRotationInterval time.Duration
	bucketRotationTicker   *time.Ticker

	observabilitySender observability.Sender
	connProvider        openConnectionProvider
}

func NewRollingCounter(newConnectionTreshold, numberOfBuckets int, bucketRotationInterval time.Duration) *RollingCounter {
	log := logrus.WithFields(logrus.Fields{
		"pkg":       "gluon/rollingcounter",
		"threshold": newConnectionTreshold,
	})

	rc := &RollingCounter{
		newConnectionThreshold: newConnectionTreshold,
		numberOfBuckets:        numberOfBuckets,
		bucketRotationInterval: bucketRotationInterval,
		log:                    log,
	}

	return rc
}

func (rc *RollingCounter) Start(ctx context.Context, obsSender observability.Sender, connProvider openConnectionProvider) {
	ctx, cancel := context.WithCancel(ctx)

	rc.ctx = ctx
	rc.cancel = cancel
	rc.observabilitySender = obsSender
	rc.connProvider = connProvider

	if rc.bucketRotationInterval <= 0 {
		rc.log.Error("bucketRotationInterval must be a positive, non-zero duration")
		return
	}

	rc.buckets = make([]int, rc.numberOfBuckets)

	// Start the ticker.
	rc.bucketRotationTicker = time.NewTicker(rc.bucketRotationInterval)

	// Start the service in a separate goroutine.
	rc.run()
}

func (rc *RollingCounter) run() {
	rc.wg.Add(1)
	go func() {
		defer rc.wg.Done()
		for {
			select {
			case <-rc.ctx.Done():
				return

			case <-rc.bucketRotationTicker.C:
				rc.thresholdCheck()
				rc.onBucketRotationTick()
			}
		}
	}()
}

func (rc *RollingCounter) Stop() {
	rc.bucketRotationTicker.Stop()
	rc.cancel()
	rc.wg.Wait()
}

func (rc *RollingCounter) withBucketLock(fn func()) {
	rc.bucketLock.Lock()
	defer rc.bucketLock.Unlock()
	fn()
}

func (rc *RollingCounter) thresholdCheck() {
	rollingCount := rc.GetRollingCount()
	if rollingCount < rc.newConnectionThreshold {
		return
	}

	openSessionCount := rc.connProvider.GetOpenSessionCount()
	rc.log.WithFields(logrus.Fields{
		"newlyOpenedIMAPConnections": rollingCount,
		"openIMAPConnections":        openSessionCount}).Info("Newly opened IMAP connections exceed threshold")

	rc.observabilitySender.AddIMAPConnectionsExceededThresholdMetric(openSessionCount, rollingCount)
}

func (rc *RollingCounter) onBucketRotationTick() {
	fn := func() {
		rc.current = (rc.current + 1) % rc.numberOfBuckets
		rc.buckets[rc.current] = 0
	}
	rc.withBucketLock(fn)
}

func (rc *RollingCounter) NewConnection() {
	fn := func() {
		rc.buckets[rc.current]++
	}
	rc.withBucketLock(fn)
}

func (rc *RollingCounter) GetRollingCount() int {
	rc.bucketLock.Lock()
	defer rc.bucketLock.Unlock()

	var rollingCount int
	for _, count := range rc.buckets {
		rollingCount += count
	}

	return rollingCount
}
