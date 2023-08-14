package backend

import (
	"context"
	"fmt"
	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/ProtonMail/gluon/internal/state"
	"strings"
)

type DBIMAPStateWrite struct {
	DBIMAPStateRead
	tx           db.Transaction
	user         *user
	stateUpdates []state.Update
}

func (d *DBIMAPStateWrite) CreateMailbox(ctx context.Context, mailbox imap.Mailbox) error {
	if mailbox.ID == ids.GluonInternalRecoveryMailboxRemoteID {
		return fmt.Errorf("attempting to create protected mailbox (recovery)")
	}

	if exists, err := d.tx.MailboxExistsWithRemoteID(ctx, mailbox.ID); err != nil {
		return err
	} else if exists {
		return nil
	}

	uidValidity, err := d.user.uidValidityGenerator.Generate()
	if err != nil {
		return err
	}

	if err := d.user.imapLimits.CheckUIDValidity(uidValidity); err != nil {
		return err
	}

	if mailboxCount, err := d.tx.GetMailboxCount(ctx); err != nil {
		return err
	} else if err := d.user.imapLimits.CheckMailBoxCount(mailboxCount); err != nil {
		return err
	}

	if _, err := d.tx.CreateMailbox(
		ctx,
		mailbox.ID,
		strings.Join(mailbox.Name, d.user.delimiter),
		mailbox.Flags,
		mailbox.PermanentFlags,
		mailbox.Attributes,
		uidValidity,
	); err != nil {
		return err
	}

	return nil
}

func (d *DBIMAPStateWrite) UpdateMessageFlags(ctx context.Context, id imap.MessageID, flags imap.FlagSet) error {
	if exists, err := d.tx.MessageExistsWithRemoteID(ctx, id); err != nil {
		return err
	} else if !exists {
		return state.ErrNoSuchMessage
	}

	return d.wrapStateUpdates(ctx, func(ctx context.Context, tx db.Transaction) ([]state.Update, error) {
		internalMsgID, err := tx.GetMessageIDFromRemoteID(ctx, id)

		if err != nil {
			if db.IsErrNotFound(err) {
				return nil, state.ErrNoSuchMessage
			}
			return nil, err
		}

		return d.user.setMessageFlags(ctx, tx, internalMsgID, flags)
	})
}

func (d *DBIMAPStateWrite) StoreSettings(ctx context.Context, settings string) error {
	return d.tx.StoreConnectorSettings(ctx, settings)
}

func (d *DBIMAPStateWrite) wrapStateUpdates(ctx context.Context, f func(ctx context.Context, tx db.Transaction) ([]state.Update, error)) error {
	updates, err := f(ctx, d.tx)
	if err == nil {
		d.stateUpdates = append(d.stateUpdates, updates...)
	}

	return err
}
