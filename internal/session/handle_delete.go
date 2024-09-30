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

func (s *Session) handleDelete(ctx context.Context, tag string, cmd *command.Delete, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeDelete)
	defer profiling.Stop(ctx, profiling.CmdTypeDelete)

	nameUTF8, err := s.decodeMailboxName(cmd.Mailbox)
	if err != nil {
		return err
	}

	if strings.EqualFold(nameUTF8, imap.Inbox) {
		return ErrDeleteInbox
	}

	selectedDeleted, err := s.state.Delete(ctx, nameUTF8)
	if err != nil {
		observability.AddOtherMetric(ctx, metrics.GenerateFailedToDeleteMailboxMetric())
		return err
	}

	ch <- response.Ok(tag).WithMessage("DELETE")

	if selectedDeleted {
		ch <- response.Bye().WithMailboxDeleted()
	}

	return nil
}
