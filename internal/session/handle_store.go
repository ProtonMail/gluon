package session

import (
	"context"
	"errors"

	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/observability"
	"github.com/ProtonMail/gluon/observability/metrics"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleStore(ctx context.Context, tag string, cmd *command.Store, mailbox *state.Mailbox, ch chan response.Response) (response.Response, error) {
	if contexts.IsUID(ctx) {
		profiling.Start(ctx, profiling.CmdTypeUIDStore)
		defer profiling.Stop(ctx, profiling.CmdTypeUIDStore)
	} else {
		profiling.Start(ctx, profiling.CmdTypeStore)
		defer profiling.Stop(ctx, profiling.CmdTypeStore)
	}

	if cmd.Silent {
		ctx = contexts.AsSilent(ctx)
	}

	if mailbox.ReadOnly() {
		return nil, ErrReadOnly
	}

	flags, err := validateStoreFlags(cmd.Flags)
	if err != nil {
		return response.Bad(tag).WithError(err), nil
	}

	if err := mailbox.Store(ctx, cmd.SeqSet, cmd.Action, flags); errors.Is(err, state.ErrNoSuchMessage) {
		return response.Bad(tag).WithError(err), nil
	} else if err != nil {
		// A result of either a failed request (API unreachable), or the message does not exist on remote.
		observability.AddMessageRelatedMetric(ctx, metrics.GenerateFailedToStoreFlagsOnMessages())

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
