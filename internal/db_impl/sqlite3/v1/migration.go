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

func (m Migration) Run(ctx context.Context, tx utils.QueryWrapper) error {
	// Create new flags table and migrate data.
	if err := migrateMessageFlags(ctx, tx); err != nil {
		return err
	}

	// Create new mailbox flags and attributes tables

	// Migrate mailbox and attributes tables

	// Create a new table for each mailbox, use id as mailbox_table name

	// Migrate all entries from UIDs table to new mailbox tables

	// Delete old flags, mailbox flags and uids table

	return nil
}

func migrateMessageFlags(ctx context.Context, tx utils.QueryWrapper) error {
	// Create new table.
	{
		query := fmt.Sprintf("CREATE TABLE `%[1]v` (`%[2]v` text NOT NULL, `%[3]v` uuid NOT NULL, "+
			"CONSTRAINT `message_flags_messages_flags` FOREIGN KEY (`%[3]v`) REFERENCES `%[4]v` (`%[5]v`) ON DELETE CASCADE, "+
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

			args := make([]any, 0, len(chunk)*3)
			for _, flag := range flags {
				args = append(args, flag.ID, remoteMessagesIDs[flag.ID], flag.Value)
			}

			if err := utils.ExecQueryAndCheckUpdatedNotZero(ctx, tx, insertQuery, args...); err != nil {
				return err
			}
		}

	}

	// Delete old messages table.
	{
		query := fmt.Sprintf("DROP TABLE %v", v0.MessageFlagsTableName)

		if err := utils.ExecQueryAndCheckUpdatedNotZero(ctx, tx, query); err != nil {
			return err
		}
	}

	return nil
}
