package session

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
)

func (s *Session) handleStatus(ctx context.Context, tag string, cmd *command.Status, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeStatus)
	defer profiling.Stop(ctx, profiling.CmdTypeStatus)

	nameUTF8, err := s.decodeMailboxName(cmd.Mailbox)
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

		for _, att := range cmd.Attributes {
			switch att {
			case command.StatusAttributeMessages:
				items = append(items, response.ItemMessages(mailbox.Count()))

			case command.StatusAttributeRecent:
				items = append(items, response.ItemRecent(mailbox.GetMessagesWithFlagCount(imap.FlagRecent)))

			case command.StatusAttributeUIDNext:
				items = append(items, response.ItemUIDNext(mailbox.UIDNext()))

			case command.StatusAttributeUIDValidity:
				items = append(items, response.ItemUIDValidity(mailbox.UIDValidity()))

			case command.StatusAttributeUnseen:
				items = append(items, response.ItemUnseen(uint32(mailbox.GetMessagesWithoutFlagCount(imap.FlagSeen))))
			}
		}

		ch <- response.Status().WithMailbox(cmd.Mailbox).WithItems(items...)

		return nil
	}); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithMessage("STATUS")

	return nil
}
