package session

import (
	"context"

	"github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleSelect(ctx context.Context, tag string, cmd *proto.Select, ch chan response.Response) error {
	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	if err := s.state.Select(ctx, nameUTF8, func(mailbox *state.Mailbox) error {
		flags, err := mailbox.Flags(ctx)
		if err != nil {
			return err
		}

		permFlags, err := mailbox.PermanentFlags(ctx)
		if err != nil {
			return err
		}

		ch <- response.Flags().WithFlags(flags)
		ch <- response.Exists().WithCount(imap.SeqID(mailbox.Count()))
		ch <- response.Recent().WithCount(uint32(len(mailbox.GetMessagesWithFlag(imap.FlagRecent))))
		ch <- response.Ok().WithItems(response.ItemPermanentFlags(permFlags)).WithMessage("Flags permitted")
		ch <- response.Ok().WithItems(response.ItemUIDNext(mailbox.UIDNext())).WithMessage("Predicted next UID")
		ch <- response.Ok().WithItems(response.ItemUIDValidity(mailbox.UIDValidity())).WithMessage("UIDs valid")

		if unseen := mailbox.GetMessagesWithoutFlag(imap.FlagSeen); len(unseen) > 0 {
			ch <- response.Ok().WithItems(response.ItemUnseen(uint32(unseen[0]))).WithMessage("Unseen messages")
		}

		return nil
	}); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithItems(response.ItemReadWrite()).WithMessage("SELECT")

	s.eventCh <- events.EventSelect{
		SessionID: s.sessionID,
		Mailbox:   nameUTF8,
	}

	return nil
}
