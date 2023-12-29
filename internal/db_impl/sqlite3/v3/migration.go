package v3

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3/utils"
	v1 "github.com/ProtonMail/gluon/internal/db_impl/sqlite3/v1"
)

type Migration struct{}

const MailboxFlagsTableNameTmp = "tmp_mailbox_flags"

const MailboxPermFlagsTableNameTmp = "tmp_mailbox_perm_flags"

const MailboxAttrsTableNameTmp = "tmp_mailbox_attr"

func (m Migration) Run(ctx context.Context, tx utils.QueryWrapper, _ imap.UIDValidityGenerator) error {
	// Create mailboxes flags table.
	{
		query := fmt.Sprintf("CREATE TABLE `%[1]v` (`%[3]v` integer NOT NULL, `%[2]v` text NOT NULL, "+
			"CONSTRAINT `mailbox_flags_mailbox_id` FOREIGN KEY (`%[3]v`) REFERENCES `%[4]v` (`%[5]v`) ON DELETE CASCADE, "+
			"PRIMARY KEY (%[2]v, %[3]v)"+
			")",
			MailboxFlagsTableNameTmp,
			v1.MailboxFlagsFieldValue,
			v1.MailboxFlagsFieldMailboxID,
			v1.MailboxesTableName,
			v1.MailboxesFieldID,
		)

		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return fmt.Errorf("failed to create mailboxes flags table: %w", err)
		}

		if err := migrateMailboxFlags(
			ctx,
			tx,
			v1.MailboxFlagsTableName,
			MailboxFlagsTableNameTmp,
			v1.MailboxFlagsFieldMailboxID,
			v1.MailboxFlagsFieldValue,
		); err != nil {
			return fmt.Errorf("failed to migrate mailbox flags table: %w", err)
		}
	}

	// Create perm mailboxes flags table.
	{
		query := fmt.Sprintf("CREATE TABLE `%[1]v` (`%[3]v` integer NOT NULL, `%[2]v` text NOT NULL, "+
			"CONSTRAINT `perm_mailbox_flags_mailbox_id` FOREIGN KEY (`%[3]v`) REFERENCES `%[4]v` (`%[5]v`) ON DELETE CASCADE, "+
			"PRIMARY KEY (%[2]v, %[3]v)"+
			")",
			MailboxPermFlagsTableNameTmp,
			v1.MailboxPermFlagsFieldValue,
			v1.MailboxPermFlagsFieldMailboxID,
			v1.MailboxesTableName,
			v1.MailboxesFieldID,
		)

		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return fmt.Errorf("failed to create mailboxes perm flags table: %w", err)
		}

		if err := migrateMailboxFlags(
			ctx,
			tx,
			v1.MailboxPermFlagsTableName,
			MailboxPermFlagsTableNameTmp,
			v1.MailboxPermFlagsFieldMailboxID,
			v1.MailboxPermFlagsFieldValue,
		); err != nil {
			return fmt.Errorf("failed to migrate mailbox perm flags table: %w", err)
		}
	}

	// Create Mailbox Attributes table.
	{
		query := fmt.Sprintf("CREATE TABLE `%[1]v` (`%[3]v` integer NOT NULL, `%[2]v` text NOT NULL, "+
			"CONSTRAINT `perm_attrs_flags_mailbox_id` FOREIGN KEY (`%[3]v`) REFERENCES `%[4]v` (`%[5]v`) ON DELETE CASCADE, "+
			"PRIMARY KEY (%[2]v, %[3]v)"+
			")",
			MailboxAttrsTableNameTmp,
			v1.MailboxAttrsFieldValue,
			v1.MailboxAttrsFieldMailboxID,
			v1.MailboxesTableName,
			v1.MailboxesFieldID,
		)

		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return fmt.Errorf("failed to create mailboxes attr table: %w", err)
		}

		if err := migrateMailboxFlags(
			ctx,
			tx,
			v1.MailboxAttrsTableName,
			MailboxAttrsTableNameTmp,
			v1.MailboxAttrsFieldMailboxID,
			v1.MailboxAttrsFieldValue,
		); err != nil {
			return fmt.Errorf("failed to migrate mailbox perm flags table: %w", err)
		}
	}

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

func migrateMailboxFlags(ctx context.Context,
	tx utils.QueryWrapper,
	fromTableName string,
	toTableName string,
	fieldID string,
	fieldValue string,
) error {
	query := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) SELECT `%v`,`%v` FROM %v ORDER BY `%v`",
		toTableName,
		fieldID,
		fieldValue,
		fieldID,
		fieldValue,
		fromTableName,
		fieldID,
	)

	if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
		return err
	}

	dropQuery := fmt.Sprintf("DROP TABLE %v", fromTableName)

	if _, err := utils.ExecQuery(ctx, tx, dropQuery); err != nil {
		return err
	}

	renameQuery := fmt.Sprintf("ALTER TABLE %v RENAME TO %v", toTableName, fromTableName)

	if _, err := utils.ExecQuery(ctx, tx, renameQuery); err != nil {
		return err
	}

	return nil
}
