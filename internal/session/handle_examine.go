package session

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleExamine(ctx context.Context, tag string, cmd *command.Examine, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeExamine)
	defer profiling.Stop(ctx, profiling.CmdTypeExamine)

	nameUTF8, err := s.decodeMailboxName(cmd.Mailbox)
	if err != nil {
		return err
	}

	if err := s.state.Examine(ctx, nameUTF8, func(mailbox *state.Mailbox) error {
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
		ch <- response.Recent().WithCount(uint32(mailbox.GetMessagesWithFlagCount(imap.FlagRecent)))
		ch <- response.Ok().WithItems(response.ItemPermanentFlags(permFlags))
		ch <- response.Ok().WithItems(response.ItemUIDNext(mailbox.UIDNext()))
		ch <- response.Ok().WithItems(response.ItemUIDValidity(mailbox.UIDValidity()))

		if unseen, ok := mailbox.GetFirstMessageWithoutFlag(imap.FlagSeen); ok {
			ch <- response.Ok().WithItems(response.ItemUnseen(uint32(unseen.Seq)))
		}

		return nil
	}); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithItems(response.ItemReadOnly()).WithMessage("EXAMINE")

	return nil
}
