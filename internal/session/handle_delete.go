package session

import (
	"context"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleDelete(ctx context.Context, tag string, cmd *proto.Del, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeDelete)
	defer profiling.Stop(ctx, profiling.CmdTypeDelete)

	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	if strings.EqualFold(nameUTF8, imap.Inbox) {
		return ErrDeleteInbox
	}

	selectedDeleted, err := s.state.Delete(ctx, nameUTF8)
	if err != nil {
		if shouldReportIMAPCommandError(err) {
			reporter.MessageWithContext(ctx,
				"Failed to delete mailbox",
				reporter.Context{"error": err, "mailbox": nameUTF8},
			)
		}

		return err
	}

	ch <- response.Ok(tag).WithMessage("DELETE")

	if selectedDeleted {
		ch <- response.Bye().WithMailboxDeleted()
	}

	return nil
}
