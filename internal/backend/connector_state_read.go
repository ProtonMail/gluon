package backend

import (
	"context"
	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/bradenaw/juniper/xslices"
	"strings"
)

type DBIMAPStateRead struct {
	rd        db.ReadOnly
	delimiter string
}

func (d *DBIMAPStateRead) GetSettings(ctx context.Context) (string, bool, error) {
	return d.rd.GetConnectorSettings(ctx)
}

func (d *DBIMAPStateRead) GetMailboxCount(ctx context.Context) (int, error) {
	return d.rd.GetMailboxCount(ctx)
}

func (d *DBIMAPStateRead) GetMailboxesWithoutAttrib(ctx context.Context) ([]imap.MailboxNoAttrib, error) {
	mboxes, err := d.rd.GetAllMailboxesNameAndRemoteID(ctx)
	if err != nil {
		return nil, err
	}

	return xslices.Map(mboxes, func(m db.MailboxNameAndRemoteID) imap.MailboxNoAttrib {
		return imap.MailboxNoAttrib{
			ID:   m.RemoteID,
			Name: strings.Split(m.Name, d.delimiter),
		}
	}), nil
}
