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

func (s *Session) handleCreate(ctx context.Context, tag string, cmd *proto.Create, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeCreate)
	defer profiling.Stop(ctx, profiling.CmdTypeCreate)

	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	if strings.EqualFold(nameUTF8, imap.Inbox) {
		return ErrCreateInbox
	}

	if err := s.state.Create(ctx, nameUTF8); err != nil {
		if shouldReportIMAPCommandError(err) {
			reporter.MessageWithContext(ctx,
				"Failed to create mailbox",
				reporter.Context{"error": err, "mailbox": nameUTF8},
			)
		}

		return err
	}

	ch <- response.Ok(tag).WithMessage("CREATE")

	return nil
}
