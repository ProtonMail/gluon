package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
)

func (s *Session) handleUID(ctx context.Context, tag string, cmd *proto.UID, mailbox *backend.Mailbox, ch chan response.Response) error {
	switch cmd := cmd.GetCommand().(type) {
	case *proto.UID_Copy:
		return s.handleCopy(backend.AsUID(ctx), tag, cmd.Copy, mailbox, ch)

	case *proto.UID_Move:
		return s.handleMove(backend.AsUID(ctx), tag, cmd.Move, mailbox, ch)

	case *proto.UID_Fetch:
		return s.handleFetch(backend.AsUID(ctx), tag, cmd.Fetch, mailbox, ch)

	case *proto.UID_Search:
		return s.handleSearch(backend.AsUID(ctx), tag, cmd.Search, mailbox, ch)

	case *proto.UID_Store:
		return s.handleStore(backend.AsUID(ctx), tag, cmd.Store, mailbox, ch)

	default:
		panic("bad command")
	}
}
