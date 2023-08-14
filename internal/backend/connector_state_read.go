package backend

import (
	"context"
	"github.com/ProtonMail/gluon/db"
)

type DBIMAPStateRead struct {
	rd db.ReadOnly
}

func (d *DBIMAPStateRead) GetSettings(ctx context.Context) (string, bool, error) {
	return d.rd.GetConnectorSettings(ctx)
}

func (d *DBIMAPStateRead) GetMailboxCount(ctx context.Context) (int, error) {
	return d.rd.GetMailboxCount(ctx)
}
