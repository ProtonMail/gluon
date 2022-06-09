package session

import (
	"context"

	"github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
)

func (s *Session) handleLogin(ctx context.Context, tag string, cmd *proto.Login, ch chan response.Response) error {
	s.userLock.Lock()
	defer s.userLock.Unlock()

	s.capsLock.Lock()
	defer s.capsLock.Unlock()

	// If already authenticated, return BAD (it seems that NO is reserved for login failures).
	// This matches the behaviour of dovecot and gmail.
	if s.state != nil {
		return response.Bad(tag).WithError(ErrAlreadyAuthenticated)
	}

	state, err := s.backend.GetState(cmd.GetUsername(), cmd.GetPassword(), s.sessionID)
	if err != nil {
		return err
	}

	s.state = state

	ch <- response.Ok(tag).WithItems(response.ItemCapability(s.caps...))

	s.eventCh <- events.EventLogin{
		SessionID: s.sessionID,
		UserID:    state.UserID(),
	}

	return nil
}
