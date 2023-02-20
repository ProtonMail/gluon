package session

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/response"
)

func (s *Session) getCaps() []imap.Capability {
	s.userLock.Lock()
	defer s.userLock.Unlock()

	if s.state != nil {
		return s.caps
	}

	caps := []imap.Capability{}
	for _, c := range s.caps {
		if imap.IsCapabilityAvailableBeforeAuth(c) {
			caps = append(caps, c)
		}
	}

	return caps
}

func (s *Session) handleCapability(_ context.Context, tag string, _ *command.Capability, ch chan response.Response) error {
	s.capsLock.Lock()
	defer s.capsLock.Unlock()

	ch <- response.Capability().WithCapabilities(s.getCaps()...)

	ch <- response.Ok(tag).WithMessage("CAPABILITY")

	return nil
}
