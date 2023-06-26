package v1

import (
	"context"
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3/utils"
	v0 "github.com/ProtonMail/gluon/internal/db_impl/sqlite3/v0"
	"github.com/bradenaw/juniper/xslices"
)

type mailboxV0 struct {
	ID          imap.InternalMailboxID
	RemoteID    imap.MailboxID
	Name        string
	UIDNext     imap.UID
	UIDValidity imap.UID
	Subscribed  bool
}

func scanMailboxV0(scanner utils.RowScanner) (mailboxV0, error) {
	var mbox mailboxV0

	if err := scanner.Scan(&mbox.ID, &mbox.RemoteID, &mbox.Name, &mbox.UIDNext, &mbox.UIDValidity, &mbox.Subscribed); err != nil {
		return mailboxV0{}, err
	}

	return mbox, nil
}

func MailboxMessageTableName(id imap.InternalMailboxID) string {
	return fmt.Sprintf("mailbox_message_%v", id)
}

func CreateMailboxMessageTableQuery(id imap.InternalMailboxID) string {
	tableName := MailboxMessageTableName(id)

	return fmt.Sprintf("CREATE TABLE `%[1]v` ("+
		"`%[2]v` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `%[3]v` bool NOT NULL DEFAULT false, `%[4]v` bool NOT NULL DEFAULT true, "+
		"`%[5]v` uuid NOT NULL UNIQUE, "+
		"`%[6]v` string NOT NULL UNIQUE, "+
		"CONSTRAINT `%[1]v_message_id` FOREIGN KEY (`%[5]v`) REFERENCES `%[7]v` (`%[8]v`) ON DELETE SET NULL)",
		tableName,
		MailboxMessagesFieldUID,
		MailboxMessagesFieldDeleted,
		MailboxMessagesFieldRecent,
		MailboxMessagesFieldMessageID,
		MailboxMessagesFieldMessageRemoteID,
		MessagesTableName,
		MessagesFieldID,
	)
}

type mailboxFlag struct {
	ID    imap.InternalMailboxID
	Value string
}

func scanMailboxFlag(scanner utils.RowScanner) (mailboxFlag, error) {
	var mf mailboxFlag

	if err := scanner.Scan(&mf.ID, &mf.Value); err != nil {
		return mailboxFlag{}, err
	}

	return mf, nil
}

func copyMailboxFlags(ctx context.Context,
	tx utils.QueryWrapper,
	oldToNewIDMap map[imap.InternalMailboxID]imap.InternalMailboxID,
	fromTableName string,
	fromFieldID string,
	fromFieldValue string,
	toTableName string,
	toFieldID string,
	toFieldValue string,
) error {
	loadQuery := fmt.Sprintf("SELECT `%v`, `%v` FROM %v", fromFieldID, fromFieldValue, fromTableName)

	flags, err := utils.MapQueryRowsFn(ctx, tx, loadQuery, scanMailboxFlag)
	if err != nil {
		return err
	}

	for _, chunk := range xslices.Chunk(flags, db.ChunkLimit) {
		insertQuery := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES %v",
			toTableName,
			toFieldID,
			toFieldValue,
			strings.Join(xslices.Repeat("(?,?)", len(chunk)), ","),
		)

		args := make([]any, 0, len(chunk)*2)

		for _, f := range chunk {
			newId, ok := oldToNewIDMap[f.ID]
			if !ok {
				return fmt.Errorf("failed to map old mailbox id %v to new", f.ID)
			}

			args = append(args, newId, f.Value)
		}

		if _, err := utils.ExecQuery(ctx, tx, insertQuery, args...); err != nil {
			return err
		}
	}

	return nil
}

type uidMessageV0 struct {
	UID       imap.UID
	Deleted   bool
	Recent    bool
	MessageID imap.InternalMessageID
	RemoteID  imap.MessageID
}

func scanUIDMessageV0(scanner utils.RowScanner) (uidMessageV0, error) {
	var m uidMessageV0

	if err := scanner.Scan(&m.UID, &m.MessageID, &m.RemoteID, &m.Recent, &m.Deleted); err != nil {
		return uidMessageV0{}, err
	}

	return m, nil
}

func migrateMailboxMessages(
	ctx context.Context,
	tx utils.QueryWrapper,
	oldToNewIDMap map[imap.InternalMailboxID]imap.InternalMailboxID,
) error {
	// nolint:dupword
	loadMessagesQuery := fmt.Sprintf("SELECT u.`%v`, u.`%v`, m.`%v`, u.`%v`, u.`%v` FROM %v AS u "+
		"JOIN %v AS m ON m.`%v` = u.`%v` "+
		"WHERE u.`%v` = ? ORDER BY u.`%v`",
		v0.UIDsFieldUID,
		v0.UIDsFieldMessageID,
		v0.MessagesFieldRemoteID,
		v0.UIDsFieldRecent,
		v0.UIDsFieldDeleted,
		v0.UIDsTableName,
		v0.MessagesTableName,
		v0.MessagesFieldID,
		v0.UIDsFieldMessageID,
		v0.UIDsFieldMailboxID,
		v0.UIDsFieldUID,
	)

	loadMessagesStmt, err := tx.PrepareStatement(ctx, loadMessagesQuery)
	if err != nil {
		return err
	}

	defer utils.WrapStmtClose(loadMessagesStmt)

	for oldMboxID, newMBoxId := range oldToNewIDMap {
		messages, err := utils.MapStmtRowsFn(ctx, loadMessagesStmt, scanUIDMessageV0, oldMboxID)
		if err != nil {
			return err
		}

		for _, chunk := range xslices.Chunk(messages, db.ChunkLimit) {
			// Mailbox messages table.
			{
				query := fmt.Sprintf("INSERT OR IGNORE INTO %v (`%v`, `%v`, `%v`, `%v`) VALUES %v",
					MailboxMessageTableName(newMBoxId),
					MailboxMessagesFieldMessageID,
					MailboxMessagesFieldMessageRemoteID,
					MailboxMessagesFieldRecent,
					MailboxMessagesFieldDeleted,
					strings.Join(xslices.Repeat("(?,?,?,?)", len(chunk)), ","),
				)

				args := make([]any, 0, len(chunk)*4)
				for _, m := range chunk {
					args = append(args, m.MessageID, m.RemoteID, m.Recent, m.Deleted)
				}

				if _, err := utils.ExecQuery(ctx, tx, query, args...); err != nil {
					return err
				}
			}

			// Message to Mailbox table.
			{
				query := fmt.Sprintf("INSERT OR IGNORE INTO %v (`%v`, `%v`) VALUES %v",
					MessageToMailboxTableName,
					MessageToMailboxFieldMessageID,
					MessageToMailboxFieldMailboxID,
					strings.Join(xslices.Repeat("(?,?)", len(chunk)), ","),
				)

				args := make([]any, 0, len(chunk)*2)
				for _, m := range chunk {
					args = append(args, m.MessageID, newMBoxId)
				}

				if _, err := utils.ExecQuery(ctx, tx, query, args...); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
