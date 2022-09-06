package session

import (
	"context"
	"errors"

	context2 "github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/reporter"
)

func (s *Session) handleStore(ctx context.Context, tag string, cmd *proto.Store, mailbox *state.Mailbox, ch chan response.Response) (response.Response, error) {
	if cmd.GetAction().GetSilent() {
		ctx = context2.AsSilent(ctx)
	}

	flags, err := validateStoreFlags(cmd.GetFlags())
	if err != nil {
		return response.Bad(tag).WithError(err), nil
	}

	if err := mailbox.Store(ctx, cmd.GetSequenceSet(), cmd.GetAction().GetOperation(), flags); errors.Is(err, state.ErrNoSuchMessage) {
		return response.Bad(tag).WithError(err), nil
	} else if err != nil {
		reporter.MessageWithContext(ctx,
			"Failed to store flags on messages",
			reporter.Context{"error": err},
		)

		return nil, err
	}

	if err := flush(ctx, mailbox, false, ch); err != nil {
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
