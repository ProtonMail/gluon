package sqlite3

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3/utils"
	v1 "github.com/ProtonMail/gluon/internal/db_impl/sqlite3/v1"
	v2 "github.com/ProtonMail/gluon/internal/db_impl/sqlite3/v2"
	"github.com/bradenaw/juniper/xslices"
)

type writeOps struct {
	readOps
	qw utils.QueryWrapper
}

func (w writeOps) CreateMailbox(
	ctx context.Context,
	mboxID imap.MailboxID,
	name string,
	flags, permFlags, attrs imap.FlagSet,
	uidValidity imap.UID,
) (*db.Mailbox, error) {
	createMBoxQuery := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`, `%v`, `%v`) VALUES (?,?,?,?) RETURNING `%v`",
		v1.MailboxesTableName,
		v1.MailboxesFieldRemoteID,
		v1.MailboxesFieldName,
		v1.MailboxesFieldUIDValidity,
		v1.MailboxesFieldSubscribed,
		v1.MailboxesFieldID,
	)

	internalID, err := utils.MapQueryRow[imap.InternalMailboxID](ctx, w.qw, createMBoxQuery,
		mboxID,
		name,
		uidValidity,
		true,
	)
	if err != nil {
		return nil, err
	}

	{
		query := v1.CreateMailboxMessageTableQuery(internalID)

		if _, err := utils.ExecQuery(ctx, w.qw, query); err != nil {
			return nil, err
		}
	}

	createFlags := func(tableName, fieldID, fieldValue string, flags imap.FlagSet) error {
		query := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES (?, ?)",
			tableName,
			fieldID,
			fieldValue,
		)

		stmt, err := w.qw.PrepareStatement(ctx, query)
		if err != nil {
			return err
		}

		defer utils.WrapStmtClose(stmt)

		for _, f := range flags.ToSliceUnsorted() {
			if _, err := utils.ExecStmt(ctx, stmt, internalID, f); err != nil {
				return err
			}
		}

		return nil
	}

	if err := createFlags(v1.MailboxFlagsTableName, v1.MailboxFlagsFieldMailboxID, v1.MailboxFlagsFieldValue, flags); err != nil {
		return nil, err
	}

	if err := createFlags(v1.MailboxPermFlagsTableName, v1.MailboxPermFlagsFieldMailboxID, v1.MailboxPermFlagsFieldValue, permFlags); err != nil {
		return nil, err
	}

	if err := createFlags(v1.MailboxAttrsTableName, v1.MailboxAttrsFieldMailboxID, v1.MailboxAttrsFieldValue, attrs); err != nil {
		return nil, err
	}

	return &db.Mailbox{
		ID:          internalID,
		RemoteID:    mboxID,
		Name:        name,
		UIDValidity: uidValidity,
		Subscribed:  true,
	}, nil
}

func (w writeOps) GetOrCreateMailbox(
	ctx context.Context,
	mboxID imap.MailboxID,
	name string,
	flags, permFlags, attrs imap.FlagSet,
	uidValidity imap.UID,
) (*db.Mailbox, error) {
	mbox, err := w.GetMailboxByRemoteID(ctx, mboxID)
	if err != nil {
		if !errors.Is(err, db.ErrNotFound) {
			return nil, err
		}
	} else {
		return mbox, nil
	}

	return w.CreateMailbox(ctx, mboxID, name, flags, permFlags, attrs, uidValidity)
}

func (w writeOps) GetOrCreateMailboxAlt(ctx context.Context, mbox imap.Mailbox, delimiter string, uidValidity imap.UID) (*db.Mailbox, error) {
	return w.GetOrCreateMailbox(
		ctx,
		mbox.ID,
		strings.Join(mbox.Name, delimiter),
		mbox.Flags,
		mbox.PermanentFlags,
		mbox.Attributes,
		uidValidity,
	)
}

func (w writeOps) RenameMailboxWithRemoteID(ctx context.Context, mboxID imap.MailboxID, name string) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = ? WHERE `%v` = ?",
		v1.MailboxesTableName,
		v1.MailboxesFieldName,
		v1.MailboxesFieldRemoteID,
	)

	return utils.ExecQueryAndCheckUpdatedNotZero(ctx, w.qw, query, name, mboxID)
}

func (w writeOps) DeleteMailboxWithRemoteID(ctx context.Context, mboxID imap.MailboxID) error {
	mbox, err := w.GetMailboxByRemoteID(ctx, mboxID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil
		}

		return err
	}

	if mbox.Subscribed {
		if err := w.AddDeletedSubscription(ctx, mbox.Name, mboxID); err != nil {
			return err
		}
	}

	{
		query := fmt.Sprintf("DROP TABLE `%v`", v1.MailboxMessageTableName(mbox.ID))

		if _, err = utils.ExecQuery(ctx, w.qw, query); err != nil {
			return err
		}
	}
	{
		query := fmt.Sprintf("DELETE FROM %v WHERE `%v` = ?",
			v1.MailboxesTableName,
			v1.MailboxesFieldRemoteID)

		if _, err = utils.ExecQuery(ctx, w.qw, query, mboxID); err != nil {
			return err
		}
	}

	return nil
}

func (w writeOps) AddMessagesToMailbox(
	ctx context.Context,
	mboxID imap.InternalMailboxID,
	messageIDs []db.MessageIDPair,
) ([]db.UIDWithFlags, error) {
	if len(messageIDs) == 0 {
		return nil, nil
	}

	for _, chunk := range xslices.Chunk(messageIDs, db.ChunkLimit/2) {
		// Insert into Mailbox table.
		{
			query := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES %v",
				v1.MailboxMessageTableName(mboxID),
				v1.MailboxMessagesFieldMessageID,
				v1.MailboxMessagesFieldMessageRemoteID,
				strings.Join(xslices.Repeat("(?,?)", len(chunk)), ","),
			)

			args := make([]any, 0, 2*len(chunk))

			for _, id := range chunk {
				args = append(args, id.InternalID, id.RemoteID)
			}

			if _, err := utils.ExecQuery(ctx, w.qw, query, args...); err != nil {
				return nil, err
			}
		}

		// Insert into Message To Mailbox table.
		{
			query := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES %v",
				v1.MessageToMailboxTableName,
				v1.MessageToMailboxFieldMessageID,
				v1.MessageToMailboxFieldMailboxID,
				strings.Join(xslices.Repeat("(?,?)", len(chunk)), ","),
			)

			args := make([]any, 0, 2*len(chunk))

			for _, id := range chunk {
				args = append(args, id.InternalID, mboxID)
			}

			if _, err := utils.ExecQuery(ctx, w.qw, query, args...); err != nil {
				return nil, err
			}
		}
	}

	return w.GetMailboxMessageUIDsWithFlagsAfterAddOrUIDBump(ctx, mboxID, xslices.Map(messageIDs, func(t db.MessageIDPair) imap.InternalMessageID {
		return t.InternalID
	}))
}

func (w writeOps) RemoveMessagesFromMailbox(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) error {
	for _, chunk := range xslices.Chunk(messageIDs, db.ChunkLimit) {
		// Delete from mailbox table.
		{
			query := fmt.Sprintf("DELETE FROM %v WHERE `%v` IN (%v)",
				v1.MailboxMessageTableName(mboxID),
				v1.MailboxMessagesFieldMessageID,
				utils.GenSQLIn(len(chunk)),
			)

			if _, err := utils.ExecQuery(ctx, w.qw, query, utils.MapSliceToAny(messageIDs)...); err != nil {
				return err
			}
		}

		// Delete from message to mailbox table.
		{
			query := fmt.Sprintf("DELETE FROM %v WHERE `%v` IN (%v) AND `%v` =?",
				v1.MessageToMailboxTableName,
				v1.MailboxMessagesFieldMessageID,
				utils.GenSQLIn(len(chunk)),
				v1.MessageToMailboxFieldMailboxID,
			)

			if _, err := utils.ExecQuery(ctx, w.qw, query, append(utils.MapSliceToAny(messageIDs), mboxID)...); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w writeOps) ClearRecentFlagInMailboxOnMessage(ctx context.Context, mboxID imap.InternalMailboxID, messageID imap.InternalMessageID) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = FALSE WHERE `%v` = ?",
		v1.MailboxMessageTableName(mboxID),
		v1.MailboxMessagesFieldRecent,
		v1.MailboxMessagesFieldMessageID,
	)

	_, err := utils.ExecQuery(ctx, w.qw, query, messageID)

	return err
}

func (w writeOps) ClearRecentFlagsInMailbox(ctx context.Context, mboxID imap.InternalMailboxID) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = FALSE WHERE `%v` = TRUE",
		v1.MailboxMessageTableName(mboxID),
		v1.MailboxMessagesFieldRecent,
		v1.MailboxMessagesFieldRecent,
	)

	_, err := utils.ExecQuery(ctx, w.qw, query)

	return err
}

func (w writeOps) CreateMailboxIfNotExists(ctx context.Context, mbox imap.Mailbox, delimiter string, uidValidity imap.UID) error {
	_, err := w.GetOrCreateMailboxAlt(ctx, mbox, delimiter, uidValidity)

	return err
}

func (w writeOps) SetMailboxMessagesDeletedFlag(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID, deleted bool) error {
	for _, chunk := range xslices.Chunk(messageIDs, db.ChunkLimit) {
		query := fmt.Sprintf("UPDATE %v SET `%v` =? WHERE `%v` IN (%v)",
			v1.MailboxMessageTableName(mboxID),
			v1.MailboxMessagesFieldDeleted,
			v1.MailboxMessagesFieldMessageID,
			utils.GenSQLIn(len(chunk)),
		)

		args := make([]any, 0, len(chunk)+2)
		args = append(args, deleted)
		args = append(args, utils.MapSliceToAny(chunk)...)

		if _, err := utils.ExecQuery(ctx, w.qw, query, args...); err != nil {
			return err
		}
	}

	return nil
}

func (w writeOps) SetMailboxSubscribed(ctx context.Context, mboxID imap.InternalMailboxID, subscribed bool) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = ? WHERE `%v` = ?",
		v1.MailboxesTableName,
		v1.MailboxesFieldSubscribed,
		v1.MailboxesFieldID,
	)

	_, err := utils.ExecQuery(ctx, w.qw, query, subscribed, mboxID)

	return err
}

func (w writeOps) UpdateRemoteMailboxID(ctx context.Context, mboxID imap.InternalMailboxID, remoteID imap.MailboxID) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = ? WHERE `%v` = ?",
		v1.MailboxesTableName,
		v1.MailboxesFieldRemoteID,
		v1.MailboxesFieldID,
	)

	return utils.ExecQueryAndCheckUpdatedNotZero(ctx, w.qw, query, remoteID, mboxID)
}

func (w writeOps) SetMailboxUIDValidity(ctx context.Context, mboxID imap.InternalMailboxID, uidValidity imap.UID) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = ? WHERE `%v` = ?",
		v1.MailboxesTableName,
		v1.MailboxesFieldUIDValidity,
		v1.MailboxesFieldID,
	)

	return utils.ExecQueryAndCheckUpdatedNotZero(ctx, w.qw, query, uidValidity, mboxID)
}

func (w writeOps) CreateMessages(ctx context.Context, reqs ...*db.CreateMessageReq) error {
	for _, chunk := range xslices.Chunk(reqs, db.ChunkLimit) {
		createMessageQuery := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`, `%v`, `%v`, `%v`, `%v`, `%v`) VALUES %v",
			v1.MessagesTableName,
			v1.MessagesFieldID,
			v1.MessagesFieldRemoteID,
			v1.MessagesFieldDate,
			v1.MessagesFieldSize,
			v1.MessagesFieldBody,
			v1.MessagesFieldBodyStructure,
			v1.MessagesFieldEnvelope,
			strings.Join(xslices.Repeat("(?,?,?,?,?,?,?)", len(chunk)), ","),
		)

		args := make([]any, 0, len(chunk)*6)
		flagArgs := make([]any, 0, len(chunk)*2)

		for _, req := range chunk {
			args = append(args,
				req.InternalID,
				req.Message.ID,
				req.Message.Date,
				req.LiteralSize,
				req.Body,
				req.Structure,
				req.Envelope)

			for _, f := range req.Message.Flags.ToSliceUnsorted() {
				flagArgs = append(flagArgs, req.InternalID, f)
			}
		}

		if _, err := utils.ExecQuery(ctx, w.qw, createMessageQuery, args...); err != nil {
			return err
		}

		for _, chunk := range xslices.Chunk(flagArgs, db.ChunkLimit) {
			createFlagsQuery := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES %v",
				v1.MessageFlagsTableName,
				v1.MessageFlagsFieldMessageID,
				v1.MessageFlagsFieldValue,
				strings.Join(xslices.Repeat("(?,?)", len(chunk)/2), ","),
			)

			if _, err := utils.ExecQuery(ctx, w.qw, createFlagsQuery, chunk...); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w writeOps) CreateMessageAndAddToMailbox(ctx context.Context, mbox imap.InternalMailboxID, req *db.CreateMessageReq) (imap.UID, imap.FlagSet, error) {
	createMessageQuery := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`, `%v`, `%v`, `%v`, `%v`, `%v`) VALUES (?,?,?,?,?,?,?)",
		v1.MessagesTableName,
		v1.MessagesFieldID,
		v1.MessagesFieldRemoteID,
		v1.MessagesFieldDate,
		v1.MessagesFieldSize,
		v1.MessagesFieldBody,
		v1.MessagesFieldBodyStructure,
		v1.MessagesFieldEnvelope,
	)

	if _, err := utils.ExecQuery(ctx, w.qw,
		createMessageQuery,
		req.InternalID,
		req.Message.ID,
		req.Message.Date,
		req.LiteralSize,
		req.Body,
		req.Structure,
		req.Envelope,
	); err != nil {
		return 0, imap.FlagSet{}, err
	}

	if req.Message.Flags.Len() != 0 {
		createFlagsQuery := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES %v",
			v1.MessageFlagsTableName,
			v1.MessageFlagsFieldMessageID,
			v1.MessageFlagsFieldValue,
			strings.Join(xslices.Repeat("(?, ?)", req.Message.Flags.Len()), ","),
		)

		args := make([]any, 0, req.Message.Flags.Len()*2)
		for _, f := range req.Message.Flags.ToSliceUnsorted() {
			args = append(args, req.InternalID, f)
		}

		if _, err := utils.ExecQuery(ctx, w.qw, createFlagsQuery, args...); err != nil {
			return 0, imap.FlagSet{}, err
		}
	}

	{
		query := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES (?,?)",
			v1.MessageToMailboxTableName,
			v1.MessageToMailboxFieldMessageID,
			v1.MessageToMailboxFieldMailboxID,
		)

		if _, err := utils.ExecQuery(ctx, w.qw, query, req.InternalID, mbox); err != nil {
			return 0, imap.FlagSet{}, err
		}
	}

	addToMboxQuery := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES (?,?) RETURNING `%v`",
		v1.MailboxMessageTableName(mbox),
		v1.MailboxMessagesFieldMessageID,
		v1.MailboxMessagesFieldMessageRemoteID,
		v1.MailboxMessagesFieldUID,
	)

	mboxUID, err := utils.MapQueryRow[imap.UID](ctx, w.qw, addToMboxQuery, req.InternalID, req.Message.ID)
	if err != nil {
		return 0, imap.FlagSet{}, err
	}

	flags := req.Message.Flags.Add(imap.FlagRecent)

	return mboxUID, flags, nil
}

func (w writeOps) MarkMessageAsDeleted(ctx context.Context, id imap.InternalMessageID) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = TRUE WHERE `%v` = ?",
		v1.MessagesTableName,
		v1.MessagesFieldDeleted,
		v1.MessagesFieldID,
	)

	_, err := utils.ExecQuery(ctx, w.qw, query, id)

	return err
}

func (w writeOps) MarkMessageAsDeletedAndAssignRandomRemoteID(ctx context.Context, id imap.InternalMessageID) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = TRUE, `%v` = ? WHERE `%v` = ?",
		v1.MessagesTableName,
		v1.MessagesFieldDeleted,
		v1.MessagesFieldRemoteID,
		v1.MessagesFieldID,
	)

	randomID := imap.MessageID(fmt.Sprintf("DELETED-%v", imap.NewInternalMessageID()))

	_, err := utils.ExecQuery(ctx, w.qw, query, randomID, id)

	return err
}

func (w writeOps) MarkMessageAsDeletedWithRemoteID(ctx context.Context, id imap.MessageID) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = TRUE WHERE `%v` = ?",
		v1.MessagesTableName,
		v1.MessagesFieldDeleted,
		v1.MessagesFieldRemoteID,
	)

	_, err := utils.ExecQuery(ctx, w.qw, query, id)

	return err
}

func (w writeOps) DeleteMessages(ctx context.Context, ids []imap.InternalMessageID) error {
	for _, chunk := range xslices.Chunk(ids, db.ChunkLimit) {
		query := fmt.Sprintf("DELETE FROM %v WHERE `%v` IN (%v)",
			v1.MessagesTableName,
			v1.MessagesFieldID,
			utils.GenSQLIn(len(chunk)),
		)

		if _, err := utils.ExecQuery(ctx, w.qw, query, utils.MapSliceToAny(chunk)...); err != nil {
			return err
		}
	}

	return nil
}

func (w writeOps) UpdateRemoteMessageID(ctx context.Context, internalID imap.InternalMessageID, remoteID imap.MessageID) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = ? WHERE `%v` = ?",
		v1.MessagesFieldID,
		v1.MessagesFieldRemoteID,
		v1.MessagesFieldID,
	)

	return utils.ExecQueryAndCheckUpdatedNotZero(ctx, w.qw, query, remoteID, internalID)
}

func (w writeOps) AddFlagToMessages(ctx context.Context, ids []imap.InternalMessageID, flag string) error {
	for _, chunk := range xslices.Chunk(ids, db.ChunkLimit) {
		query := fmt.Sprintf("INSERT OR IGNORE INTO %v (`%v`, `%v`) VALUES %v",
			v1.MessageFlagsTableName,
			v1.MessageFlagsFieldMessageID,
			v1.MessageFlagsFieldValue,
			strings.Join(xslices.Repeat("(?, ?)", len(chunk)), ","),
		)

		args := make([]any, 0, len(chunk)*2)

		for _, id := range chunk {
			args = append(args, id, flag)
		}

		if _, err := utils.ExecQuery(ctx, w.qw, query, args...); err != nil {
			return err
		}
	}

	return nil
}

func (w writeOps) RemoveFlagFromMessages(ctx context.Context, ids []imap.InternalMessageID, flag string) error {
	for _, chunk := range xslices.Chunk(ids, db.ChunkLimit) {
		query := fmt.Sprintf("DELETE FROM %v WHERE `%v` IN (%v) AND `%v` = ?",
			v1.MessageFlagsTableName,
			v1.MessageFlagsFieldMessageID,
			utils.GenSQLIn(len(chunk)),
			v1.MessageFlagsFieldValue,
		)

		if _, err := utils.ExecQuery(ctx, w.qw, query, append(utils.MapSliceToAny(chunk), flag)...); err != nil {
			return err
		}
	}

	return nil
}

func (w writeOps) SetFlagsOnMessages(ctx context.Context, ids []imap.InternalMessageID, flags imap.FlagSet) error {
	// GODT-2522: can silently ignore duplicates with INSERT OR IGNORE INTO ... if constraint exists.
	flagSlice := flags.ToSliceUnsorted()

	flagsSQLIn := utils.GenSQLIn(len(flagSlice))

	for _, chunk := range xslices.Chunk(ids, db.ChunkLimit/2) {
		deleteQuery := fmt.Sprintf("DELETE FROM %v WHERE `%v` IN (%v) AND `%v` NOT IN(%v)",
			v1.MessageFlagsTableName,
			v1.MessageFlagsFieldMessageID,
			utils.GenSQLIn(len(chunk)),
			v1.MessageFlagsFieldValue,
			flagsSQLIn,
		)

		insertQuery := fmt.Sprintf("INSERT OR IGNORE INTO %v (`%v`, `%v`) VALUES %v",
			v1.MessageFlagsTableName,
			v1.MessageFlagsFieldMessageID,
			v1.MessageFlagsFieldValue,
			strings.Join(xslices.Repeat("(?,?)", len(flagSlice)), ","),
		)

		deleteArgs := make([]any, 0, len(ids)+len(flagSlice))
		deleteArgs = append(deleteArgs, utils.MapSliceToAny(chunk)...)
		deleteArgs = append(deleteArgs, utils.MapSliceToAny(flagSlice)...)

		if _, err := utils.ExecQuery(ctx, w.qw, deleteQuery, deleteArgs...); err != nil {
			return err
		}

		insertArgs := make([]any, 0, len(flagSlice)*2*len(chunk))

		for _, id := range chunk {
			for _, flag := range flagSlice {
				insertArgs = append(insertArgs, id, flag)
			}
		}

		if _, err := utils.ExecQuery(ctx, w.qw, insertQuery, insertArgs...); err != nil {
			return err
		}
	}

	return nil
}

func (w writeOps) AddDeletedSubscription(ctx context.Context, mboxName string, mboxID imap.MailboxID) error {
	updateQuery := fmt.Sprintf("UPDATE %v SET `%v` = ? WHERE `%v` = ?",
		v1.DeletedSubscriptionsTableName,
		v1.DeletedSubscriptionsFieldRemoteID,
		v1.DeletedSubscriptionsFieldName,
	)

	count, err := utils.ExecQuery(ctx, w.qw, updateQuery, mboxID, mboxName)
	if err != nil {
		return err
	}

	if count == 0 {
		createQuery := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES (?, ?)",
			v1.DeletedSubscriptionsTableName,
			v1.DeletedSubscriptionsFieldName,
			v1.DeletedSubscriptionsFieldRemoteID,
		)

		if _, err := utils.ExecQuery(ctx, w.qw, createQuery, mboxName, mboxID); err != nil {
			return err
		}
	}

	return nil
}

func (w writeOps) RemoveDeletedSubscriptionWithName(ctx context.Context, mboxName string) (int, error) {
	query := fmt.Sprintf("DELETE FROM %v WHERE `%v` = ?",
		v1.DeletedSubscriptionsTableName,
		v1.DeletedSubscriptionsFieldName,
	)

	return utils.ExecQuery(ctx, w.qw, query, mboxName)
}

func (w writeOps) StoreConnectorSettings(ctx context.Context, settings string) error {
	query := fmt.Sprintf("UPDATE `%v` SET `%v`=? WHERE `%v`=?",
		v2.ConnectorSettingsTableName,
		v2.ConnectorSettingsFieldValue,
		v2.ConnectorSettingsFieldID,
	)

	_, err := utils.ExecQuery(ctx, w.qw, query, settings, v2.ConnectorSettingsDefaultID)

	return err
}

func (w writeOps) AddFlagsToAllMailboxes(ctx context.Context, flags ...string) error {
	flagsJoined := strings.Join(xslices.Map(flags, func(s string) string {
		return "('" + s + "')"
	}), ",")

	queryInsert := fmt.Sprintf(
		"INSERT OR IGNORE INTO %v (`%v`, `%v`) SELECT `%v`,`value` FROM %v CROSS JOIN (WITH T(value) AS (VALUES %v) SELECT * FROM T)",
		v1.MailboxFlagsTableName,
		v1.MailboxFlagsFieldMailboxID,
		v1.MailboxFlagsFieldValue,
		v1.MailboxesFieldID,
		v1.MailboxesTableName,
		flagsJoined,
	)

	_, err := utils.ExecQuery(ctx, w.qw, queryInsert)

	return err
}

func (w writeOps) AddPermFlagsToAllMailboxes(ctx context.Context, flags ...string) error {
	flagsJoined := strings.Join(xslices.Map(flags, func(s string) string {
		return "('" + s + "')"
	}), ",")

	queryInsert := fmt.Sprintf(
		"INSERT OR IGNORE INTO %v (`%v`, `%v`) SELECT `%v`,`value` FROM %v CROSS JOIN (WITH T(value) AS (VALUES %v) SELECT * FROM T)",
		v1.MailboxPermFlagsTableName,
		v1.MailboxPermFlagsFieldMailboxID,
		v1.MailboxPermFlagsFieldValue,
		v1.MailboxesFieldID,
		v1.MailboxesTableName,
		flagsJoined,
	)

	_, err := utils.ExecQuery(ctx, w.qw, queryInsert)

	return err
}
