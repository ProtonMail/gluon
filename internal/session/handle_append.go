package session

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleAppend(ctx context.Context, tag string, cmd *proto.Append, ch chan response.Response) error {
	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	if strings.EqualFold(nameUTF8, imap.Inbox) {
		nameUTF8 = imap.Inbox
	}

	flags, err := validateStoreFlags(cmd.GetFlags())
	if err != nil {
		return response.Bad(tag).WithError(err)
	}

	if err := s.state.Mailbox(ctx, nameUTF8, func(mailbox *backend.Mailbox) error {
		messageUID, err := mailbox.Append(ctx, cmd.GetMessage(), flags, toTime(cmd.GetDateTime()))
		if err != nil {
			return err
		}

		if err := flush(ctx, mailbox, true, ch); err != nil {
			return err
		}

		ch <- response.Ok(tag).WithItems(response.ItemAppendUID(mailbox.UIDValidity(), messageUID))

		return nil
	}); errors.Is(err, backend.ErrNoSuchMailbox) {
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
