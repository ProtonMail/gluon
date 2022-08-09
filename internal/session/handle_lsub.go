package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleLsub(ctx context.Context, tag string, cmd *proto.Lsub, ch chan response.Response) error {
	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	if err := s.state.List(ctx, cmd.GetReference(), nameUTF8, true, ch); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithMessage("LSUB")

	return nil
}
