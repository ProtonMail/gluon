package session

import (
	"context"
	"github.com/ProtonMail/gluon/imap/command"

	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
)

func (s *Session) handleUID(ctx context.Context, tag string, cmd *command.UID, mailbox *state.Mailbox, ch chan response.Response) (response.Response, error) {
	switch cmd := cmd.Command.(type) {
	case *command.Copy:
		return s.handleCopy(contexts.AsUID(ctx), tag, cmd, mailbox, ch)

	case *command.Move:
		return s.handleMove(contexts.AsUID(ctx), tag, cmd, mailbox, ch)

	case *command.Fetch:
		return s.handleFetch(contexts.AsUID(ctx), tag, cmd, mailbox, ch)

	case *command.Search:
		return s.handleSearch(contexts.AsUID(ctx), tag, cmd, mailbox, ch)

	case *command.Store:
		return s.handleStore(contexts.AsUID(ctx), tag, cmd, mailbox, ch)

	default:
		panic("bad command")
	}
}
