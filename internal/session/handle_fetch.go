package session

import (
	"context"
	"errors"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
)

func (s *Session) handleFetch(ctx context.Context, tag string, cmd *proto.Fetch, mailbox *state.Mailbox, ch chan response.Response) error {
	if err := mailbox.Fetch(ctx, cmd.GetSequenceSet(), cmd.GetAttributes(), ch); errors.Is(err, state.ErrNoSuchMessage) {
		return response.Bad(tag).WithError(err)
	} else if err != nil {
		return err
	}

	var items []response.Item

	if mailbox.ExpungeIssued() {
		items = append(items, response.ItemExpungeIssued())
	}

	ch <- response.Ok(tag).
		WithItems(items...).
		WithMessage(okMessage(ctx))

	return nil
}
