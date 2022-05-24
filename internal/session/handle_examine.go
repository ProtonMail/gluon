package session

import (
	"context"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleExamine(ctx context.Context, tag string, cmd *proto.Examine, ch chan response.Response) error {
	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	if strings.EqualFold(nameUTF8, imap.Inbox) {
		nameUTF8 = imap.Inbox
	}

	if err := s.state.Examine(ctx, nameUTF8, func(mailbox *backend.Mailbox) error {
		flags, err := mailbox.Flags(ctx)
		if err != nil {
			return err
		}

		permFlags, err := mailbox.PermanentFlags(ctx)
		if err != nil {
			return err
		}

		ch <- response.Flags().WithFlags(flags)
		ch <- response.Exists().WithCount(mailbox.Count())
		ch <- response.Recent().WithCount(len(mailbox.GetMessagesWithFlag(imap.FlagRecent)))
		ch <- response.Ok().WithItems(response.ItemPermanentFlags(permFlags))
		ch <- response.Ok().WithItems(response.ItemUIDNext(mailbox.UIDNext()))
		ch <- response.Ok().WithItems(response.ItemUIDValidity(mailbox.UIDValidity()))

		if unseen := mailbox.GetMessagesWithoutFlag(imap.FlagSeen); len(unseen) > 0 {
			ch <- response.Ok().WithItems(response.ItemUnseen(unseen[0]))
		}

		s.name = mailbox.Name()

		return nil
	}); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithItems(response.ItemReadOnly())

	return nil
}
