package session

import (
	"context"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleStatus(ctx context.Context, tag string, cmd *proto.Status, ch chan response.Response) error {
	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	if err := s.state.Mailbox(ctx, nameUTF8, func(mailbox *state.Mailbox) error {
		if mailbox.Selected() {
			if err := flush(ctx, mailbox, true, ch); err != nil {
				return err
			}
		}

		var items []response.Item

		for _, att := range cmd.GetAttributes() {
			switch {
			case strings.EqualFold(att, imap.StatusMessages):
				items = append(items, response.ItemMessages(mailbox.Count()))

			case strings.EqualFold(att, imap.StatusRecent):
				items = append(items, response.ItemRecent(len(mailbox.GetMessagesWithFlag(imap.FlagRecent))))

			case strings.EqualFold(att, imap.StatusUIDNext):
				items = append(items, response.ItemUIDNext(mailbox.UIDNext()))

			case strings.EqualFold(att, imap.StatusUIDValidity):
				items = append(items, response.ItemUIDValidity(mailbox.UIDValidity()))

			case strings.EqualFold(att, imap.StatusUnseen):
				items = append(items, response.ItemUnseen(uint32(len(mailbox.GetMessagesWithoutFlag(imap.FlagSeen)))))
			}
		}

		ch <- response.Status().WithMailbox(nameUTF8).WithItems(items...)

		return nil
	}); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithMessage("STATUS")

	return nil
}
