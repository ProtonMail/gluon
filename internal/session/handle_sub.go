package session

import (
	"context"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleSub(ctx context.Context, tag string, cmd *proto.Sub, ch chan response.Response) error {
	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	if strings.EqualFold(nameUTF8, imap.Inbox) {
		nameUTF8 = imap.Inbox
	}

	if err := s.state.Subscribe(ctx, nameUTF8); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithMessage("SUB")

	return nil
}
