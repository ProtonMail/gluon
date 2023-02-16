package session

import (
	"context"
	"fmt"
	"github.com/ProtonMail/gluon/imap/command"

	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleList(ctx context.Context, tag string, cmd *command.List, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeList)
	defer profiling.Stop(ctx, profiling.CmdTypeList)

	nameUTF8, err := s.decodeMailboxName(cmd.ListMailbox)
	if err != nil {
		return err
	}

	return s.state.List(ctx, cmd.Mailbox, nameUTF8, false, func(matches map[string]state.Match) error {
		for _, match := range matches {
			nameUtf7, err := utf7.Encoding.NewEncoder().String(match.Name)
			if err != nil {
				return fmt.Errorf("failed to convert name to utf7")
			}
			select {
			case ch <- response.List().
				WithName(nameUtf7).
				WithDelimiter(match.Delimiter).
				WithAttributes(match.Atts):

			case <-ctx.Done():
				return ctx.Err()
			}
		}

		ch <- response.Ok(tag).WithMessage("LIST")

		return nil
	})
}
