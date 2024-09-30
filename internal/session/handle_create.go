package session

import (
	"context"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/observability"
	"github.com/ProtonMail/gluon/observability/metrics"
	"github.com/ProtonMail/gluon/profiling"
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
		observability.AddMessageRelatedMetric(ctx, metrics.GenerateFailedToCreateMailbox())
		return err
	}

	ch <- response.Ok(tag).WithMessage("CREATE")

	return nil
}
