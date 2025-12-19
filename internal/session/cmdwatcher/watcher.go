package cmdwatcher

import (
	"context"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	defaultProgressInterval = 5 * time.Second
	defaultProgressMessage  = "Still working ..."
)

type commandMetadata struct {
	startTime        time.Time
	sanitizedPayload string
}

type SendProgressMessageFn func(msg string) error

type Service struct {
	ctx                   context.Context
	rwLock                sync.RWMutex
	commandWatch          map[uuid.UUID]commandMetadata
	log                   *logrus.Entry
	ticker                *time.Ticker
	sessionID             int
	sendProgressMessageFn SendProgressMessageFn
}

func NewAndRun(ctx context.Context, sessionID int, panicHandler async.PanicHandler, sendProgressMessageFn SendProgressMessageFn) *Service {
	watcher := &Service{
		ctx:                   ctx,
		commandWatch:          make(map[uuid.UUID]commandMetadata),
		log:                   logrus.WithField("pkg", "gluon/command-watcher").WithField("session", sessionID),
		ticker:                time.NewTicker(defaultProgressInterval),
		sessionID:             sessionID,
		sendProgressMessageFn: sendProgressMessageFn,
	}

	go func() {
		defer async.HandlePanic(panicHandler)
		watcher.monitorCommands()
	}()

	return watcher
}

func (c *Service) withRLock(fn func()) {
	c.rwLock.RLock()
	defer c.rwLock.RUnlock()
	fn()
}

func (c *Service) withWLock(fn func()) {
	c.rwLock.Lock()
	defer c.rwLock.Unlock()
	fn()
}

func (c *Service) monitorCommands() {
	defer c.ticker.Stop()
	for {
		select {
		case <-c.ticker.C:
			c.checkAndReportProgress()
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Service) checkAndReportProgress() {
	hasLongRunningCommand := false
	var sanitizedCommandPayload string

	c.withRLock(func() {
		for _, cmd := range c.commandWatch {
			if time.Since(cmd.startTime) > defaultProgressInterval {
				hasLongRunningCommand = true
				sanitizedCommandPayload = cmd.sanitizedPayload
				return
			}
		}
	})

	if !hasLongRunningCommand {
		return
	}

	log := c.log.WithField("cmd", sanitizedCommandPayload)

	if err := c.sendProgressMessageFn(defaultProgressMessage); err != nil {
		log.WithError(err).Error("Failed to send progress message")
	}

	log.Info("Sent progress message")
}

type TrackCommandFn func(sanitizedPayload string) func()

func (c *Service) TrackCommand(sanitizedPayload string) (cleanupFn func()) {
	id := uuid.New()
	c.withWLock(func() {
		c.commandWatch[id] = commandMetadata{
			startTime:        time.Now(),
			sanitizedPayload: sanitizedPayload,
		}
	})

	return func() {
		c.withWLock(func() {
			delete(c.commandWatch, id)
		})
	}
}
