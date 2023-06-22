package session

import (
	"context"

	"github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleSelect(ctx context.Context, tag string, cmd *command.Select, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeSelect)
	defer profiling.Stop(ctx, profiling.CmdTypeSelect)

	nameUTF8, err := s.decodeMailboxName(cmd.Mailbox)
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

		uidNext, err := mailbox.UIDNext(ctx)
		if err != nil {
			return err
		}

		ch <- response.Flags().WithFlags(flags)
		ch <- response.Exists().WithCount(imap.SeqID(mailbox.Count()))
		ch <- response.Recent().WithCount(uint32(mailbox.GetMessagesWithFlagCount(imap.FlagRecent)))
		ch <- response.Ok().WithItems(response.ItemPermanentFlags(permFlags)).WithMessage("Flags permitted")
		ch <- response.Ok().WithItems(response.ItemUIDNext(uidNext)).WithMessage("Predicted next UID")
		ch <- response.Ok().WithItems(response.ItemUIDValidity(mailbox.UIDValidity())).WithMessage("UIDs valid")

		if unseen, ok := mailbox.GetFirstMessageWithoutFlag(imap.FlagSeen); ok {
			ch <- response.Ok().WithItems(response.ItemUnseen(uint32(unseen.Seq))).WithMessage("Unseen messages")
		}

		return nil
	}); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithItems(response.ItemReadWrite()).WithMessage("SELECT")

	s.eventCh <- events.Select{
		SessionID: s.sessionID,
		Mailbox:   nameUTF8,
	}

	return nil
}
