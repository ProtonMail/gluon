package session

import (
	"context"

	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleUnsub(ctx context.Context, tag string, cmd *command.Unsubscribe, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeUnsubscribe)
	defer profiling.Stop(ctx, profiling.CmdTypeUnsubscribe)

	nameUTF8, err := s.decodeMailboxName(cmd.Mailbox)
	if err != nil {
		return err
	}

	if err := s.state.Unsubscribe(ctx, nameUTF8); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithMessage("UNSUBSCRIBE")

	return nil
}
