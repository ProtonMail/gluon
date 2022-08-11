package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleLsub(ctx context.Context, tag string, cmd *proto.Lsub, ch chan response.Response) error {
	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	return s.state.List(ctx, cmd.GetReference(), nameUTF8, true, func(matches map[string]backend.Match) error {
		for _, match := range matches {
			ch <- response.Lsub().
				WithName(match.Name).
				WithDelimiter(match.Delimiter).
				WithAttributes(match.Atts)
		}

		ch <- response.Ok(tag).WithMessage("LSUB")

		return nil
	})
}
