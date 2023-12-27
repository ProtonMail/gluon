package v3

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3/utils"
	v1 "github.com/ProtonMail/gluon/internal/db_impl/sqlite3/v1"
)

type Migration struct{}

func (m Migration) Run(ctx context.Context, tx utils.QueryWrapper, _ imap.UIDValidityGenerator) error {
	{
		// Create an index on message id field to speed up lookup queries for message flags.
		query := fmt.Sprintf("create index message_flags_message_id_index on %v (%v)",
			v1.MessageFlagsTableName,
			v1.MessageFlagsFieldMessageID,
		)

		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return fmt.Errorf("failed to create message flags index: %w", err)
		}
	}

	{
		// Create an index on message id field to speed up lookup queries for message flags.
		query := fmt.Sprintf("create index mailbox_flags_mailbox_id_index on %v (%v)",
			v1.MailboxFlagsTableName,
			v1.MailboxFlagsFieldMailboxID,
		)

		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return fmt.Errorf("failed to create mailbox flags index : %w", err)
		}
	}

	{
		// Create an index on message id field to speed up lookup queries for message flags.
		query := fmt.Sprintf("create index mailbox_perm_flags_mailbox_id_index on %v (%v)",
			v1.MailboxPermFlagsTableName,
			v1.MailboxPermFlagsFieldMailboxID,
		)

		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return fmt.Errorf("failed to create mailbox perm flags index : %w", err)
		}
	}

	{
		// Create an index on message id field to speed up lookup queries for message flags.
		query := fmt.Sprintf("create index mailbox_attr_mailbox_id_index on %v (%v)",
			v1.MailboxAttrsTableName,
			v1.MailboxAttrsFieldMailboxID,
		)

		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return fmt.Errorf("failed to create mailbox attr flags index : %w", err)
		}
	}

	return nil
}
