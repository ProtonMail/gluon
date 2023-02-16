package session

import (
	"context"
	"github.com/ProtonMail/gluon/imap/command"

	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleLogout(ctx context.Context, tag string, _ *command.Logout) error {
	profiling.Start(ctx, profiling.CmdTypeLogout)
	defer profiling.Stop(ctx, profiling.CmdTypeLogout)

	s.userLock.Lock()
	defer s.userLock.Unlock()

	s.capsLock.Lock()
	defer s.capsLock.Unlock()

	if err := response.Bye().Send(s); err != nil {
		return err
	}

	if err := response.Ok(tag).WithMessage("LOGOUT").Send(s); err != nil {
		return err
	}

	return nil
}
