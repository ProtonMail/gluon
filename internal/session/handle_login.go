package session

import (
	"context"

	"github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleLogin(ctx context.Context, tag string, cmd *proto.Login, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeLogin)
	defer profiling.Stop(ctx, profiling.CmdTypeLogin)

	s.userLock.Lock()
	defer s.userLock.Unlock()

	s.capsLock.Lock()
	defer s.capsLock.Unlock()

	// If already authenticated, return BAD (it seems that NO is reserved for login failures).
	// This matches the behaviour of dovecot and gmail.
	if s.state != nil {
		return response.Bad(tag).WithError(ErrAlreadyAuthenticated)
	}

	state, err := s.backend.GetState(ctx, cmd.GetUsername(), cmd.GetPassword(), s.sessionID)
	if err != nil {
		s.eventCh <- events.LoginFailed{
			SessionID: s.sessionID,
			Username:  cmd.GetUsername(),
		}

		return err
	}

	s.state = state

	ch <- response.Ok(tag).WithItems(response.ItemCapability(s.caps...)).WithMessage("Logged in")

	s.eventCh <- events.Login{
		SessionID: s.sessionID,
		UserID:    state.UserID(),
	}

	// We set the IMAP ID extension value after login, since it's possible that the client may have sent it before.
	// This ensures that the ID is correctly set for the connection.
	state.SetConnMetadataKeyValue(imap.IMAPIDConnMetadataKey, s.imapID)

	return nil
}
