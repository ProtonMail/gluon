package session

import (
	"context"
	"errors"
	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/ProtonMail/gluon/rfc5322"
)

func (s *Session) handleAppend(ctx context.Context, tag string, cmd *command.Append, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeAppend)
	defer profiling.Stop(ctx, profiling.CmdTypeAppend)

	nameUTF8, err := s.decodeMailboxName(cmd.Mailbox)
	if err != nil {
		return err
	}

	flags, err := validateStoreFlags(cmd.Flags)
	if err != nil {
		return response.Bad(tag).WithError(err)
	}

	if err := s.state.AppendOnlyMailbox(ctx, nameUTF8, func(mailbox state.AppendOnlyMailbox, isSameMBox bool) error {
		isDrafts, err := mailbox.IsDrafts(ctx)
		if err != nil {
			return err
		}

		if !isDrafts {
			if err := rfc5322.ValidateMessageHeaderFields(cmd.Literal); err != nil {
				return response.Bad(tag).WithError(err)
			}
		}

		messageUID, err := mailbox.Append(ctx, cmd.Literal, flags, cmd.DateTime)
		if err != nil {
			if shouldReportIMAPCommandError(err) {
				reporter.MessageWithContext(ctx,
					"Failed to append message to mailbox from state",
					reporter.Context{"error": err, "mailbox": nameUTF8},
				)
			}

			return err
		}

		if isSameMBox {
			if err := flush(ctx, mailbox, true, ch); err != nil {
				return err
			}
		}

		ch <- response.Ok(tag).WithItems(response.ItemAppendUID(mailbox.UIDValidity(), messageUID)).WithMessage("APPEND")

		return nil
	}); errors.Is(err, state.ErrNoSuchMailbox) {
		return response.No(tag).WithError(err).WithItems(response.ItemTryCreate())
	} else if err != nil {
		return err
	}

	return nil
}
