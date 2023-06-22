package v1

import (
	"context"
	"fmt"
	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3/utils"
	v0 "github.com/ProtonMail/gluon/internal/db_impl/sqlite3/v0"
	"github.com/bradenaw/juniper/xslices"
	"strings"
)

type Migration struct{}

func (m Migration) Run(ctx context.Context, tx utils.QueryWrapper, generator imap.UIDValidityGenerator) error {
	// Migrate Messages And Flags.
	if err := migrateMessagesAndFlags(ctx, tx); err != nil {
		return err
	}

	// Migrate Mailboxes.
	if err := migrateMailboxes(ctx, tx, generator); err != nil {
		return fmt.Errorf("failed to migrate mailboxes: %w", err)
	}

	// Migrate Mailbox Messages.

	if err := deleteOldTables(ctx, tx); err != nil {
		return fmt.Errorf("failed to remove old tables: %w", err)
	}

	return nil
}

func migrateMailboxes(ctx context.Context, tx utils.QueryWrapper, generator imap.UIDValidityGenerator) error {
	// Create mailboxes table
	{
		query := fmt.Sprintf("CREATE TABLE `%[1]v` (`%[2]v` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `%[3]v` text NOT NULL UNIQUE, "+
			"`%[4]v` text NOT NULL UNIQUE, `%[5]v` integer NOT NULL , "+
			"`%[6]v` bool NOT NULL DEFAULT true)",
			MailboxesTableName,
			MailboxesFieldID,
			MailboxesFieldRemoteID,
			MailboxesFieldName,
			MailboxesFieldUIDValidity,
			MailboxesFieldSubscribed,
		)

		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return err
		}
	}

	// Create mailboxes flags table.
	{
		query := fmt.Sprintf("CREATE TABLE `%[1]v` (`%[2]v` text NOT NULL, `%[3]v` uuid NOT NULL, "+
			"CONSTRAINT `mailbox_flags_mailbox_id` FOREIGN KEY (`%[3]v`) REFERENCES `%[4]v` (`%[5]v`) ON DELETE CASCADE, "+
			"PRIMARY KEY (%[2]v, %[3]v)"+
			")",
			MailboxFlagsTableName,
			MailboxFlagsFieldValue,
			MailboxFlagsFieldMailboxID,
			MailboxesTableName,
			MailboxesFieldID,
		)

		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return err
		}
	}

	// Create perm mailboxes flags table.
	{
		query := fmt.Sprintf("CREATE TABLE `%[1]v` (`%[2]v` text NOT NULL, `%[3]v` uuid NOT NULL, "+
			"CONSTRAINT `perm_mailbox_flags_mailbox_id` FOREIGN KEY (`%[3]v`) REFERENCES `%[4]v` (`%[5]v`) ON DELETE CASCADE, "+
			"PRIMARY KEY (%[2]v, %[3]v)"+
			")",
			MailboxPermFlagsTableName,
			MailboxPermFlagsFieldValue,
			MailboxPermFlagsFieldMailboxID,
			MailboxesTableName,
			MailboxesFieldID,
		)

		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return err
		}
	}

	// Create Mailbox Attributes table.
	{
		query := fmt.Sprintf("CREATE TABLE `%[1]v` (`%[2]v` text NOT NULL, `%[3]v` uuid NOT NULL, "+
			"CONSTRAINT `perm_attrs_flags_mailbox_id` FOREIGN KEY (`%[3]v`) REFERENCES `%[4]v` (`%[5]v`) ON DELETE CASCADE, "+
			"PRIMARY KEY (%[2]v, %[3]v)"+
			")",
			MailboxAttrsTableName,
			MailboxAttrsFieldValue,
			MailboxAttrsFieldMailboxID,
			MailboxesTableName,
			MailboxesFieldID,
		)

		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return err
		}
	}

	// Create Message To Mailbox table.
	{
		query := fmt.Sprintf("CREATE TABLE `%[1]v` (`%[2]v` uuid NOT NULL, `%[3]v` integer NOT NULL, "+
			"CONSTRAINT `message_to_mailbox_message_id` FOREIGN KEY (`%[2]v`) REFERENCES `%[4]v` (`%[5]v`) ON DELETE CASCADE, "+
			"CONSTRAINT `message_to_mailbox_message_id` FOREIGN KEY (`%[3]v`) REFERENCES `%[6]v` (`%[7]v`) ON DELETE CASCADE, "+
			"PRIMARY KEY (%[2]v, %[3]v)"+
			")",
			MessageToMailboxTableName,
			MessageToMailboxFieldMessageID,
			MessageToMailboxFieldMailboxID,
			MessagesTableName,
			MessagesFieldID,
			MailboxesTableName,
			MailboxesFieldID,
		)

		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return err
		}
	}

	// Migrate mailboxes and assign new UID validity.
	loadExistingQuery := fmt.Sprintf("SELECT * FROM %v", v0.MailboxesTableName)

	mailboxes, err := utils.MapQueryRowsFn(ctx, tx, loadExistingQuery, scanMailboxV0)
	if err != nil {
		return err
	}

	for _, chunk := range xslices.Chunk(mailboxes, db.ChunkLimit) {
		query := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`, `%v`, `%v`) VALUES %v RETURNING `%v`",
			MailboxesTableName,
			MailboxesFieldRemoteID,
			MailboxesFieldName,
			MailboxesFieldUIDValidity,
			MailboxesFieldSubscribed,
			strings.Join(xslices.Repeat("(?,?,?,?)", len(chunk)), ","),
			MailboxesFieldID,
		)

		args := make([]any, 0, len(chunk)*5)

		for _, m := range chunk {
			newUIDValidity, err := generator.Generate()
			if err != nil {
				return err
			}

			args = append(args, m.RemoteID, m.Name, newUIDValidity, m.Subscribed)
		}

		newMailboxIDs, err := utils.MapQueryRows[imap.InternalMailboxID](ctx, tx, query, args...)
		if err != nil {
			return err
		}

		for _, mboxID := range newMailboxIDs {
			query := CreateMailboxMessageTableQuery(mboxID)

			if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
				return err
			}
		}
	}

	// Copy Messages data
	oldToNewIDMap := make(map[imap.InternalMailboxID]imap.InternalMailboxID, len(mailboxes))

	for _, m := range mailboxes {
		query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ? LIMIT 1",
			MailboxesFieldID,
			MailboxesTableName,
			MailboxesFieldName,
		)

		newID, err := utils.MapQueryRow[imap.InternalMailboxID](ctx, tx, query, m.Name)
		if err != nil {
			return err
		}

		oldToNewIDMap[m.ID] = newID
	}

	// Copy mailbox flags.
	if err := copyMailboxFlags(
		ctx,
		tx,
		oldToNewIDMap,
		v0.MailboxFlagsTableName,
		v0.MailboxFlagsFieldMailboxID,
		v0.MailboxFlagsFieldValue,
		MailboxFlagsTableName,
		MailboxFlagsFieldMailboxID,
		MailboxFlagsFieldValue,
	); err != nil {
		return err
	}

	// Copy mailbox perm flags.
	if err := copyMailboxFlags(
		ctx,
		tx,
		oldToNewIDMap,
		v0.MailboxPermFlagsTableName,
		v0.MailboxPermFlagsFieldMailboxID,
		v0.MailboxPermFlagsFieldValue,
		MailboxPermFlagsTableName,
		MailboxPermFlagsFieldMailboxID,
		MailboxPermFlagsFieldValue,
	); err != nil {
		return err
	}

	// Copy mailbox attributes.
	if err := copyMailboxFlags(
		ctx,
		tx,
		oldToNewIDMap,
		v0.MailboxAttrsTableName,
		v0.MailboxAttrsFieldMailboxID,
		v0.MailboxAttrsFieldValue,
		MailboxAttrsTableName,
		MailboxAttrsFieldMailboxID,
		MailboxAttrsFieldValue,
	); err != nil {
		return err
	}

	return migrateMailboxMessages(ctx, tx, oldToNewIDMap)
}

func migrateMessagesAndFlags(ctx context.Context, tx utils.QueryWrapper) error {
	if err := migrateMessages(ctx, tx); err != nil {
		return fmt.Errorf("failed to migrate messages: %w", err)
	}

	if err := migrateMessageFlags(ctx, tx); err != nil {
		return fmt.Errorf("failed to migrate message flags: %w", err)
	}

	return nil
}

func migrateMessages(ctx context.Context, tx utils.QueryWrapper) error {
	// Create new messages table.
	{
		query := fmt.Sprintf("CREATE TABLE `%[9]v` (`%[1]v` uuid NOT NULL, `%[2]v` text NOT NULL UNIQUE, "+
			"`%[3]v` datetime NOT NULL, `%[4]v` integer NOT NULL, `%[5]v` text NOT NULL, `%[6]v` text NOT NULL, "+
			"`%[7]v` text NOT NULL, `%[8]v` bool NOT NULL DEFAULT false, PRIMARY KEY (`%[1]v`))",
			MessagesFieldID,
			MessagesFieldRemoteID,
			MessagesFieldDate,
			MessagesFieldSize,
			MessagesFieldBody,
			MessagesFieldBodyStructure,
			MessagesFieldEnvelope,
			MessagesFieldDeleted,
			MessagesTableName,
		)

		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return err
		}
	}

	// Copy messages.
	{
		query := fmt.Sprintf("INSERT INTO %v SELECT * FROM %v", MessagesTableName, v0.MessagesTableName)
		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return err
		}
	}

	return nil
}

func migrateMessageFlags(ctx context.Context, tx utils.QueryWrapper) error {
	// Create new table.
	{
		query := fmt.Sprintf("CREATE TABLE `%[1]v` (`%[2]v` text NOT NULL, `%[3]v` uuid NOT NULL, "+
			"CONSTRAINT `message_flags_message_id` FOREIGN KEY (`%[3]v`) REFERENCES `%[4]v` (`%[5]v`) ON DELETE CASCADE, "+
			"PRIMARY KEY (%[2]v, %[3]v)"+
			")",
			MessageFlagsTableName,
			MessageFlagsFieldValue,
			MessageFlagsFieldMessageID,
			MessagesTableName,
			MessagesFieldID,
		)

		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return err
		}
	}

	// Migrate existing values to new table.
	{
		type V0MessageFlags struct {
			ID    imap.InternalMessageID
			Value string
		}
		selectQuery := fmt.Sprintf("SELECT `%v`, `%v` FROM %v", v0.MessageFlagsFieldMessageID, v0.MessageFlagsFieldValue, v0.MessageFlagsTableName)

		flags, err := utils.MapQueryRowsFn(ctx, tx, selectQuery, func(scanner utils.RowScanner) (V0MessageFlags, error) {
			var mf V0MessageFlags
			if err := scanner.Scan(&mf.ID, &mf.Value); err != nil {
				return V0MessageFlags{}, err
			}

			return mf, nil
		})
		if err != nil {
			return err
		}

		if len(flags) != 0 {
			remoteMessageIDsQuery := fmt.Sprintf("SELECT `%v`, `%v` FROM %v", v0.MessagesFieldID, v0.MessagesFieldRemoteID, v0.MessagesTableName)
			remoteMessagesIDs := make(map[imap.InternalMessageID]imap.MessageID)

			if err := utils.QueryForEachRow(ctx, tx, remoteMessageIDsQuery, func(scanner utils.RowScanner) error {
				var id imap.InternalMessageID
				var remoteID imap.MessageID

				if err := scanner.Scan(&id, &remoteID); err != nil {
					return err
				}

				remoteMessagesIDs[id] = remoteID

				return nil
			}); err != nil {
				return err
			}

			for _, chunk := range xslices.Chunk(flags, db.ChunkLimit) {
				insertQuery := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES %v",
					MessageFlagsTableName,
					MessageFlagsFieldMessageID,
					MessageFlagsFieldValue,
					strings.Join(xslices.Repeat("(?,?)", len(chunk)), ","),
				)

				args := make([]any, 0, len(chunk)*2)
				for _, flag := range flags {
					args = append(args, flag.ID, flag.Value)
				}

				if _, err := utils.ExecQuery(ctx, tx, insertQuery, args...); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func deleteOldTables(ctx context.Context, tx utils.QueryWrapper) error {
	// Drop Messages Flags table.
	{
		query := fmt.Sprintf("DROP TABLE `%v`", v0.MessageFlagsTableName)

		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return err
		}
	}

	// Drop mailbox flags table.
	{
		query := fmt.Sprintf("DROP TABLE `%v`", v0.MailboxFlagsTableName)
		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return err
		}
	}

	// Drop mailbox perm flags table.
	{
		query := fmt.Sprintf("DROP TABLE `%v`", v0.MailboxPermFlagsTableName)
		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return err
		}
	}

	// Drop mailbox attr table.
	{
		query := fmt.Sprintf("DROP TABLE `%v`", v0.MailboxAttrsTableName)
		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return err
		}
	}

	// Drop ui_ds table
	{
		query := fmt.Sprintf("DROP TABLE `%v`", v0.UIDsTableName)
		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return err
		}
	}

	// Drop Mailboxes table.
	{
		query := fmt.Sprintf("DROP TABLE `%v`", v0.MailboxesTableName)
		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return err
		}
	}

	// Drop messages table.
	{
		query := fmt.Sprintf("DROP TABLE `%v`", v0.MessagesTableName)
		if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
			return err
		}
	}

	return nil
}
