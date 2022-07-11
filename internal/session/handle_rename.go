package session

import (
	"context"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleRename(ctx context.Context, tag string, cmd *proto.Rename, ch chan response.Response) error {
	oldNameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	if strings.EqualFold(oldNameUTF8, imap.Inbox) {
		oldNameUTF8 = imap.Inbox
	}

	newNameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetNewName())
	if err != nil {
		return err
	}

	if err := s.state.Rename(ctx, oldNameUTF8, newNameUTF8); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithMessage("RENAME")

	return nil
}
