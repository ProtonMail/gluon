package backend

import (
	"context"
	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/internal/state"
)

type DBIMAPState struct {
	user   *user
	client db.Client
}

func NewDBIMAPState(client db.Client) *DBIMAPState {
	return &DBIMAPState{client: client}
}

func (d *DBIMAPState) Read(ctx context.Context, f func(context.Context, connector.IMAPStateRead) error) error {
	return d.client.Read(ctx, func(ctx context.Context, only db.ReadOnly) error {
		rd := DBIMAPStateRead{rd: only}

		return f(ctx, &rd)
	})
}

func (d *DBIMAPState) Write(ctx context.Context, f func(context.Context, connector.IMAPStateWrite) error) error {
	var stateUpdates []state.Update

	err := d.client.Write(ctx, func(ctx context.Context, tx db.Transaction) error {
		wr := DBIMAPStateWrite{DBIMAPStateRead: DBIMAPStateRead{rd: tx}, tx: tx, user: d.user}

		err := f(ctx, &wr)

		stateUpdates = wr.stateUpdates

		return err
	})

	if err == nil {
		d.user.queueStateUpdate(stateUpdates...)
	}

	return err
}
