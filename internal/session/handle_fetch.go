package session

import (
	"context"
	"errors"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/reporter"
)

func (s *Session) handleFetch(ctx context.Context, tag string, cmd *proto.Fetch, mailbox *state.Mailbox, ch chan response.Response) (response.Response, error) {
	if err := mailbox.Fetch(ctx, cmd.GetSequenceSet(), cmd.GetAttributes(), ch); errors.Is(err, state.ErrNoSuchMessage) {
		return response.Bad(tag).WithError(err), nil
	} else if err != nil {
		reporter.MessageWithContext(ctx,
			"Failed to fetch messages",
			reporter.Context{"error": err},
		)

		return nil, err
	}

	var items []response.Item

	if mailbox.ExpungeIssued() {
		items = append(items, response.ItemExpungeIssued())
	}

	return response.Ok(tag).
		WithItems(items...).
		WithMessage(okMessage(ctx)), nil
}
