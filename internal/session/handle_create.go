package session

import (
	"context"
	"github.com/ProtonMail/gluon/imap/command"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/ProtonMail/gluon/reporter"
)

func (s *Session) handleCreate(ctx context.Context, tag string, cmd *command.Create, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeCreate)
	defer profiling.Stop(ctx, profiling.CmdTypeCreate)

	nameUTF8, err := s.decodeMailboxName(cmd.Mailbox)
	if err != nil {
		return err
	}

	if strings.EqualFold(nameUTF8, imap.Inbox) {
		return ErrCreateInbox
	}

	if err := s.state.Create(ctx, nameUTF8); err != nil {
		reporter.MessageWithContext(ctx,
			"Failed to create mailbox",
			reporter.Context{"error": err},
		)

		return err
	}

	ch <- response.Ok(tag).WithMessage("CREATE")

	return nil
}
