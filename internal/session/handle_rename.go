package session

import (
	"context"
	"github.com/ProtonMail/gluon/imap/command"

	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleRename(ctx context.Context, tag string, cmd *command.Rename, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeRename)
	defer profiling.Stop(ctx, profiling.CmdTypeRename)

	oldNameUTF8, err := s.decodeMailboxName(cmd.From)
	if err != nil {
		return err
	}

	newNameUTF8, err := s.decodeMailboxName(cmd.To)
	if err != nil {
		return err
	}

	if err := s.state.Rename(ctx, oldNameUTF8, newNameUTF8); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithMessage("RENAME")

	return nil
}
