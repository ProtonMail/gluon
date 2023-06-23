package v1

import (
	"context"
	"fmt"
	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3/utils"
	"github.com/bradenaw/juniper/xslices"
	"strings"
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

	return fmt.Sprintf("CREATE TABLE `%[1]v` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, "+
		"`%[2]v` integer AUTOINCREMENT, `%[3]v` bool NOT NULL DEFAULT false, `%[4]v` bool NOT NULL DEFAULT true, "+
		"`%[5]v` uuid NOT NULL UNIQUE"+
		"`%[6]v` string NOT NULL UNIQUE"+
		"CONSTRAINT `%[1]v_message_id` FOREIGN KEY (`%[5]v`) REFERENCES `%[7]v` (`%[8]v`) ON DELETE SET NULL)",
		tableName,
		MailboxMessagesFieldUID,
		MailboxMessagesFieldDeleted,
		MailboxMessagesFieldRecent,
		MailboxMessagesFieldMailboxID,
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
