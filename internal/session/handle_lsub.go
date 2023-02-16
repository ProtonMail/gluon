package session

import (
	"context"
	"github.com/ProtonMail/gluon/imap/command"

	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleLsub(ctx context.Context, tag string, cmd *command.LSub, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeLSub)
	defer profiling.Stop(ctx, profiling.CmdTypeLSub)

	nameUTF8, err := s.decodeMailboxName(cmd.LSubMailbox)
	if err != nil {
		return err
	}

	return s.state.List(ctx, cmd.Mailbox, nameUTF8, true, func(matches map[string]state.Match) error {
		for _, match := range matches {
			nameUtf7, err := utf7.Encoding.NewEncoder().String(match.Name)
			if err != nil {
				panic(err)
			}
			select {
			case ch <- response.Lsub().
				WithName(nameUtf7).
				WithDelimiter(match.Delimiter).
				WithAttributes(match.Atts):

			case <-ctx.Done():
				return ctx.Err()
			}
		}

		ch <- response.Ok(tag).WithMessage("LSUB")

		return nil
	})
}
