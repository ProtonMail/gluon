package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/response"
)

func (s *Session) handleOther(ctx context.Context, cmd *IMAPCommand) chan response.Response {
	ch := make(chan response.Response)

	go func() {
		defer close(ch)

		if err := s.handleCommand(ctx, cmd, ch); err != nil {
			if res, ok := response.FromError(err); ok {
				ch <- res
			} else {
				ch <- response.No(cmd.tag).WithError(err)
			}
		}
	}()

	return ch
}

func (s *Session) handleCommand(ctx context.Context, cmd *IMAPCommand, ch chan response.Response) error {
	switch {
	case
		cmd.GetCapability() != nil,
		cmd.GetNoop() != nil:
		return s.handleAnyCommand(ctx, cmd, ch)

	case
		cmd.GetAuth() != nil,
		cmd.GetLogin() != nil:
		return s.handleNotAuthenticatedCommand(ctx, cmd, ch)

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
		return s.handleAuthenticatedCommand(ctx, cmd, ch)

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
		return s.handleSelectedCommand(ctx, cmd, ch)

	default:
		panic("bad command")
	}
}

func (s *Session) handleAnyCommand(ctx context.Context, cmd *IMAPCommand, ch chan response.Response) error {
	switch {
	case cmd.GetCapability() != nil:
		// 6.1.1 CAPABILITY Command
		return s.handleCapability(ctx, cmd.tag, cmd.GetCapability(), ch)

	case cmd.GetNoop() != nil:
		// 6.1.2 NOOP Command
		return s.handleNoop(ctx, cmd.tag, cmd.GetNoop(), ch)

	default:
		panic("bad command")
	}
}

func (s *Session) handleNotAuthenticatedCommand(ctx context.Context, cmd *IMAPCommand, ch chan response.Response) error {
	switch {
	case cmd.GetAuth() != nil:
		// 6.2.2. AUTHENTICATE Command
		return ErrNotImplemented

	case cmd.GetLogin() != nil:
		// 6.2.3. LOGIN Command
		return s.handleLogin(ctx, cmd.tag, cmd.GetLogin(), ch)

	default:
		panic("bad command")
	}
}

func (s *Session) handleAuthenticatedCommand(ctx context.Context, cmd *IMAPCommand, ch chan response.Response) error {
	s.userLock.Lock()
	defer s.userLock.Unlock()

	if s.state == nil {
		return ErrNotAuthenticated
	}

	switch {
	case cmd.GetSelect() != nil:
		// 6.3.1. SELECT Command
		return s.handleSelect(ctx, cmd.tag, cmd.GetSelect(), ch)

	case cmd.GetExamine() != nil:
		// 6.3.2. EXAMINE Command
		return s.handleExamine(ctx, cmd.tag, cmd.GetExamine(), ch)

	case cmd.GetCreate() != nil:
		// 6.3.3. CREATE Command
		return s.handleCreate(ctx, cmd.tag, cmd.GetCreate(), ch)

	case cmd.GetDel() != nil:
		// 6.3.4. DELETE Command
		return s.handleDelete(ctx, cmd.tag, cmd.GetDel(), ch)

	case cmd.GetRename() != nil:
		// 6.3.5. RENAME Command
		return s.handleRename(ctx, cmd.tag, cmd.GetRename(), ch)

	case cmd.GetSub() != nil:
		// 6.3.6. SUBSCRIBE Command
		return s.handleSub(ctx, cmd.tag, cmd.GetSub(), ch)

	case cmd.GetUnsub() != nil:
		// 6.3.7. UNSUBSCRIBE Command
		return s.handleUnsub(ctx, cmd.tag, cmd.GetUnsub(), ch)

	case cmd.GetList() != nil:
		// 6.3.8. LIST Command
		return s.handleList(ctx, cmd.tag, cmd.GetList(), ch)

	case cmd.GetLsub() != nil:
		// 6.3.9. Lsub Command
		return s.handleLsub(ctx, cmd.tag, cmd.GetLsub(), ch)

	case cmd.GetStatus() != nil:
		// 6.3.10. STATUS Command
		return s.handleStatus(ctx, cmd.tag, cmd.GetStatus(), ch)

	case cmd.GetAppend() != nil:
		// 6.3.11. APPEND Command
		return s.handleAppend(ctx, cmd.tag, cmd.GetAppend(), ch)

	default:
		panic("bad command")
	}
}

func (s *Session) handleSelectedCommand(ctx context.Context, cmd *IMAPCommand, ch chan response.Response) error {
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

		if err := s.handleWithMailbox(ctx, cmd, mailbox, ch); err != nil {
			return err
		}

		return flush(ctx, mailbox, false, ch)
	})
}

func (s *Session) handleWithMailbox(ctx context.Context, cmd *IMAPCommand, mailbox *backend.Mailbox, ch chan response.Response) error {
	switch {
	case cmd.GetCheck() != nil:
		// 6.4.1. CHECK Command
		return s.handleCheck(ctx, cmd.tag, cmd.GetCheck(), mailbox, ch)

	case cmd.GetClose() != nil:
		// 6.4.2. CLOSE Command
		return s.handleClose(ctx, cmd.tag, cmd.GetClose(), mailbox, ch)

	case cmd.GetExpunge() != nil:
		// 6.4.3. EXPUNGE Command
		return s.handleExpunge(ctx, cmd.tag, cmd.GetExpunge(), mailbox, ch)

	case cmd.GetUidExpunge() != nil:
		// RFC4315 UIDPLUS Extension
		return s.handleUIDExpunge(ctx, cmd.tag, cmd.GetUidExpunge(), mailbox, ch)

	case cmd.GetUnselect() != nil:
		// RFC3691 UNSELECT Extension
		return s.handleUnselect(ctx, cmd.tag, cmd.GetUnselect(), mailbox, ch)

	case cmd.GetSearch() != nil:
		// 6.4.4. SEARCH Command
		return s.handleSearch(ctx, cmd.tag, cmd.GetSearch(), mailbox, ch)

	case cmd.GetFetch() != nil:
		// 6.4.5. FETCH Command
		return s.handleFetch(ctx, cmd.tag, cmd.GetFetch(), mailbox, ch)

	case cmd.GetStore() != nil:
		// 6.4.6. STORE Command
		return s.handleStore(ctx, cmd.tag, cmd.GetStore(), mailbox, ch)

	case cmd.GetCopy() != nil:
		// 6.4.7. COPY Command
		return s.handleCopy(ctx, cmd.tag, cmd.GetCopy(), mailbox, ch)

	case cmd.GetUid() != nil:
		// 6.4.8. UID Command
		return s.handleUID(ctx, cmd.tag, cmd.GetUid(), mailbox, ch)

	case cmd.GetMove() != nil:
		// RFC6851 MOVE Command
		return s.handleMove(ctx, cmd.tag, cmd.GetMove(), mailbox, ch)

	default:
		panic("bad command")
	}
}
