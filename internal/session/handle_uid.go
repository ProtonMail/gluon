package session

import (
	"context"

	context2 "github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleUID(ctx context.Context, tag string, cmd *proto.UID, mailbox *state.Mailbox, profiler profiling.CmdProfiler, ch chan response.Response) error {
	switch cmd := cmd.GetCommand().(type) {
	case *proto.UID_Copy:
		profiler.Start(profiling.CmdTypeUIDCopy)
		defer profiler.Stop(profiling.CmdTypeUIDCopy)

		return s.handleCopy(context2.AsUID(ctx), tag, cmd.Copy, mailbox, ch)

	case *proto.UID_Move:
		profiler.Start(profiling.CmdTypeUIDMove)
		defer profiler.Stop(profiling.CmdTypeUIDMove)

		return s.handleMove(context2.AsUID(ctx), tag, cmd.Move, mailbox, ch)

	case *proto.UID_Fetch:
		profiler.Start(profiling.CmdTypeUIDFetch)
		defer profiler.Stop(profiling.CmdTypeUIDFetch)

		return s.handleFetch(context2.AsUID(ctx), tag, cmd.Fetch, mailbox, ch)

	case *proto.UID_Search:
		profiler.Start(profiling.CmdTypeUIDSearch)
		defer profiler.Stop(profiling.CmdTypeUIDSearch)

		return s.handleSearch(context2.AsUID(ctx), tag, cmd.Search, mailbox, ch)

	case *proto.UID_Store:
		profiler.Start(profiling.CmdTypeUIDStore)
		defer profiler.Stop(profiling.CmdTypeUIDStore)

		return s.handleStore(context2.AsUID(ctx), tag, cmd.Store, mailbox, ch)

	default:
		panic("bad command")
	}
}
