package session

import (
	"context"
	"errors"
	"time"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/reporter"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleAppend(ctx context.Context, tag string, cmd *proto.Append, ch chan response.Response) error {
	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	flags, err := validateStoreFlags(cmd.GetFlags())
	if err != nil {
		return response.Bad(tag).WithError(err)
	}

	if err := s.state.EmptyMailbox(ctx, nameUTF8, func(mailbox state.NoLoadMailbox, isSameMBox bool) error {
		messageUID, err := mailbox.Append(ctx, cmd.GetMessage(), flags, toTime(cmd.GetDateTime()))
		if err != nil {
			reporter.MessageWithContext(ctx,
				"Failed to append message to mailbox from state",
				reporter.Context{"error": err},
			)

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

func toTime(dt *proto.DateTime) time.Time {
	if dt == nil {
		return time.Now()
	}

	zone := dt.Zone.Hour*3600 + dt.Zone.Minute*60

	if !dt.Zone.Sign {
		zone *= -1
	}

	return time.Date(
		int(dt.Date.Year),
		time.Month(dt.Date.Month),
		int(dt.Date.Day),
		int(dt.Time.Hour),
		int(dt.Time.Minute),
		int(dt.Time.Second),
		0,
		time.FixedZone("zone", int(zone)),
	)
}
