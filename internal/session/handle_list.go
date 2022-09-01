package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleList(ctx context.Context, tag string, cmd *proto.List, ch chan response.Response) error {
	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	return s.state.List(ctx, cmd.GetReference(), nameUTF8, false, func(matches map[string]state.Match) error {
		for _, match := range matches {
			select {
			case ch <- response.List().
				WithName(match.Name).
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
