package session

import (
	"context"
	"runtime/pprof"
	"strconv"

	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleOther(
	ctx context.Context,
	tag string,
	cmd *proto.Command,
	profiler profiling.CmdProfiler,
) chan response.Response {
	ch := make(chan response.Response, channelBufferCount)

	go func() {
		labels := pprof.Labels("go", "handleOther()", "SessionID", strconv.Itoa(s.sessionID))
		pprof.Do(ctx, labels, func(_ context.Context) {
			defer close(ch)

			ctx := backend.NewStateContext(ctx, s.state)

			if err := s.handleCommand(ctx, tag, cmd, ch, profiler); err != nil {
				if res, ok := response.FromError(err); ok {
					ch <- res
				} else {
					ch <- response.No(tag).WithError(err)
				}
			}
		})
	}()

	return ch
}

func (s *Session) handleCommand(
	ctx context.Context,
	tag string,
	cmd *proto.Command,
	ch chan response.Response,
	profiler profiling.CmdProfiler,
) error {
	switch {
	case
		cmd.GetCapability() != nil,
		cmd.GetIdGet() != nil,
		cmd.GetIdSet() != nil,
		cmd.GetNoop() != nil:
		return s.handleAnyCommand(ctx, tag, cmd, ch, profiler)

	case
		cmd.GetAuth() != nil,
		cmd.GetLogin() != nil:
		return s.handleNotAuthenticatedCommand(ctx, tag, cmd, ch, profiler)

	case
		cmd.GetSelect() != nil,
		cmd.GetExamine() != nil,
		cmd.GetCreate() != nil,
		cmd.GetDel() != nil,
		cmd.GetRename() != nil,
		cmd.GetSub() != nil,
		cmd.GetUnsub() != nil,
		cmd.GetList() != nil,
		cmd.GetLsub() != nil,
		cmd.GetStatus() != nil,
		cmd.GetAppend() != nil:
		return s.handleAuthenticatedCommand(ctx, tag, cmd, ch, profiler)
	case
		cmd.GetCheck() != nil,
		cmd.GetClose() != nil,
		cmd.GetExpunge() != nil,
		cmd.GetUidExpunge() != nil,
		cmd.GetUnselect() != nil,
		cmd.GetSearch() != nil,
		cmd.GetFetch() != nil,
		cmd.GetStore() != nil,
		cmd.GetCopy() != nil,
		cmd.GetMove() != nil,
		cmd.GetUid() != nil:
		return s.handleSelectedCommand(ctx, tag, cmd, ch, profiler)

	default:
		panic("bad command")
	}
}

func (s *Session) handleAnyCommand(
	ctx context.Context,
	tag string,
	cmd *proto.Command,
	ch chan response.Response,
	profiler profiling.CmdProfiler,
) error {
	switch {
	case cmd.GetCapability() != nil:
		// 6.1.1 CAPABILITY Command
		return s.handleCapability(ctx, tag, cmd.GetCapability(), ch)

	case cmd.GetNoop() != nil:
		profiler.Start(profiling.CmdTypeNoop)
		defer profiler.Stop(profiling.CmdTypeNoop)
		// 6.1.2 NOOP Command
		return s.handleNoop(ctx, tag, cmd.GetNoop(), ch)

	case cmd.GetIdSet() != nil:
		profiler.Start(profiling.CmdTypeID)
		defer profiler.Stop(profiling.CmdTypeID)
		// RFC 2971 ID
		return s.handleIDSet(ctx, tag, cmd.GetIdSet(), ch)
	case cmd.GetIdGet() != nil:
		profiler.Start(profiling.CmdTypeID)
		defer profiler.Stop(profiling.CmdTypeID)
		// RFC 2971 ID
		return s.handleIDGet(ctx, tag, ch)

	default:
		panic("bad command")
	}
}

func (s *Session) handleNotAuthenticatedCommand(
	ctx context.Context,
	tag string,
	cmd *proto.Command,
	ch chan response.Response,
	profiler profiling.CmdProfiler,
) error {
	switch {
	case cmd.GetAuth() != nil:
		// 6.2.2. AUTHENTICATE Command
		return ErrNotImplemented

	case cmd.GetLogin() != nil:
		profiler.Start(profiling.CmdTypeLogin)
		defer profiler.Stop(profiling.CmdTypeLogin)
		// 6.2.3. LOGIN Command
		return s.handleLogin(ctx, tag, cmd.GetLogin(), ch)

	default:
		panic("bad command")
	}
}

func (s *Session) handleAuthenticatedCommand(
	ctx context.Context,
	tag string,
	cmd *proto.Command,
	ch chan response.Response,
	profiler profiling.CmdProfiler,
) error {
	s.userLock.Lock()
	defer s.userLock.Unlock()

	if s.state == nil {
		return ErrNotAuthenticated
	}

	switch {
	case cmd.GetSelect() != nil:
		profiler.Start(profiling.CmdTypeSelect)
		defer profiler.Stop(profiling.CmdTypeSelect)
		// 6.3.1. SELECT Command
		return s.handleSelect(ctx, tag, cmd.GetSelect(), ch)

	case cmd.GetExamine() != nil:
		profiler.Start(profiling.CmdTypeExamine)
		defer profiler.Stop(profiling.CmdTypeExamine)
		// 6.3.2. EXAMINE Command
		return s.handleExamine(ctx, tag, cmd.GetExamine(), ch)

	case cmd.GetCreate() != nil:
		profiler.Start(profiling.CmdTypeCreate)
		defer profiler.Stop(profiling.CmdTypeCreate)
		// 6.3.3. CREATE Command
		return s.handleCreate(ctx, tag, cmd.GetCreate(), ch)

	case cmd.GetDel() != nil:
		profiler.Start(profiling.CmdTypeDelete)
		defer profiler.Stop(profiling.CmdTypeDelete)
		// 6.3.4. DELETE Command
		return s.handleDelete(ctx, tag, cmd.GetDel(), ch)

	case cmd.GetRename() != nil:
		profiler.Start(profiling.CmdTypeRename)
		defer profiler.Stop(profiling.CmdTypeRename)
		// 6.3.5. RENAME Command
		return s.handleRename(ctx, tag, cmd.GetRename(), ch)

	case cmd.GetSub() != nil:
		profiler.Start(profiling.CmdTypeSubscribe)
		defer profiler.Stop(profiling.CmdTypeSubscribe)
		// 6.3.6. SUBSCRIBE Command
		return s.handleSub(ctx, tag, cmd.GetSub(), ch)

	case cmd.GetUnsub() != nil:
		profiler.Start(profiling.CmdTypeUnsubscribe)
		defer profiler.Stop(profiling.CmdTypeUnsubscribe)
		// 6.3.7. UNSUBSCRIBE Command
		return s.handleUnsub(ctx, tag, cmd.GetUnsub(), ch)

	case cmd.GetList() != nil:
		profiler.Start(profiling.CmdTypeList)
		defer profiler.Stop(profiling.CmdTypeList)
		// 6.3.8. LIST Command
		return s.handleList(ctx, tag, cmd.GetList(), ch)

	case cmd.GetLsub() != nil:
		profiler.Start(profiling.CmdTypeLSub)
		defer profiler.Stop(profiling.CmdTypeLSub)
		// 6.3.9. Lsub Command
		return s.handleLsub(ctx, tag, cmd.GetLsub(), ch)

	case cmd.GetStatus() != nil:
		profiler.Start(profiling.CmdTypeStatus)
		defer profiler.Stop(profiling.CmdTypeStatus)
		// 6.3.10. STATUS Command
		return s.handleStatus(ctx, tag, cmd.GetStatus(), ch)

	case cmd.GetAppend() != nil:
		profiler.Start(profiling.CmdTypeAppend)
		defer profiler.Stop(profiling.CmdTypeAppend)
		// 6.3.11. APPEND Command
		return s.handleAppend(ctx, tag, cmd.GetAppend(), ch)

	default:
		panic("bad command")
	}
}

func (s *Session) handleSelectedCommand(
	ctx context.Context,
	tag string,
	cmd *proto.Command,
	ch chan response.Response,
	profiler profiling.CmdProfiler,
) error {
	s.userLock.Lock()
	defer s.userLock.Unlock()

	if s.state == nil {
		return ErrNotAuthenticated
	}

	// TODO(REFACTOR): Should we flush both before and after?
	return s.state.Selected(ctx, func(mailbox *backend.Mailbox) error {
		if err := flush(ctx, mailbox, false, ch); err != nil {
			return err
		}

		if err := s.handleWithMailbox(ctx, tag, cmd, mailbox, ch, profiler); err != nil {
			return err
		}

		return flush(ctx, mailbox, false, ch)
	})
}

func (s *Session) handleWithMailbox(
	ctx context.Context,
	tag string,
	cmd *proto.Command,
	mailbox *backend.Mailbox,
	ch chan response.Response,
	profiler profiling.CmdProfiler,
) error {
	switch {
	case cmd.GetCheck() != nil:
		profiler.Start(profiling.CmdTypeCheck)
		defer profiler.Stop(profiling.CmdTypeCheck)
		// 6.4.1. CHECK Command
		return s.handleCheck(ctx, tag, cmd.GetCheck(), mailbox, ch)

	case cmd.GetClose() != nil:
		profiler.Start(profiling.CmdTypeClose)
		defer profiler.Stop(profiling.CmdTypeClose)
		// 6.4.2. CLOSE Command
		return s.handleClose(ctx, tag, cmd.GetClose(), mailbox, ch)

	case cmd.GetExpunge() != nil:
		profiler.Start(profiling.CmdTypeExpunge)
		defer profiler.Stop(profiling.CmdTypeExpunge)
		// 6.4.3. EXPUNGE Command
		return s.handleExpunge(ctx, tag, cmd.GetExpunge(), mailbox, ch)

	case cmd.GetUidExpunge() != nil:
		profiler.Start(profiling.CmdTypeExpunge)
		defer profiler.Stop(profiling.CmdTypeExpunge)
		// RFC4315 UIDPLUS Extension
		return s.handleUIDExpunge(ctx, tag, cmd.GetUidExpunge(), mailbox, ch)

	case cmd.GetUnselect() != nil:
		profiler.Start(profiling.CmdTypeUnselect)
		defer profiler.Stop(profiling.CmdTypeUnselect)
		// RFC3691 UNSELECT Extension
		return s.handleUnselect(ctx, tag, cmd.GetUnselect(), mailbox, ch)

	case cmd.GetSearch() != nil:
		profiler.Start(profiling.CmdTypeSearch)
		defer profiler.Stop(profiling.CmdTypeSearch)
		// 6.4.4. SEARCH Command
		return s.handleSearch(ctx, tag, cmd.GetSearch(), mailbox, ch)

	case cmd.GetFetch() != nil:
		profiler.Start(profiling.CmdTypeFetch)
		defer profiler.Stop(profiling.CmdTypeFetch)
		// 6.4.5. FETCH Command
		return s.handleFetch(ctx, tag, cmd.GetFetch(), mailbox, ch)

	case cmd.GetStore() != nil:
		profiler.Start(profiling.CmdTypeStore)
		defer profiler.Stop(profiling.CmdTypeStore)
		// 6.4.6. STORE Command
		return s.handleStore(ctx, tag, cmd.GetStore(), mailbox, ch)

	case cmd.GetCopy() != nil:
		profiler.Start(profiling.CmdTypeCopy)
		defer profiler.Stop(profiling.CmdTypeCopy)
		// 6.4.7. COPY Command
		return s.handleCopy(ctx, tag, cmd.GetCopy(), mailbox, ch)

	case cmd.GetUid() != nil:
		// 6.4.8. UID Command
		return s.handleUID(ctx, tag, cmd.GetUid(), mailbox, profiler, ch)

	case cmd.GetMove() != nil:
		profiler.Start(profiling.CmdTypeMove)
		defer profiler.Stop(profiling.CmdTypeMove)
		// RFC6851 MOVE Command
		return s.handleMove(ctx, tag, cmd.GetMove(), mailbox, ch)

	default:
		panic("bad command")
	}
}
