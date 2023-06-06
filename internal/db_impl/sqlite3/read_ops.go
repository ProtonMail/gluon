package sqlite3

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	v0 "github.com/ProtonMail/gluon/internal/db_impl/sqlite3/v0"
	"github.com/bradenaw/juniper/xmaps"
	"github.com/bradenaw/juniper/xslices"
)

type readOps struct {
	qw QueryWrapper
}

func (r readOps) MailboxExistsWithID(ctx context.Context, mboxID imap.InternalMailboxID) (bool, error) {
	query := fmt.Sprintf("SELEC 1 FROM %[1]v WHERE `%[2]v` = ? LIMIT 1",
		v0.MailboxesTableName,
		v0.MailboxesFieldID,
	)

	return QueryExists(ctx, r.qw, query, mboxID)
}

func (r readOps) MailboxExistsWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %[1]v WHERE `%[2]v` = ? LIMIT 1",
		v0.MailboxesTableName,
		v0.MailboxesFieldRemoteID,
		v0.MessagesFieldID,
	)

	return QueryExists(ctx, r.qw, query, mboxID)
}

func (r readOps) MailboxExistsWithName(ctx context.Context, name string) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %[1]v WHERE `%[2]v` = ? LIMIT 1",
		v0.MailboxesTableName,
		v0.MailboxesFieldName,
	)

	return QueryExists(ctx, r.qw, query, name)
}

func (r readOps) GetMailboxIDFromRemoteID(ctx context.Context, mboxID imap.MailboxID) (imap.InternalMailboxID, error) {
	query := fmt.Sprintf("SELECT `%[2]v` FROM %[1]v WHERE `%[3]v` = ?",
		v0.MailboxesTableName,
		v0.MailboxesFieldID,
		v0.MailboxesFieldRemoteID,
	)

	return MapQueryRow[imap.InternalMailboxID](ctx, r.qw, query, mboxID)
}

func (r readOps) GetMailboxName(ctx context.Context, mboxID imap.InternalMailboxID) (string, error) {
	query := fmt.Sprintf("SELECT `%[2]v` FROM %[1]v WHERE `%[3]v` = ?",
		v0.MailboxesTableName,
		v0.MailboxesFieldName,
		v0.MailboxesFieldID,
	)

	return MapQueryRow[string](ctx, r.qw, query, mboxID)
}

func (r readOps) GetMailboxNameWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (string, error) {
	query := fmt.Sprintf("SELECT `%[2]v` FROM %[1]v WHERE `%[3]v` = ?",
		v0.MailboxesTableName,
		v0.MailboxesFieldName,
		v0.MailboxesFieldRemoteID,
	)

	return MapQueryRow[string](ctx, r.qw, query, mboxID)
}

func (r readOps) GetMailboxMessageIDPairs(ctx context.Context, mboxID imap.InternalMailboxID) ([]db.MessageIDPair, error) {
	query := fmt.Sprintf("SELECT `%[2]v`, `%[3]v` FROM %[1]v WHERE `%[1]v`.`%[2]v` IN (SELECT `%[4]v`.`%[5]v` FROM %[4]v WHERE `%[4]v`.`%[6]v` = ?)",
		v0.MessagesTableName,
		v0.MessagesFieldID,
		v0.MessagesFieldRemoteID,
		v0.UIDsTableName,
		v0.UIDsFieldMessageID,
		v0.UIDsFieldMailboxID,
	)

	return MapQueryRowsFn(ctx, r.qw, query, func(scanner RowScanner) (db.MessageIDPair, error) {
		var id db.MessageIDPair

		if err := scanner.Scan(&id.InternalID, &id.RemoteID); err != nil {
			return db.MessageIDPair{}, err
		}

		return id, nil
	}, mboxID)
}

func (r readOps) GetAllMailboxesWithAttr(ctx context.Context) ([]*db.Mailbox, error) {
	query := fmt.Sprintf("SELECT * FROM %v", v0.MailboxesTableName)

	mailboxes, err := MapQueryRowsFn(ctx, r.qw, query, ScanMailbox)
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

	defer WrapStmtClose(stmt)

	for _, mbox := range mailboxes {
		attrs, err := MapStmtRows[string](ctx, stmt, mbox.ID)
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

	return MapQueryRows[imap.MailboxID](ctx, r.qw, query)
}

func (r readOps) GetMailboxByName(ctx context.Context, name string) (*db.Mailbox, error) {
	query := fmt.Sprintf("SELECT * FROM %v WHERE `%v` = ?", v0.MailboxesTableName, v0.MailboxesFieldName)

	return MapQueryRowFn(ctx, r.qw, query, ScanMailbox, name)
}

func (r readOps) GetMailboxByID(ctx context.Context, mboxID imap.InternalMailboxID) (*db.Mailbox, error) {
	query := fmt.Sprintf("SELECT * FROM %v WHERE `%v` = ?", v0.MailboxesTableName, v0.MailboxesFieldID)

	return MapQueryRowFn(ctx, r.qw, query, ScanMailbox, mboxID)
}

func (r readOps) GetMailboxByRemoteID(ctx context.Context, mboxID imap.MailboxID) (*db.Mailbox, error) {
	query := fmt.Sprintf("SELECT * FROM %v WHERE `%v` = ?", v0.MailboxesTableName, v0.MailboxesFieldRemoteID)

	return MapQueryRowFn(ctx, r.qw, query, ScanMailbox, mboxID)
}

func (r readOps) GetMailboxRecentCount(ctx context.Context, mboxID imap.InternalMailboxID) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE `%v` = TRUE AND `%v` = ?",
		v0.UIDsTableName,
		v0.UIDsFieldRecent,
		v0.UIDsFieldMailboxID,
	)

	return MapQueryRow[int](ctx, r.qw, query, mboxID)
}

func (r readOps) GetMailboxMessageCount(ctx context.Context, mboxID imap.InternalMailboxID) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE `%v` = ?",
		v0.UIDsTableName,
		v0.UIDsFieldMailboxID,
	)

	return MapQueryRow[int](ctx, r.qw, query, mboxID)
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

	return MapQueryRow[int](ctx, r.qw, query, internalID)
}

func (r readOps) GetMailboxFlags(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v0.MailboxFlagsFieldValue,
		v0.MailboxFlagsTableName,
		v0.MailboxFlagsFieldMailboxID,
	)

	flags, err := MapQueryRows[string](ctx, r.qw, query, mboxID)
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

	flags, err := MapQueryRows[string](ctx, r.qw, query, mboxID)
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

	flags, err := MapQueryRows[string](ctx, r.qw, query, mboxID)
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

	return MapQueryRow[imap.UID](ctx, r.qw, query, mboxID)
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
		" JOIN `messages` AS `t1` ON `ui_ds`.`uid_message` = `t1`.`id`" +
		" LEFT JOIN `message_flags` AS `t2` ON `ui_ds`.`uid_message` = `t2`.`message_flags` WHERE `mailbox_ui_ds` = ?" +
		" GROUP BY `ui_ds`.`uid_message` ORDER BY `ui_ds`.`uid`"

	return MapQueryRowsFn(ctx, r.qw, query, func(scanner RowScanner) (db.SnapshotMessageResult, error) {
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
			GenSQLIn(len(chunk)),
		)

		r, err := MapQueryRows[imap.InternalMailboxID](ctx, r.qw, query, MapSliceToAny(chunk)...)
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
			GenSQLIn(len(chunk)),
			v0.UIDsFieldMailboxID,
		)

		r, err := MapQueryRows[imap.InternalMessageID](ctx, r.qw, query, append(MapSliceToAny(chunk), mboxID)...)
		if err != nil {
			return nil, err
		}

		result = append(result, r...)
	}

	return result, nil
}

func (r readOps) GetMailboxCount(ctx context.Context) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %v", v0.MailboxesTableName)

	return MapQueryRow[int](ctx, r.qw, query)
}

func (r readOps) GetMailboxMessageUIDsWithFlagsAfterAddOrUIDBump(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]db.UIDWithFlags, error) {
	result := make([]db.UIDWithFlags, 0, len(messageIDs))

	for _, chunk := range xslices.Chunk(messageIDs, db.ChunkLimit) {
		query := fmt.Sprintf("SELECT `t1`.`remote_id`, GROUP_CONCAT(`t2`.`value`) AS `flags`, `ui_ds`.`recent`, `ui_ds`.`deleted`, `ui_ds`.`uid`, `ui_ds`.`uid_message` FROM `ui_ds`"+
			" JOIN `messages` AS `t1` ON `ui_ds`.`uid_message` = `t1`.`id`"+
			" LEFT JOIN `message_flags` AS `t2` ON `ui_ds`.`uid_message` = `t2`.`message_flags` WHERE `mailbox_ui_ds` = ? AND `uid_message` in (%v)"+
			" GROUP BY `ui_ds`.`uid_message` ORDER BY `ui_ds`.`uid`",
			GenSQLIn(len(chunk)))

		args := make([]any, 0, len(chunk)+1)
		args = append(args, mboxID)
		args = append(args, MapSliceToAny(chunk)...)

		r, err := MapQueryRowsFn(ctx, r.qw, query, func(scanner RowScanner) (db.UIDWithFlags, error) {
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
	query := fmt.Sprintf("SELECT 1 FROM %v WHERE `%v` = ? LIMIT 1", v0.MessagesTableName, v0.MessagesFieldID)

	return QueryExists(ctx, r.qw, query, id)
}

func (r readOps) MessageExistsWithRemoteID(ctx context.Context, id imap.MessageID) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %v WHERE `%v` = ? LIMIT 1", v0.MessagesTableName, v0.MessagesFieldRemoteID)

	return QueryExists(ctx, r.qw, query, id)
}

func (r readOps) GetMessageNoEdges(ctx context.Context, id imap.InternalMessageID) (*db.Message, error) {
	query := fmt.Sprintf("SELECT * FROM %v WHERE `%v` = ?", v0.MessagesTableName, v0.MessagesFieldID)

	return MapQueryRowFn(ctx, r.qw, query, ScanMessage, id)
}

func (r readOps) GetTotalMessageCount(ctx context.Context) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %v", v0.MessagesTableName)

	return MapQueryRow[int](ctx, r.qw, query)
}

func (r readOps) GetMessageRemoteID(ctx context.Context, id imap.InternalMessageID) (imap.MessageID, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?", v0.MessagesFieldRemoteID, v0.MessagesTableName, v0.MessagesFieldID)

	return MapQueryRow[imap.MessageID](ctx, r.qw, query, id)
}

func (r readOps) GetImportedMessageData(ctx context.Context, id imap.InternalMessageID) (*db.Message, error) {
	flagsQuery := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v0.MessageFlagsFieldValue,
		v0.MessageFlagsTableName,
		v0.MessageFlagsFieldMessageID,
	)

	messageQuery := fmt.Sprintf("SELECT * FROM %v WHERE `%v` = ?",
		v0.MessagesTableName,
		v0.MessagesFieldID,
	)

	msg, err := MapQueryRowFn(ctx, r.qw, messageQuery, ScanMessage, id)
	if err != nil {
		return nil, err
	}

	flags, err := MapQueryRowsFn(ctx, r.qw, flagsQuery, func(scanner RowScanner) (*db.MessageFlag, error) {
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
		v0.MessagesFieldDate,
		v0.MessagesFieldSize,
		v0.MessagesTableName,
		v0.MessagesFieldID,
	)

	type DateSize struct {
		Date time.Time
		Size int
	}

	dt, err := MapQueryRowFn(ctx, r.qw, query, func(scanner RowScanner) (DateSize, error) {
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

	return MapQueryRows[imap.InternalMailboxID](ctx, r.qw, query, id)
}

func (r readOps) GetMessagesFlags(ctx context.Context, ids []imap.InternalMessageID) ([]db.MessageFlagSet, error) {
	var result = make([]db.MessageFlagSet, 0, len(ids))

	flagQuery := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v0.MessageFlagsFieldValue,
		v0.MessageFlagsTableName,
		v0.MessageFlagsFieldMessageID,
	)

	remoteIDQuery := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v0.MessagesFieldRemoteID,
		v0.MessagesTableName,
		v0.MessagesFieldID,
	)

	flagStmt, err := r.qw.PrepareStatement(ctx, flagQuery)
	if err != nil {
		return nil, err
	}

	defer WrapStmtClose(flagStmt)

	remoteIDStmt, err := r.qw.PrepareStatement(ctx, remoteIDQuery)
	if err != nil {
		return nil, err
	}

	defer WrapStmtClose(remoteIDStmt)

	// GODT:2522 - Would SELECT GROUP BY id and then reconstructing the flag list over that be faster?
	// GODT:2522 - Store remote ID in message flags

	for _, id := range ids {
		flags, err := MapStmtRows[string](ctx, flagStmt, id)
		if err != nil {
			return nil, err
		}

		remoteID, err := MapStmtRow[imap.MessageID](ctx, remoteIDStmt, id)
		if err != nil {
			return nil, err
		}

		result = append(result, db.MessageFlagSet{
			ID:       id,
			RemoteID: remoteID,
			FlagSet:  imap.NewFlagSetFromSlice(flags),
		})
	}

	return result, nil
}

func (r readOps) GetMessageIDsMarkedAsDelete(ctx context.Context) ([]imap.InternalMessageID, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = TRUE",
		v0.MessagesFieldID,
		v0.MessagesTableName,
		v0.MessagesFieldDeleted,
	)

	return MapQueryRows[imap.InternalMessageID](ctx, r.qw, query)
}

func (r readOps) GetMessageIDFromRemoteID(ctx context.Context, id imap.MessageID) (imap.InternalMessageID, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v0.MessagesFieldID,
		v0.MessagesTableName,
		v0.MessagesFieldRemoteID,
	)

	return MapQueryRow[imap.InternalMessageID](ctx, r.qw, query, id)
}

func (r readOps) GetMessageDeletedFlag(ctx context.Context, id imap.InternalMessageID) (bool, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v0.MessagesFieldDeleted,
		v0.MessagesTableName,
		v0.MessagesFieldID,
	)

	return MapQueryRow[bool](ctx, r.qw, query, id)
}

func (r readOps) GetAllMessagesIDsAsMap(ctx context.Context) (map[imap.InternalMessageID]struct{}, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v", v0.MessagesFieldID, v0.MessagesTableName)

	ids, err := MapQueryRows[imap.InternalMessageID](ctx, r.qw, query)
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

	deletedSubscriptions, err := MapQueryRowsFn(ctx, r.qw, query, func(scanner RowScanner) (*db.DeletedSubscription, error) {
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
