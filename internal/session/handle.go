package session

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/logging"
)

func (s *Session) handlePanic() {
	if s.panicHandler != nil {
		s.panicHandler.HandlePanic()
	}
}

func (s *Session) handleOther(
	ctx context.Context,
	tag string,
	cmd command.Payload,
) <-chan response.Response {
	resCh := make(chan response.Response, 8)

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

	return resCh
}

// handleCommand returns a response instance if a command needs to force an exit of the client.
func (s *Session) handleCommand(
	ctx context.Context,
	tag string,
	cmd command.Payload,
	ch chan response.Response,
) error {
	switch cmd.(type) {
	case
		*command.Capability,
		*command.IDGet,
		*command.IDSet,
		*command.Noop:
		return s.handleAnyCommand(ctx, tag, cmd, ch)

	case
		*command.Login:
		return s.handleNotAuthenticatedCommand(ctx, tag, cmd, ch)

	case
		*command.Select,
		*command.Examine,
		*command.Create,
		*command.Delete,
		*command.Rename,
		*command.Subscribe,
		*command.Unsubscribe,
		*command.List,
		*command.LSub,
		*command.Status,
		*command.Append:
		return s.handleAuthenticatedCommand(ctx, tag, cmd, ch)
	case
		*command.Check,
		*command.Close,
		*command.Expunge,
		*command.UIDExpunge,
		*command.Unselect,
		*command.Search,
		*command.Fetch,
		*command.Store,
		*command.Copy,
		*command.Move,
		*command.UID:
		return s.handleSelectedCommand(ctx, tag, cmd, ch)

	default:
		return fmt.Errorf("bad command")
	}
}

func (s *Session) handleAnyCommand(
	ctx context.Context,
	tag string,
	cmd command.Payload,
	ch chan response.Response,
) error {
	switch cmd := cmd.(type) {
	case *command.Capability:
		// 6.1.1 CAPABILITY Command
		return s.handleCapability(ctx, tag, cmd, ch)

	case *command.Noop:
		// 6.1.2 NOOP Command
		return s.handleNoop(ctx, tag, cmd, ch)

	case *command.IDSet:
		// RFC 2971 ID
		return s.handleIDSet(ctx, tag, cmd, ch)

	case *command.IDGet:
		// RFC 2971 ID
		return s.handleIDGet(ctx, tag, ch)

	default:
		return fmt.Errorf("bad command")
	}
}

func (s *Session) handleNotAuthenticatedCommand(
	ctx context.Context,
	tag string,
	cmd command.Payload,
	ch chan response.Response,
) error {
	switch cmd := cmd.(type) {
	case *command.Login:
		// 6.2.3. LOGIN Command
		return s.handleLogin(ctx, tag, cmd, ch)

	default:
		return fmt.Errorf("bad command")
	}
}

func (s *Session) handleAuthenticatedCommand(
	ctx context.Context,
	tag string,
	cmd command.Payload,
	ch chan response.Response,
) error {
	s.userLock.Lock()
	defer s.userLock.Unlock()

	if s.state == nil {
		return ErrNotAuthenticated
	}

	switch cmd := cmd.(type) {
	case *command.Select:
		// 6.3.1. SELECT Command
		return s.handleSelect(ctx, tag, cmd, ch)

	case *command.Examine:
		// 6.3.2. EXAMINE Command
		return s.handleExamine(ctx, tag, cmd, ch)

	case *command.Create:
		// 6.3.3. CREATE Command
		return s.handleCreate(ctx, tag, cmd, ch)

	case *command.Delete:
		// 6.3.4. DELETE Command
		return s.handleDelete(ctx, tag, cmd, ch)

	case *command.Rename:
		// 6.3.5. RENAME Command
		return s.handleRename(ctx, tag, cmd, ch)

	case *command.Subscribe:
		// 6.3.6. SUBSCRIBE Command
		return s.handleSub(ctx, tag, cmd, ch)

	case *command.Unsubscribe:
		// 6.3.7. UNSUBSCRIBE Command
		return s.handleUnsub(ctx, tag, cmd, ch)

	case *command.List:
		// 6.3.8. LIST Command
		return s.handleList(ctx, tag, cmd, ch)

	case *command.LSub:
		// 6.3.9. Lsub Command
		return s.handleLsub(ctx, tag, cmd, ch)

	case *command.Status:
		// 6.3.10. STATUS Command
		return s.handleStatus(ctx, tag, cmd, ch)

	case *command.Append:
		// 6.3.11. APPEND Command
		return s.handleAppend(ctx, tag, cmd, ch)

	default:
		return fmt.Errorf("bad command")
	}
}

func (s *Session) handleSelectedCommand(
	ctx context.Context,
	tag string,
	cmd command.Payload,
	ch chan response.Response,
) error {
	s.userLock.Lock()
	defer s.userLock.Unlock()

	if s.state == nil {
		return ErrNotAuthenticated
	}

	return s.state.Selected(ctx, func(mailbox *state.Mailbox) error {
		okResponse, err := s.handleWithMailbox(ctx, tag, cmd, mailbox, ch)

		// Allow state updates to be applied if the command failed. It might resolve
		// some invalid state problem.
		if flushErr := flush(ctx, mailbox, false, ch); flushErr != nil {
			if err != nil {
				return fmt.Errorf("%w (flush err:%v)", err, flushErr)
			}

			return flushErr
		}

		if err != nil {
			return err
		}

		ch <- okResponse

		return nil
	})
}

func (s *Session) handleWithMailbox(
	ctx context.Context,
	tag string,
	cmd command.Payload,
	mailbox *state.Mailbox,
	ch chan response.Response,
) (response.Response, error) {
	switch cmd := cmd.(type) {
	case *command.Check:
		// 6.4.1. CHECK Command
		return s.handleCheck(ctx, tag, cmd, mailbox, ch)

	case *command.Close:
		// 6.4.2. CLOSE Command
		return s.handleClose(ctx, tag, cmd, mailbox, ch)

	case *command.Expunge:
		// 6.4.3. EXPUNGE Command
		return s.handleExpunge(ctx, tag, cmd, mailbox, ch)

	case *command.UIDExpunge:
		// RFC4315 UIDPLUS Extension
		return s.handleUIDExpunge(ctx, tag, cmd, mailbox, ch)

	case *command.Unselect:
		// RFC3691 UNSELECT Extension
		return s.handleUnselect(ctx, tag, cmd, mailbox, ch)

	case *command.Search:
		// 6.4.4. SEARCH Command
		return s.handleSearch(ctx, tag, cmd, mailbox, ch)

	case *command.Fetch:
		// 6.4.5. FETCH Command
		return s.handleFetch(ctx, tag, cmd, mailbox, ch)

	case *command.Store:
		// 6.4.6. STORE Command
		return s.handleStore(ctx, tag, cmd, mailbox, ch)

	case *command.Copy:
		// 6.4.7. COPY Command
		return s.handleCopy(ctx, tag, cmd, mailbox, ch)

	case *command.UID:
		// 6.4.8. UID Command
		return s.handleUID(ctx, tag, cmd, mailbox, ch)

	case *command.Move:
		// RFC6851 MOVE Command
		return s.handleMove(ctx, tag, cmd, mailbox, ch)

	default:
		return nil, fmt.Errorf("bad command")
	}
}
