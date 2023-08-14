package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	v2 "github.com/ProtonMail/gluon/internal/db_impl/sqlite3/v2"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3/utils"
	v1 "github.com/ProtonMail/gluon/internal/db_impl/sqlite3/v1"
	"github.com/bradenaw/juniper/xmaps"
	"github.com/bradenaw/juniper/xslices"
)

type readOps struct {
	qw utils.QueryWrapper
}

func (r readOps) MailboxExistsWithID(ctx context.Context, mboxID imap.InternalMailboxID) (bool, error) {
	query := fmt.Sprintf("SELEC 1 FROM %[1]v WHERE `%[2]v` = ? LIMIT 1",
		v1.MailboxesTableName,
		v1.MailboxesFieldID,
	)

	return utils.QueryExists(ctx, r.qw, query, mboxID)
}

func (r readOps) MailboxExistsWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %[1]v WHERE `%[2]v` = ? LIMIT 1",
		v1.MailboxesTableName,
		v1.MailboxesFieldRemoteID,
		v1.MessagesFieldID,
	)

	return utils.QueryExists(ctx, r.qw, query, mboxID)
}

func (r readOps) MailboxExistsWithName(ctx context.Context, name string) (bool, error) {
	query := fmt.Sprintf("SELECT 1 FROM %[1]v WHERE `%[2]v` = ? LIMIT 1",
		v1.MailboxesTableName,
		v1.MailboxesFieldName,
	)

	return utils.QueryExists(ctx, r.qw, query, name)
}

func (r readOps) GetMailboxIDFromRemoteID(ctx context.Context, mboxID imap.MailboxID) (imap.InternalMailboxID, error) {
	query := fmt.Sprintf("SELECT `%[2]v` FROM %[1]v WHERE `%[3]v` = ?",
		v1.MailboxesTableName,
		v1.MailboxesFieldID,
		v1.MailboxesFieldRemoteID,
	)

	return utils.MapQueryRow[imap.InternalMailboxID](ctx, r.qw, query, mboxID)
}

func (r readOps) GetMailboxName(ctx context.Context, mboxID imap.InternalMailboxID) (string, error) {
	query := fmt.Sprintf("SELECT `%[2]v` FROM %[1]v WHERE `%[3]v` = ?",
		v1.MailboxesTableName,
		v1.MailboxesFieldName,
		v1.MailboxesFieldID,
	)

	return utils.MapQueryRow[string](ctx, r.qw, query, mboxID)
}

func (r readOps) GetMailboxNameWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (string, error) {
	query := fmt.Sprintf("SELECT `%[2]v` FROM %[1]v WHERE `%[3]v` = ?",
		v1.MailboxesTableName,
		v1.MailboxesFieldName,
		v1.MailboxesFieldRemoteID,
	)

	return utils.MapQueryRow[string](ctx, r.qw, query, mboxID)
}

func (r readOps) GetMailboxMessageIDPairs(ctx context.Context, mboxID imap.InternalMailboxID) ([]db.MessageIDPair, error) {
	query := fmt.Sprintf("SELECT `%[2]v`, `%[3]v` FROM %[1]v",
		v1.MailboxMessageTableName(mboxID),
		v1.MailboxMessagesFieldMessageID,
		v1.MailboxMessagesFieldMessageRemoteID,
	)

	return utils.MapQueryRowsFn(ctx, r.qw, query, func(scanner utils.RowScanner) (db.MessageIDPair, error) {
		var id db.MessageIDPair

		if err := scanner.Scan(&id.InternalID, &id.RemoteID); err != nil {
			return db.MessageIDPair{}, err
		}

		return id, nil
	})
}

func (r readOps) GetAllMailboxesWithAttr(ctx context.Context) ([]*db.MailboxWithAttr, error) {
	query := fmt.Sprintf("SELECT * FROM %v", v1.MailboxesTableName)

	mailboxes, err := utils.MapQueryRowsFn(ctx, r.qw, query, ScanMailboxWithAttr)
	if err != nil {
		return nil, err
	}

	attrQuery := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v1.MailboxAttrsFieldValue,
		v1.MailboxAttrsTableName,
		v1.MailboxAttrsFieldMailboxID,
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

		mbox.Attributes = imap.NewFlagSet(attrs...)
	}

	return mailboxes, nil
}

func (r readOps) GetAllMailboxesAsRemoteIDs(ctx context.Context) ([]imap.MailboxID, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v", v1.MessagesFieldRemoteID, v1.MailboxesTableName)

	return utils.MapQueryRows[imap.MailboxID](ctx, r.qw, query)
}

func (r readOps) GetMailboxByName(ctx context.Context, name string) (*db.Mailbox, error) {
	query := fmt.Sprintf("SELECT * FROM %v WHERE `%v` = ?", v1.MailboxesTableName, v1.MailboxesFieldName)

	return utils.MapQueryRowFn(ctx, r.qw, query, ScanMailbox, name)
}

func (r readOps) GetMailboxByID(ctx context.Context, mboxID imap.InternalMailboxID) (*db.Mailbox, error) {
	query := fmt.Sprintf("SELECT * FROM %v WHERE `%v` = ?", v1.MailboxesTableName, v1.MailboxesFieldID)

	return utils.MapQueryRowFn(ctx, r.qw, query, ScanMailbox, mboxID)
}

func (r readOps) GetMailboxByRemoteID(ctx context.Context, mboxID imap.MailboxID) (*db.Mailbox, error) {
	query := fmt.Sprintf("SELECT * FROM %v WHERE `%v` = ?", v1.MailboxesTableName, v1.MailboxesFieldRemoteID)

	return utils.MapQueryRowFn(ctx, r.qw, query, ScanMailbox, mboxID)
}

func (r readOps) GetMailboxRecentCount(ctx context.Context, mboxID imap.InternalMailboxID) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %v WHERE `%v` = TRUE",
		v1.MailboxMessageTableName(mboxID),
		v1.MailboxMessagesFieldRecent,
	)

	return utils.MapQueryRow[int](ctx, r.qw, query)
}

func (r readOps) GetMailboxMessageCount(ctx context.Context, mboxID imap.InternalMailboxID) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %v",
		v1.MailboxMessageTableName(mboxID),
	)

	return utils.MapQueryRow[int](ctx, r.qw, query)
}

func (r readOps) GetMailboxMessageCountWithRemoteID(ctx context.Context, mboxID imap.MailboxID) (int, error) {
	internalID, err := r.GetMailboxIDFromRemoteID(ctx, mboxID)
	if err != nil {
		return 0, err
	}

	query := fmt.Sprintf("SELECT COUNT(*) FROM %v",
		v1.MailboxMessageTableName(internalID),
	)

	return utils.MapQueryRow[int](ctx, r.qw, query)
}

func (r readOps) GetMailboxFlags(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v1.MailboxFlagsFieldValue,
		v1.MailboxFlagsTableName,
		v1.MailboxFlagsFieldMailboxID,
	)

	flags, err := utils.MapQueryRows[string](ctx, r.qw, query, mboxID)
	if err != nil {
		return imap.FlagSet{}, err
	}

	return imap.NewFlagSetFromSlice(flags), nil
}

func (r readOps) GetMailboxPermanentFlags(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v1.MailboxPermFlagsFieldValue,
		v1.MailboxPermFlagsTableName,
		v1.MailboxPermFlagsFieldMailboxID,
	)

	flags, err := utils.MapQueryRows[string](ctx, r.qw, query, mboxID)
	if err != nil {
		return imap.FlagSet{}, err
	}

	return imap.NewFlagSetFromSlice(flags), nil
}

func (r readOps) GetMailboxAttributes(ctx context.Context, mboxID imap.InternalMailboxID) (imap.FlagSet, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v1.MailboxAttrsFieldValue,
		v1.MailboxAttrsTableName,
		v1.MailboxAttrsFieldMailboxID,
	)

	flags, err := utils.MapQueryRows[string](ctx, r.qw, query, mboxID)
	if err != nil {
		return imap.FlagSet{}, err
	}

	return imap.NewFlagSetFromSlice(flags), nil
}

func (r readOps) GetMailboxUID(ctx context.Context, mboxID imap.InternalMailboxID) (imap.UID, error) {
	query := "SELECT `seq` FROM sqlite_sequence WHERE `name` = ?"

	// Until a value is inserted in to mailbox the sequence table will not yet be initialized.
	uid, err := utils.MapQueryRow[imap.UID](ctx, r.qw, query, v1.MailboxMessageTableName(mboxID))
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return imap.UID(1), nil
		}

		return 0, err
	}

	return uid.Add(1), nil
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
	query := fmt.Sprintf("SELECT `m`.`%[1]v`, GROUP_CONCAT(`f`.`%[2]v`) AS `flags`, `m`.`%[3]v`, `m`.`%[4]v`, "+
		"`m`.`%[5]v`, `m`.`%[6]v` FROM %[9]v AS m "+
		"LEFT JOIN `%[7]v` AS f ON `f`.`%[8]v` = `m`.`%[6]v` "+
		"GROUP BY `m`.`%[6]v` ORDER BY `m`.`%[5]v`",
		v1.MailboxMessagesFieldMessageRemoteID,
		v1.MessageFlagsFieldValue,
		v1.MailboxMessagesFieldRecent,
		v1.MailboxMessagesFieldDeleted,
		v1.MailboxMessagesFieldUID,
		v1.MailboxMessagesFieldMessageID,
		v1.MessageFlagsTableName,
		v1.MessageFlagsFieldMessageID,
		v1.MailboxMessageTableName(mboxID),
	)

	return utils.MapQueryRowsFn(ctx, r.qw, query, func(scanner utils.RowScanner) (db.SnapshotMessageResult, error) {
		var r db.SnapshotMessageResult
		var flags sql.NullString

		if err := scanner.Scan(&r.RemoteID, &flags, &r.Recent, &r.Deleted, &r.UID, &r.InternalID); err != nil {
			return db.SnapshotMessageResult{}, err
		}

		r.Flags = flags.String

		return r, nil
	})
}

func (r readOps) MailboxTranslateRemoteIDs(ctx context.Context, mboxIDs []imap.MailboxID) ([]imap.InternalMailboxID, error) {
	result := make([]imap.InternalMailboxID, 0, len(mboxIDs))

	for _, chunk := range xslices.Chunk(mboxIDs, db.ChunkLimit) {
		query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` IN (%v)",
			v1.MailboxesFieldID,
			v1.MailboxesTableName,
			v1.MailboxesFieldRemoteID,
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
		query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` IN (%v)",
			v1.MailboxMessagesFieldMessageID,
			v1.MailboxMessageTableName(mboxID),
			v1.MailboxMessagesFieldMessageID,
			utils.GenSQLIn(len(chunk)),
		)

		r, err := utils.MapQueryRows[imap.InternalMessageID](ctx, r.qw, query, utils.MapSliceToAny(chunk)...)
		if err != nil {
			return nil, err
		}

		result = append(result, r...)
	}

	return result, nil
}

func (r readOps) GetMailboxCount(ctx context.Context) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %v", v1.MailboxesTableName)

	return utils.MapQueryRow[int](ctx, r.qw, query)
}

func (r readOps) GetMailboxMessageUIDsWithFlagsAfterAddOrUIDBump(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]db.UIDWithFlags, error) {
	result := make([]db.UIDWithFlags, 0, len(messageIDs))

	for _, chunk := range xslices.Chunk(messageIDs, db.ChunkLimit) {
		query := fmt.Sprintf("SELECT `m`.`%[1]v`, GROUP_CONCAT(`f`.`%[2]v`) AS `flags`, `m`.`%[3]v`, `m`.`%[4]v`, "+
			"`m`.`%[5]v`, `m`.`%[6]v` FROM %[9]v AS m "+
			"LEFT JOIN `%[7]v` AS f ON `f`.`%[8]v` = `m`.`%[6]v` "+
			"WHERE `m`.`%[6]v` IN (%[10]v) "+
			"GROUP BY `m`.`%[6]v` ORDER BY `m`.`%[5]v`",
			v1.MailboxMessagesFieldMessageRemoteID,
			v1.MessageFlagsFieldValue,
			v1.MailboxMessagesFieldRecent,
			v1.MailboxMessagesFieldDeleted,
			v1.MailboxMessagesFieldUID,
			v1.MailboxMessagesFieldMessageID,
			v1.MessageFlagsTableName,
			v1.MessageFlagsFieldMessageID,
			v1.MailboxMessageTableName(mboxID),
			utils.GenSQLIn(len(chunk)),
		)

		args := utils.MapSliceToAny(chunk)

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

func (r readOps) GetImportedMessageData(ctx context.Context, id imap.InternalMessageID) (*db.MessageWithFlags, error) {
	flagsQuery := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v1.MessageFlagsFieldValue,
		v1.MessageFlagsTableName,
		v1.MessageFlagsFieldMessageID,
	)

	messageQuery := fmt.Sprintf("SELECT * FROM %v WHERE `%v` = ?",
		v1.MessagesTableName,
		v1.MessagesFieldID,
	)

	msg, err := utils.MapQueryRowFn(ctx, r.qw, messageQuery, ScanMessageWithFlags, id)
	if err != nil {
		return nil, err
	}

	flags, err := utils.MapQueryRows[string](ctx, r.qw, flagsQuery, id)
	if err != nil {
		return nil, err
	}

	msg.Flags = imap.NewFlagSet(flags...)

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
	query := fmt.Sprintf("SELECT `%[3]v` FROM %[1]v WHERE `%[2]v` = ?",
		v1.MessageToMailboxTableName,
		v1.MessageToMailboxFieldMessageID,
		v1.MessageToMailboxFieldMailboxID,
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
		v1.DeletedSubscriptionsFieldName,
		v1.DeletedSubscriptionsFieldRemoteID,
		v1.DeletedSubscriptionsTableName,
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

func (r readOps) GetConnectorSettings(ctx context.Context) (string, bool, error) {
	query := fmt.Sprintf("SELECT `%v` FROM %v WHERE `%v` = ?",
		v2.ConnectorSettingsFieldValue,
		v2.ConnectorSettingsTableName,
		v2.ConnectorSettingsFieldID,
	)

	var hasValue bool

	value, err := utils.MapQueryRowFn(ctx, r.qw, query, func(scanner utils.RowScanner) (string, error) {
		var value sql.NullString

		if err := scanner.Scan(&value); err != nil {
			return "", err
		}

		if !value.Valid {
			return "", nil
		}

		hasValue = true
		return value.String, nil
	}, v2.ConnectorSettingsDefaultID)

	return value, hasValue, err
}
