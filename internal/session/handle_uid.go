package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
)

func (s *Session) handleUID(ctx context.Context, tag string, cmd *proto.UID, mailbox *state.Mailbox, ch chan response.Response) (response.Response, error) {
	switch cmd := cmd.GetCommand().(type) {
	case *proto.UID_Copy:
		return s.handleCopy(contexts.AsUID(ctx), tag, cmd.Copy, mailbox, ch)

	case *proto.UID_Move:
		return s.handleMove(contexts.AsUID(ctx), tag, cmd.Move, mailbox, ch)

	case *proto.UID_Fetch:
		return s.handleFetch(contexts.AsUID(ctx), tag, cmd.Fetch, mailbox, ch)

	case *proto.UID_Search:
		return s.handleSearch(contexts.AsUID(ctx), tag, cmd.Search, mailbox, ch)

	case *proto.UID_Store:
		return s.handleStore(contexts.AsUID(ctx), tag, cmd.Store, mailbox, ch)

	default:
		panic("bad command")
	}
}
