package cmdwatcher

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/unleash"
	"github.com/ProtonMail/gluon/internal/unleash/featureflags"
	"github.com/sirupsen/logrus"
)

const (
	defaultProgressInterval = 5 * time.Second
	defaultProgressMessage  = "Still working ..."
	thunderbirdIMAPName     = "thunderbird"
)

type commandMetadata struct {
	startTime        time.Time
	sanitizedPayload string
	imapID           imap.IMAPID
}

type SendProgressMessageFn func(msg string) error

type Service struct {
	ctx                   context.Context
	lock                  sync.Mutex
	currentCommand        *commandMetadata
	log                   *logrus.Entry
	ticker                *time.Ticker
	progressInterval      time.Duration
	sessionID             int
	sendProgressMessageFn SendProgressMessageFn
	featureFlagProvider   unleash.FeatureFlagValueProvider
}

type Option func(*Service)

func WithProgressInterval(interval time.Duration) Option {
	return func(s *Service) {
		s.progressInterval = interval
	}
}

func NewAndRun(
	ctx context.Context,
	sessionID int,
	panicHandler async.PanicHandler,
	featureFlagProvider unleash.FeatureFlagValueProvider,
	sendProgressMessageFn SendProgressMessageFn,
	opts ...Option,
) *Service {
	watcher := &Service{
		ctx:                   ctx,
		log:                   logrus.WithField("pkg", "gluon/command-watcher").WithField("session", sessionID),
		progressInterval:      defaultProgressInterval,
		sessionID:             sessionID,
		sendProgressMessageFn: sendProgressMessageFn,
		featureFlagProvider:   featureFlagProvider,
	}

	for _, opt := range opts {
		opt(watcher)
	}

	watcher.ticker = time.NewTicker(watcher.progressInterval)

	go func() {
		defer async.HandlePanic(panicHandler)
		watcher.monitorCommands()
	}()

	return watcher
}

func (s *Service) withLock(fn func()) {
	s.lock.Lock()
	defer s.lock.Unlock()
	fn()
}

func (s *Service) monitorCommands() {
	defer s.ticker.Stop()
	for {
		select {
		case <-s.ticker.C:
			s.checkAndReportProgress()
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Service) checkAndReportProgress() {
	if s.featureFlagProvider.GetFlagValue(featureflags.CommandWatcherGlobalDisabled) {
		return
	}

	var cmd commandMetadata
	var hasCmd bool

	s.withLock(func() {
		// Clarification: As the progress interval and command time limit are coupled, a command must have been checked
		// once already in order to be considered "long-lasting" and a generic OK heartbeat to be sent to the client.
		if s.currentCommand != nil && time.Since(s.currentCommand.startTime) > s.progressInterval {
			cmd = *s.currentCommand
			hasCmd = true
		}
	})

	if !hasCmd {
		return
	}

	if !hasThunderbird(cmd.imapID) && s.featureFlagProvider.GetFlagValue(featureflags.CommandWatcherNonThunderbirdDisabled) {
		return
	}

	log := s.log.WithField("cmd", cmd.sanitizedPayload)

	if err := s.sendProgressMessageFn(defaultProgressMessage); err != nil {
		log.WithError(err).Error("Failed to send progress message")
	}

	log.Info("Sent progress message")
}

type TrackCommandFn func(sanitizedPayload string) func()

func (s *Service) TrackedWithImapID(imapID imap.IMAPID) TrackCommandFn {
	return func(sanitizedPayload string) func() {
		return s.trackCommand(sanitizedPayload, imapID)
	}
}

func (s *Service) trackCommand(sanitizedPayload string, imapID imap.IMAPID) (cleanupFn func()) {
	s.withLock(func() {
		if s.currentCommand != nil {
			s.log.WithFields(logrus.Fields{
				"existing": s.currentCommand.sanitizedPayload,
				"new":      sanitizedPayload,
			}).Error("Tracking another command while one is still being tracked")
		}

		s.currentCommand = &commandMetadata{
			startTime:        time.Now(),
			sanitizedPayload: sanitizedPayload,
			imapID:           imapID,
		}
	})

	return func() {
		s.withLock(func() {
			s.currentCommand = nil
		})
	}
}

func hasThunderbird(imapID imap.IMAPID) bool {
	return strings.Contains(strings.ToLower(imapID.Name), thunderbirdIMAPName)
}
