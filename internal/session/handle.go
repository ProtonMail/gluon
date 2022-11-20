package session

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/gluon/queue"
)

func (s *Session) handleOther(
	ctx context.Context,
	tag string,
	cmd *proto.Command,
) <-chan response.Response {
	resCh := make(chan response.Response)

	outCh := queue.NewQueuedChannel[response.Response](0, 0)

	go func() {
		defer outCh.Close()

		for res := range resCh {
			outCh.Enqueue(res)
		}
	}()

	s.handleWG.Go(func() {
		logging.DoAnnotated(state.NewStateContext(ctx, s.state), func(ctx context.Context) {
			defer close(resCh)

			if err := s.handleCommand(ctx, tag, cmd, resCh); err != nil {
				if res, ok := response.FromError(err); ok {
					resCh <- res
				} else {
					resCh <- response.No(tag).WithError(err)
				}
			}
		}, logging.Labels{
			"Action":    "Handling IMAP command",
			"SessionID": s.sessionID,
		})
	})

	return outCh.GetChannel()
}

// handleCommand returns a response instance if a command needs to force an exit of the client.
func (s *Session) handleCommand(
	ctx context.Context,
	tag string,
	cmd *proto.Command,
	ch chan response.Response,
) error {
	switch {
	case
		cmd.GetCapability() != nil,
		cmd.GetIdGet() != nil,
		cmd.GetIdSet() != nil,
		cmd.GetNoop() != nil:
		return s.handleAnyCommand(ctx, tag, cmd, ch)

	case
		cmd.GetAuth() != nil,
		cmd.GetLogin() != nil:
		return s.handleNotAuthenticatedCommand(ctx, tag, cmd, ch)

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
		return s.handleAuthenticatedCommand(ctx, tag, cmd, ch)
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
		return s.handleSelectedCommand(ctx, tag, cmd, ch)

	default:
		return fmt.Errorf("bad command")
	}
}

func (s *Session) handleAnyCommand(
	ctx context.Context,
	tag string,
	cmd *proto.Command,
	ch chan response.Response,
) error {
	switch {
	case cmd.GetCapability() != nil:
		// 6.1.1 CAPABILITY Command
		return s.handleCapability(ctx, tag, cmd.GetCapability(), ch)

	case cmd.GetNoop() != nil:
		// 6.1.2 NOOP Command
		return s.handleNoop(ctx, tag, cmd.GetNoop(), ch)

	case cmd.GetIdSet() != nil:
		// RFC 2971 ID
		return s.handleIDSet(ctx, tag, cmd.GetIdSet(), ch)

	case cmd.GetIdGet() != nil:
		// RFC 2971 ID
		return s.handleIDGet(ctx, tag, ch)

	default:
		return fmt.Errorf("bad command")
	}
}

func (s *Session) handleNotAuthenticatedCommand(
	ctx context.Context,
	tag string,
	cmd *proto.Command,
	ch chan response.Response,
) error {
	switch {
	case cmd.GetAuth() != nil:
		// 6.2.2. AUTHENTICATE Command
		return ErrNotImplemented

	case cmd.GetLogin() != nil:
		// 6.2.3. LOGIN Command
		return s.handleLogin(ctx, tag, cmd.GetLogin(), ch)

	default:
		return fmt.Errorf("bad command")
	}
}

func (s *Session) handleAuthenticatedCommand(
	ctx context.Context,
	tag string,
	cmd *proto.Command,
	ch chan response.Response,
) error {
	s.userLock.Lock()
	defer s.userLock.Unlock()

	if s.state == nil {
		return ErrNotAuthenticated
	}

	switch {
	case cmd.GetSelect() != nil:
		// 6.3.1. SELECT Command
		return s.handleSelect(ctx, tag, cmd.GetSelect(), ch)

	case cmd.GetExamine() != nil:
		// 6.3.2. EXAMINE Command
		return s.handleExamine(ctx, tag, cmd.GetExamine(), ch)

	case cmd.GetCreate() != nil:
		// 6.3.3. CREATE Command
		return s.handleCreate(ctx, tag, cmd.GetCreate(), ch)

	case cmd.GetDel() != nil:
		// 6.3.4. DELETE Command
		return s.handleDelete(ctx, tag, cmd.GetDel(), ch)

	case cmd.GetRename() != nil:
		// 6.3.5. RENAME Command
		return s.handleRename(ctx, tag, cmd.GetRename(), ch)

	case cmd.GetSub() != nil:
		// 6.3.6. SUBSCRIBE Command
		return s.handleSub(ctx, tag, cmd.GetSub(), ch)

	case cmd.GetUnsub() != nil:
		// 6.3.7. UNSUBSCRIBE Command
		return s.handleUnsub(ctx, tag, cmd.GetUnsub(), ch)

	case cmd.GetList() != nil:
		// 6.3.8. LIST Command
		return s.handleList(ctx, tag, cmd.GetList(), ch)

	case cmd.GetLsub() != nil:
		// 6.3.9. Lsub Command
		return s.handleLsub(ctx, tag, cmd.GetLsub(), ch)

	case cmd.GetStatus() != nil:
		// 6.3.10. STATUS Command
		return s.handleStatus(ctx, tag, cmd.GetStatus(), ch)

	case cmd.GetAppend() != nil:
		// 6.3.11. APPEND Command
		return s.handleAppend(ctx, tag, cmd.GetAppend(), ch)

	default:
		return fmt.Errorf("bad command")
	}
}

func (s *Session) handleSelectedCommand(
	ctx context.Context,
	tag string,
	cmd *proto.Command,
	ch chan response.Response,
) error {
	s.userLock.Lock()
	defer s.userLock.Unlock()

	if s.state == nil {
		return ErrNotAuthenticated
	}

	return s.state.Selected(ctx, func(mailbox *state.Mailbox) error {
		okResponse, err := s.handleWithMailbox(ctx, tag, cmd, mailbox, ch)
		if err != nil {
			return err
		}

		if err := flush(ctx, mailbox, false, ch); err != nil {
			return err
		}

		ch <- okResponse

		return nil
	})
}

func (s *Session) handleWithMailbox(
	ctx context.Context,
	tag string,
	cmd *proto.Command,
	mailbox *state.Mailbox,
	ch chan response.Response,
) (response.Response, error) {
	switch {
	case cmd.GetCheck() != nil:
		// 6.4.1. CHECK Command
		return s.handleCheck(ctx, tag, cmd.GetCheck(), mailbox, ch)

	case cmd.GetClose() != nil:
		// 6.4.2. CLOSE Command
		return s.handleClose(ctx, tag, cmd.GetClose(), mailbox, ch)

	case cmd.GetExpunge() != nil:
		// 6.4.3. EXPUNGE Command
		return s.handleExpunge(ctx, tag, cmd.GetExpunge(), mailbox, ch)

	case cmd.GetUidExpunge() != nil:
		// RFC4315 UIDPLUS Extension
		return s.handleUIDExpunge(ctx, tag, cmd.GetUidExpunge(), mailbox, ch)

	case cmd.GetUnselect() != nil:
		// RFC3691 UNSELECT Extension
		return s.handleUnselect(ctx, tag, cmd.GetUnselect(), mailbox, ch)

	case cmd.GetSearch() != nil:
		// 6.4.4. SEARCH Command
		return s.handleSearch(ctx, tag, cmd.GetSearch(), mailbox, ch)

	case cmd.GetFetch() != nil:
		// 6.4.5. FETCH Command
		return s.handleFetch(ctx, tag, cmd.GetFetch(), mailbox, ch)

	case cmd.GetStore() != nil:
		// 6.4.6. STORE Command
		return s.handleStore(ctx, tag, cmd.GetStore(), mailbox, ch)

	case cmd.GetCopy() != nil:
		// 6.4.7. COPY Command
		return s.handleCopy(ctx, tag, cmd.GetCopy(), mailbox, ch)

	case cmd.GetUid() != nil:
		// 6.4.8. UID Command
		return s.handleUID(ctx, tag, cmd.GetUid(), mailbox, ch)

	case cmd.GetMove() != nil:
		// RFC6851 MOVE Command
		return s.handleMove(ctx, tag, cmd.GetMove(), mailbox, ch)

	default:
		return nil, fmt.Errorf("bad command")
	}
}
