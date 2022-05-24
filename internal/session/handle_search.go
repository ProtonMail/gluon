package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
)

func (s *Session) handleSearch(ctx context.Context, tag string, cmd *proto.Search, mailbox *backend.Mailbox, ch chan response.Response) error {
	seq, err := mailbox.Search(ctx, cmd.GetKeys())
	if err != nil {
		return err
	}

	ch <- response.Search(seq...)

	var items []response.Item

	if mailbox.ExpungeIssued() {
		items = append(items, response.ItemExpungeIssued())
	}

	ch <- response.Ok(tag).
		WithItems(items...).
		WithMessage(okMessage(ctx))

	return nil
}
