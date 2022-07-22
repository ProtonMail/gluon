package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleUID(ctx context.Context, tag string, cmd *proto.UID, mailbox *backend.Mailbox, profiler profiling.CmdProfiler, ch chan response.Response) error {
	switch cmd := cmd.GetCommand().(type) {
	case *proto.UID_Copy:
		profiler.Start(profiling.CmdTypeUIDCopy)
		defer profiler.Stop(profiling.CmdTypeUIDCopy)

		return s.handleCopy(backend.AsUID(ctx), tag, cmd.Copy, mailbox, ch)

	case *proto.UID_Move:
		profiler.Start(profiling.CmdTypeUIDMove)
		defer profiler.Stop(profiling.CmdTypeUIDMove)

		return s.handleMove(backend.AsUID(ctx), tag, cmd.Move, mailbox, ch)

	case *proto.UID_Fetch:
		profiler.Start(profiling.CmdTypeUIDFetch)
		defer profiler.Stop(profiling.CmdTypeUIDFetch)

		return s.handleFetch(backend.AsUID(ctx), tag, cmd.Fetch, mailbox, ch)

	case *proto.UID_Search:
		profiler.Start(profiling.CmdTypeUIDSearch)
		defer profiler.Stop(profiling.CmdTypeUIDSearch)

		return s.handleSearch(backend.AsUID(ctx), tag, cmd.Search, mailbox, ch)

	case *proto.UID_Store:
		profiler.Start(profiling.CmdTypeUIDStore)
		defer profiler.Stop(profiling.CmdTypeUIDStore)

		return s.handleStore(backend.AsUID(ctx), tag, cmd.Store, mailbox, ch)

	default:
		panic("bad command")
	}
}
