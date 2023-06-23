package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	v1 "github.com/ProtonMail/gluon/internal/db_impl/sqlite3/v1"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3/utils"
	v0 "github.com/ProtonMail/gluon/internal/db_impl/sqlite3/v0"
	"github.com/bradenaw/juniper/xmaps"
	"github.com/bradenaw/juniper/xslices"
)

type readOps struct {
	qw utils.QueryWrapper
}

func (r readOps) MailboxExistsWithID(ctx context.Context, mboxID imap.InternalMailboxID) (bool, error) {
	query := fmt.Sprintf("SELEC 1 FROM %[1]v WHERE `%[2]v` = ? LIMIT 1",
		v0.MailboxesTableName,
		v0.MailboxesFieldID,
	)

	return utils.QueryExists(ctx, r.qw, query, mboxID)
}

func (r readOps) MailboxExistsWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %[1]v WHERE `%[2]v` = ? LIMIT 1",
		v0.MailboxesTableName,
		v0.MailboxesFieldRemoteID,
		v0.MessagesFieldID,
	)

	return utils.QueryExists(ctx, r.qw, query, mboxID)
}

func (r readOps) MailboxExistsWithName(ctx context.Context, name string) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %[1]v WHERE `%[2]v` = ? LIMIT 1",
		v0.MailboxesTableName,
		v0.MailboxesFieldName,
	)

	return utils.QueryExists(ctx, r.qw, query, name)
}

func (r readOps) GetMailboxIDFromRemoteID(ctx context.Context, mboxID imap.MailboxID) (imap.InternalMailboxID, error) {
	query := fmt.Sprintf("SELECT `%[2]v` FROM %[1]v WHERE `%[3]v` = ?",
		v0.MailboxesTableName,
		v0.MailboxesFieldID,
		v0.MailboxesFieldRemoteID,
	)

	return utils.MapQueryRow[imap.InternalMailboxID](ctx, r.qw, query, mboxID)
}

func (r readOps) GetMailboxName(ctx context.Context, mboxID imap.InternalMailboxID) (string, error) {
	query := fmt.Sprintf("SELECT `%[2]v` FROM %[1]v WHERE `%[3]v` = ?",
		v0.MailboxesTableName,
		v0.MailboxesFieldName,
		v0.MailboxesFieldID,
	)

	return utils.MapQueryRow[string](ctx, r.qw, query, mboxID)
}

func (r readOps) GetMailboxNameWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (string, error) {
	query := fmt.Sprintf("SELECT `%[2]v` FROM %[1]v WHERE `%[3]v` = ?",
		v0.MailboxesTableName,
		v0.MailboxesFieldName,
		v0.MailboxesFieldRemoteID,
	)

	return utils.MapQueryRow[string](ctx, r.qw, query, mboxID)
}

func (r readOps) GetMailboxMessageIDPairs(ctx context.Context, mboxID imap.InternalMailboxID) ([]db.MessageIDPair, error) {
	query := fmt.Sprintf("SELECT `%[2]v`, `%[3]v` FROM %[1]v WHERE `%[1]v`.`%[2]v` IN (SELECT `%[4]v`.`%[5]v` FROM %[4]v WHERE `%[4]v`.`%[6]v` = ?)",
		v1.MessagesTableName,
		v1.MessagesFieldID,
		v1.MessagesFieldRemoteID,
		v0.UIDsTableName,
		v0.UIDsFieldMessageID,
		v0.UIDsFieldMailboxID,
	)

	return utils.MapQueryRowsFn(ctx, r.qw, query, func(scanner utils.RowScanner) (db.MessageIDPair, error) {
		var id db.MessageIDPair

		if err := scanner.Scan(&id.InternalID, &id.RemoteID); err != nil {
			return db.MessageIDPair{}, err
		}

		return id, nil
	}, mboxID)
}

func (r readOps) GetAllMailboxesWithAttr(ctx context.Context) ([]*db.Mailbox, error) {
	query := fmt.Sprintf("SELECT * FROM %v", v0.MailboxesTableName)

	mailboxes, err := utils.MapQueryRowsFn(ctx, r.qw, query, ScanMailbox)
	if err != nil {
		return nil, err
	}

	attrQuery := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v0.MailboxAttrsFieldValue,
		v0.MailboxAttrsTableName,
		v0.MailboxAttrsFieldMailboxID,
	)

	stmt, err := r.qw.PrepareStatement(ctx, attrQuery)
	if err != nil {
		return nil, err
	}

	defer utils.WrapStmtClose(stmt)

	for _, mbox := range mailboxes {
		attrs, err := utils.MapStmtRows[string](ctx, stmt, mbox.ID)
		if err != nil {
			return nil, err
		}

		mbox.Attributes = xslices.Map(attrs, func(t string) *db.MailboxAttr {
			return &db.MailboxAttr{Value: t}
		})
	}

	return mailboxes, nil
}

func (r readOps) GetAllMailboxesAsRemoteIDs(ctx context.Context) ([]imap.MailboxID, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v", v0.MessagesFieldRemoteID, v0.MailboxesTableName)

	return utils.MapQueryRows[imap.MailboxID](ctx, r.qw, query)
}

func (r readOps) GetMailboxByName(ctx context.Context, name string) (*db.Mailbox, error) {
	query := fmt.Sprintf("SELECT * FROM %v WHERE `%v` = ?", v0.MailboxesTableName, v0.MailboxesFieldName)

	return utils.MapQueryRowFn(ctx, r.qw, query, ScanMailbox, name)
}

func (r readOps) GetMailboxByID(ctx context.Context, mboxID imap.InternalMailboxID) (*db.Mailbox, error) {
	query := fmt.Sprintf("SELECT * FROM %v WHERE `%v` = ?", v0.MailboxesTableName, v0.MailboxesFieldID)

	return utils.MapQueryRowFn(ctx, r.qw, query, ScanMailbox, mboxID)
}

func (r readOps) GetMailboxByRemoteID(ctx context.Context, mboxID imap.MailboxID) (*db.Mailbox, error) {
	query := fmt.Sprintf("SELECT * FROM %v WHERE `%v` = ?", v0.MailboxesTableName, v0.MailboxesFieldRemoteID)

	return utils.MapQueryRowFn(ctx, r.qw, query, ScanMailbox, mboxID)
}

func (r readOps) GetMailboxRecentCount(ctx context.Context, mboxID imap.InternalMailboxID) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE `%v` = TRUE AND `%v` = ?",
		v0.UIDsTableName,
		v0.UIDsFieldRecent,
		v0.UIDsFieldMailboxID,
	)

	return utils.MapQueryRow[int](ctx, r.qw, query, mboxID)
}

func (r readOps) GetMailboxMessageCount(ctx context.Context, mboxID imap.InternalMailboxID) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE `%v` = ?",
		v0.UIDsTableName,
		v0.UIDsFieldMailboxID,
	)

	return utils.MapQueryRow[int](ctx, r.qw, query, mboxID)
}

func (r readOps) GetMailboxMessageCountWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (int, error) {
	internalID, err := r.GetMailboxIDFromRemoteID(ctx, mboxID)
	if err != nil {
		return 0, err
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE `%v` = ?",
		v0.UIDsTableName,
		v0.UIDsFieldMailboxID,
	)

	return utils.MapQueryRow[int](ctx, r.qw, query, internalID)
}

func (r readOps) GetMailboxFlags(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v0.MailboxFlagsFieldValue,
		v0.MailboxFlagsTableName,
		v0.MailboxFlagsFieldMailboxID,
	)

	flags, err := utils.MapQueryRows[string](ctx, r.qw, query, mboxID)
	if err != nil {
		return imap.FlagSet{}, err
	}

	return imap.NewFlagSetFromSlice(flags), nil
}

func (r readOps) GetMailboxPermanentFlags(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v0.MailboxPermFlagsFieldValue,
		v0.MailboxPermFlagsTableName,
		v0.MailboxPermFlagsFieldMailboxID,
	)

	flags, err := utils.MapQueryRows[string](ctx, r.qw, query, mboxID)
	if err != nil {
		return imap.FlagSet{}, err
	}

	return imap.NewFlagSetFromSlice(flags), nil
}

func (r readOps) GetMailboxAttributes(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v0.MailboxAttrsFieldValue,
		v0.MailboxAttrsTableName,
		v0.MailboxAttrsFieldMailboxID,
	)

	flags, err := utils.MapQueryRows[string](ctx, r.qw, query, mboxID)
	if err != nil {
		return imap.FlagSet{}, err
	}

	return imap.NewFlagSetFromSlice(flags), nil
}

func (r readOps) GetMailboxUID(ctx context.Context, mboxID imap.InternalMailboxID) (imap.UID, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ? ",
		v0.MailboxesFieldUIDNext,
		v0.MailboxesTableName,
		v0.MailboxesFieldID,
	)

	return utils.MapQueryRow[imap.UID](ctx, r.qw, query, mboxID)
}

func (r readOps) GetMailboxMessageCountAndUID(ctx context.Context, mboxID imap.InternalMailboxID) (int, imap.UID, error) {
	count, err := r.GetMailboxMessageCount(ctx, mboxID)
	if err != nil {
		return 0, 0, err
	}

	uid, err := r.GetMailboxUID(ctx, mboxID)
	if err != nil {
		return 0, 0, err
	}

	return count, uid, nil
}

func (r readOps) GetMailboxMessageForNewSnapshot(ctx context.Context, mboxID imap.InternalMailboxID) ([]db.SnapshotMessageResult, error) {
	query := "SELECT `t1`.`remote_id`, GROUP_CONCAT(`t2`.`value`) AS `flags`, `ui_ds`.`recent`, `ui_ds`.`deleted`, `ui_ds`.`uid`, `ui_ds`.`uid_message` FROM `ui_ds`" +
		" JOIN `messages_v2` AS `t1` ON `ui_ds`.`uid_message` = `t1`.`id`" +
		" LEFT JOIN `message_flags_v2` AS `t2` ON `ui_ds`.`uid_message` = `t2`.`message_id` WHERE `mailbox_ui_ds` = ?" +
		" GROUP BY `ui_ds`.`uid_message` ORDER BY `ui_ds`.`uid`"

	return utils.MapQueryRowsFn(ctx, r.qw, query, func(scanner utils.RowScanner) (db.SnapshotMessageResult, error) {
		var r db.SnapshotMessageResult
		var flags sql.NullString

		if err := scanner.Scan(&r.RemoteID, &flags, &r.Recent, &r.Deleted, &r.UID, &r.InternalID); err != nil {
			return db.SnapshotMessageResult{}, err
		}

		r.Flags = flags.String

		return r, nil
	}, mboxID)
}

func (r readOps) MailboxTranslateRemoteIDs(ctx context.Context, mboxIDs []imap.MailboxID) ([]imap.InternalMailboxID, error) {
	result := make([]imap.InternalMailboxID, 0, len(mboxIDs))

	for _, chunk := range xslices.Chunk(mboxIDs, db.ChunkLimit) {
		query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` IN (%v)",
			v0.MailboxesFieldID,
			v0.MailboxesTableName,
			v0.MailboxesFieldRemoteID,
			utils.GenSQLIn(len(chunk)),
		)

		r, err := utils.MapQueryRows[imap.InternalMailboxID](ctx, r.qw, query, utils.MapSliceToAny(chunk)...)
		if err != nil {
			return nil, err
		}

		result = append(result, r...)
	}

	return result, nil
}

func (r readOps) MailboxFilterContains(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []db.MessageIDPair) ([]imap.InternalMessageID, error) {
	return r.MailboxFilterContainsInternalID(ctx, mboxID, xslices.Map(messageIDs, func(t db.MessageIDPair) imap.InternalMessageID {
		return t.InternalID
	}))
}

func (r readOps) MailboxFilterContainsInternalID(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]imap.InternalMessageID, error) {
	result := make([]imap.InternalMessageID, 0, len(messageIDs))

	for _, chunk := range xslices.Chunk(messageIDs, db.ChunkLimit) {
		query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` IN (%v) AND `%v` = ?",
			v0.UIDsFieldMessageID,
			v0.UIDsTableName,
			v0.UIDsFieldMessageID,
			utils.GenSQLIn(len(chunk)),
			v0.UIDsFieldMailboxID,
		)

		r, err := utils.MapQueryRows[imap.InternalMessageID](ctx, r.qw, query, append(utils.MapSliceToAny(chunk), mboxID)...)
		if err != nil {
			return nil, err
		}

		result = append(result, r...)
	}

	return result, nil
}

func (r readOps) GetMailboxCount(ctx context.Context) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %v", v0.MailboxesTableName)

	return utils.MapQueryRow[int](ctx, r.qw, query)
}

func (r readOps) GetMailboxMessageUIDsWithFlagsAfterAddOrUIDBump(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]db.UIDWithFlags, error) {
	result := make([]db.UIDWithFlags, 0, len(messageIDs))

	for _, chunk := range xslices.Chunk(messageIDs, db.ChunkLimit) {
		query := fmt.Sprintf("SELECT `t1`.`remote_id`, GROUP_CONCAT(`t2`.`value`) AS `flags`, `ui_ds`.`recent`, `ui_ds`.`deleted`, `ui_ds`.`uid`, `ui_ds`.`uid_message` FROM `ui_ds`"+
			" JOIN `messages_v2` AS `t1` ON `ui_ds`.`uid_message` = `t1`.`id`"+
			" LEFT JOIN `message_flags_v2` AS `t2` ON `ui_ds`.`uid_message` = `t2`.`message_id` WHERE `mailbox_ui_ds` = ? AND `uid_message` in (%v)"+
			" GROUP BY `ui_ds`.`uid_message` ORDER BY `ui_ds`.`uid`",
			utils.GenSQLIn(len(chunk)))

		args := make([]any, 0, len(chunk)+1)
		args = append(args, mboxID)
		args = append(args, utils.MapSliceToAny(chunk)...)

		r, err := utils.MapQueryRowsFn(ctx, r.qw, query, func(scanner utils.RowScanner) (db.UIDWithFlags, error) {
			var r db.UIDWithFlags
			var flags sql.NullString

			if err := scanner.Scan(&r.RemoteID, &flags, &r.Recent, &r.Deleted, &r.UID, &r.InternalID); err != nil {
				return db.UIDWithFlags{}, err
			}

			r.Flags = flags.String

			return r, nil
		}, args...)
		if err != nil {
			return nil, err
		}

		result = append(result, r...)
	}

	return result, nil
}

func (r readOps) MessageExists(ctx context.Context, id imap.InternalMessageID) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %v WHERE `%v` = ? LIMIT 1", v1.MessagesTableName, v1.MessagesFieldID)

	return utils.QueryExists(ctx, r.qw, query, id)
}

func (r readOps) MessageExistsWithRemoteID(ctx context.Context, id imap.MessageID) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %v WHERE `%v` = ? LIMIT 1", v1.MessagesTableName, v1.MessagesFieldRemoteID)

	return utils.QueryExists(ctx, r.qw, query, id)
}

func (r readOps) GetMessageNoEdges(ctx context.Context, id imap.InternalMessageID) (*db.Message, error) {
	query := fmt.Sprintf("SELECT * FROM %v WHERE `%v` = ?", v1.MessagesTableName, v1.MessagesFieldID)

	return utils.MapQueryRowFn(ctx, r.qw, query, ScanMessage, id)
}

func (r readOps) GetTotalMessageCount(ctx context.Context) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %v", v1.MessagesTableName)

	return utils.MapQueryRow[int](ctx, r.qw, query)
}

func (r readOps) GetMessageRemoteID(ctx context.Context, id imap.InternalMessageID) (imap.MessageID, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?", v1.MessagesFieldRemoteID, v1.MessagesTableName, v1.MessagesFieldID)

	return utils.MapQueryRow[imap.MessageID](ctx, r.qw, query, id)
}

func (r readOps) GetImportedMessageData(ctx context.Context, id imap.InternalMessageID) (*db.Message, error) {
	flagsQuery := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v1.MessageFlagsFieldValue,
		v1.MessageFlagsTableName,
		v1.MessageFlagsFieldMessageID,
	)

	messageQuery := fmt.Sprintf("SELECT * FROM %v WHERE `%v` = ?",
		v1.MessagesTableName,
		v1.MessagesFieldID,
	)

	msg, err := utils.MapQueryRowFn(ctx, r.qw, messageQuery, ScanMessage, id)
	if err != nil {
		return nil, err
	}

	flags, err := utils.MapQueryRowsFn(ctx, r.qw, flagsQuery, func(scanner utils.RowScanner) (*db.MessageFlag, error) {
		mf := new(db.MessageFlag)

		if err := scanner.Scan(&mf.Value); err != nil {
			return nil, err
		}

		return mf, nil
	}, id)
	if err != nil {
		return nil, err
	}

	msg.Flags = flags

	return msg, nil
}

func (r readOps) GetMessageDateAndSize(ctx context.Context, id imap.InternalMessageID) (time.Time, int, error) {
	query := fmt.Sprintf("SELECT `%v`, `%v` FROM %v WHERE `%v` =?",
		v1.MessagesFieldDate,
		v1.MessagesFieldSize,
		v1.MessagesTableName,
		v1.MessagesFieldID,
	)

	type DateSize struct {
		Date time.Time
		Size int
	}

	dt, err := utils.MapQueryRowFn(ctx, r.qw, query, func(scanner utils.RowScanner) (DateSize, error) {
		var dt DateSize

		if err := scanner.Scan(&dt.Date, &dt.Size); err != nil {
			return DateSize{}, err
		}

		return dt, nil
	}, id)
	if err != nil {
		return time.Time{}, 0, err
	}

	return dt.Date, dt.Size, nil
}

func (r readOps) GetMessageMailboxIDs(ctx context.Context, id imap.InternalMessageID) ([]imap.InternalMailboxID, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v0.UIDsFieldMailboxID,
		v0.UIDsTableName,
		v0.UIDsFieldMessageID,
	)

	return utils.MapQueryRows[imap.InternalMailboxID](ctx, r.qw, query, id)
}

func (r readOps) GetMessagesFlags(ctx context.Context, ids []imap.InternalMessageID) ([]db.MessageFlagSet, error) {
	var result = make([]db.MessageFlagSet, 0, len(ids))

	for _, chunk := range xslices.Chunk(ids, db.ChunkLimit) {
		flagQuery := fmt.Sprintf("SELECT GROUP_CONCAT(f.`%v`, ','), m.`%v`, m.`%v` FROM %v AS m "+
			"LEFT JOIN %v AS f ON f.`%v` = m.`%v` "+
			"WHERE m.`%v` IN (%v) "+
			"GROUP BY m.`%v`",
			v1.MessageFlagsFieldValue,
			v1.MessagesFieldID,
			v1.MessagesFieldRemoteID,
			v1.MessagesTableName,
			v1.MessageFlagsTableName,
			v1.MessageFlagsFieldMessageID,
			v1.MessagesFieldID,
			v1.MessagesFieldID,
			utils.GenSQLIn(len(chunk)),
			v1.MessagesFieldID,
		)

		args := utils.MapSliceToAny(chunk)

		type DBFlag struct {
			MessageID imap.InternalMessageID
			RemoteID  imap.MessageID
			Value     string
		}

		r, err := utils.MapQueryRowsFn(ctx, r.qw, flagQuery, func(scanner utils.RowScanner) (db.MessageFlagSet, error) {
			var f db.MessageFlagSet
			var flags sql.NullString

			if err := scanner.Scan(&flags, &f.ID, &f.RemoteID); err != nil {
				return db.MessageFlagSet{}, err
			}

			if flags.Valid {
				f.FlagSet = imap.NewFlagSetFromSlice(strings.Split(flags.String, ","))
			}

			return f, nil
		}, args...)
		if err != nil {
			return nil, err
		}

		result = append(result, r...)
	}

	return result, nil
}

func (r readOps) GetMessageIDsMarkedAsDelete(ctx context.Context) ([]imap.InternalMessageID, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = TRUE",
		v1.MessagesFieldID,
		v1.MessagesTableName,
		v1.MessagesFieldDeleted,
	)

	return utils.MapQueryRows[imap.InternalMessageID](ctx, r.qw, query)
}

func (r readOps) GetMessageIDFromRemoteID(ctx context.Context, id imap.MessageID) (imap.InternalMessageID, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v1.MessagesFieldID,
		v1.MessagesTableName,
		v1.MessagesFieldRemoteID,
	)

	return utils.MapQueryRow[imap.InternalMessageID](ctx, r.qw, query, id)
}

func (r readOps) GetMessageDeletedFlag(ctx context.Context, id imap.InternalMessageID) (bool, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v1.MessagesFieldDeleted,
		v1.MessagesTableName,
		v1.MessagesFieldID,
	)

	return utils.MapQueryRow[bool](ctx, r.qw, query, id)
}

func (r readOps) GetAllMessagesIDsAsMap(ctx context.Context) (map[imap.InternalMessageID]struct{}, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v", v1.MessagesFieldID, v1.MessagesTableName)

	ids, err := utils.MapQueryRows[imap.InternalMessageID](ctx, r.qw, query)
	if err != nil {
		return nil, err
	}

	return xmaps.SetFromSlice(ids), nil
}

func (r readOps) GetDeletedSubscriptionSet(ctx context.Context) (map[imap.MailboxID]*db.DeletedSubscription, error) {
	query := fmt.Sprintf("SELECT `%v`, `%v` FROM %v",
		v0.DeletedSubscriptionsFieldName,
		v0.DeletedSubscriptionsFieldRemoteID,
		v0.DeletedSubscriptionsTableName,
	)

	deletedSubscriptions, err := utils.MapQueryRowsFn(ctx, r.qw, query, func(scanner utils.RowScanner) (*db.DeletedSubscription, error) {
		ds := new(db.DeletedSubscription)

		if err := scanner.Scan(&ds.Name, &ds.RemoteID); err != nil {
			return nil, err
		}

		return ds, nil
	})
	if err != nil {
		return nil, err
	}

	result := make(map[imap.MailboxID]*db.DeletedSubscription, len(deletedSubscriptions))

	for _, v := range deletedSubscriptions {
		result[v.RemoteID] = v
	}

	return result, nil
}
