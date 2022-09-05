package session

import (
	"context"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleDelete(ctx context.Context, tag string, cmd *proto.Del, ch chan response.Response) (response.Response, error) {
	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(nameUTF8, imap.Inbox) {
		return nil, ErrDeleteInbox
	}

	selectedDeleted, err := s.state.Delete(ctx, nameUTF8)
	if err != nil {
		reporter.MessageWithContext(ctx,
			"Failed to delete mailbox",
			reporter.Context{"error": err},
		)

		return nil, err
	}

	ch <- response.Ok(tag).WithMessage("DELETE")

	var rep response.Response

	if selectedDeleted {
		rep = response.Bye().WithMailboxDeleted()
	}

	return rep, nil
}
