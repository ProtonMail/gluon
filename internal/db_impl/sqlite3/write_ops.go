package sqlite3

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3/utils"
	v0 "github.com/ProtonMail/gluon/internal/db_impl/sqlite3/v0"
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
	createMBoxQuery := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`, `%v`, `%v`, `%v`) VALUES (?,?,?,?,?) RETURNING `%v`",
		v0.MailboxesTableName,
		v0.MailboxesFieldRemoteID,
		v0.MailboxesFieldName,
		v0.MailboxesFieldUIDNext,
		v0.MailboxesFieldUIDValidity,
		v0.MailboxesFieldSubscribed,
		v0.MailboxesFieldID,
	)

	internalID, err := utils.MapQueryRow[imap.InternalMailboxID](ctx, w.qw, createMBoxQuery,
		mboxID,
		name,
		imap.UID(1),
		uidValidity,
		true,
	)
	if err != nil {
		return nil, err
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

	if err := createFlags(v0.MailboxFlagsTableName, v0.MailboxFlagsFieldMailboxID, v0.MailboxFlagsFieldValue, flags); err != nil {
		return nil, err
	}

	if err := createFlags(v0.MailboxPermFlagsTableName, v0.MailboxPermFlagsFieldMailboxID, v0.MailboxPermFlagsFieldValue, permFlags); err != nil {
		return nil, err
	}

	if err := createFlags(v0.MailboxAttrsTableName, v0.MailboxAttrsFieldMailboxID, v0.MailboxAttrsFieldValue, attrs); err != nil {
		return nil, err
	}

	return &db.Mailbox{
		ID:             internalID,
		RemoteID:       mboxID,
		Name:           name,
		UIDNext:        1,
		UIDValidity:    uidValidity,
		Subscribed:     true,
		Flags:          nil,
		PermanentFlags: nil,
		Attributes:     nil,
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
		v0.MailboxesTableName,
		v0.MailboxesFieldName,
		v0.MailboxesFieldRemoteID,
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

	query := fmt.Sprintf("DELETE FROM %v WHERE `%v` = ?",
		v0.MailboxesTableName,
		v0.MailboxesFieldRemoteID)

	_, err = utils.ExecQuery(ctx, w.qw, query, mboxID)

	return err
}

func (w writeOps) BumpMailboxUIDNext(ctx context.Context, mboxID imap.InternalMailboxID, count int) error {
	mboxUID, err := w.GetMailboxUID(ctx, mboxID)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("UPDATE %v SET `%v` = ? WHERE `%v` = ?",
		v0.MailboxesTableName,
		v0.MailboxesFieldUIDNext,
		v0.MailboxesFieldID,
	)

	return utils.ExecQueryAndCheckUpdatedNotZero(ctx, w.qw, query, mboxUID.Add(uint32(count)), mboxID)
}

func (w writeOps) AddMessagesToMailbox(
	ctx context.Context,
	mboxID imap.InternalMailboxID,
	messageIDs []imap.InternalMessageID,
) ([]db.UIDWithFlags, error) {
	if len(messageIDs) == 0 {
		return nil, nil
	}

	mboxUID, err := w.GetMailboxUID(ctx, mboxID)
	if err != nil {
		return nil, err
	}

	for chunkIdx, chunk := range xslices.Chunk(messageIDs, db.ChunkLimit/2) {
		query := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`, `%v`) VALUES %v",
			v0.UIDsTableName,
			v0.UIDsFieldMailboxID,
			v0.UIDsFieldUID,
			v0.UIDsFieldMessageID,
			strings.Join(xslices.Repeat("(?,?,?)", len(chunk)), ","),
		)

		args := make([]any, 0, 3*len(chunk))

		for idIdx, id := range chunk {
			nextUID := mboxUID.Add(uint32((chunkIdx * db.ChunkLimit / 2) + idIdx))
			args = append(args, mboxID, nextUID, id)
		}

		if _, err := utils.ExecQuery(ctx, w.qw, query, args...); err != nil {
			return nil, err
		}
	}

	if err := w.BumpMailboxUIDNext(ctx, mboxID, len(messageIDs)); err != nil {
		return nil, err
	}

	return w.GetMailboxMessageUIDsWithFlagsAfterAddOrUIDBump(ctx, mboxID, messageIDs)
}

func (w writeOps) BumpMailboxUIDsForMessage(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]db.UIDWithFlags, error) {
	mboxUID, err := w.GetMailboxUID(ctx, mboxID)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("UPDATE %v SET `%v` = ? WHERE `%v` = ? AND `%v`= ?",
		v0.UIDsTableName,
		v0.UIDsFieldUID,
		v0.UIDsFieldMessageID,
		v0.UIDsFieldMailboxID,
	)

	stmt, err := w.qw.PrepareStatement(ctx, query)
	if err != nil {
		return nil, err
	}

	defer utils.WrapStmtClose(stmt)

	for idx, id := range messageIDs {
		nextUID := mboxUID.Add(uint32(idx))

		if err := utils.ExecStmtAndCheckUpdatedNotZero(ctx, stmt, nextUID, id, mboxID); err != nil {
			return nil, err
		}
	}

	if err := w.BumpMailboxUIDNext(ctx, mboxID, len(messageIDs)); err != nil {
		return nil, err
	}

	return w.GetMailboxMessageUIDsWithFlagsAfterAddOrUIDBump(ctx, mboxID, messageIDs)
}

func (w writeOps) RemoveMessagesFromMailbox(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) error {
	for _, chunk := range xslices.Chunk(messageIDs, db.ChunkLimit) {
		query := fmt.Sprintf("DELETE FROM %v WHERE `%v` IN (%v) AND `%v` =?",
			v0.UIDsTableName,
			v0.UIDsFieldMessageID,
			utils.GenSQLIn(len(chunk)),
			v0.UIDsFieldMailboxID,
		)

		if _, err := utils.ExecQuery(ctx, w.qw, query, append(utils.MapSliceToAny(messageIDs), mboxID)...); err != nil {
			return err
		}
	}

	return nil
}

func (w writeOps) ClearRecentFlagInMailboxOnMessage(ctx context.Context, mboxID imap.InternalMailboxID, messageID imap.InternalMessageID) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = FALSE WHERE `%v` = ? AND `%v` =?",
		v0.UIDsTableName,
		v0.UIDsFieldRecent,
		v0.UIDsFieldMailboxID,
		v0.UIDsFieldMessageID,
	)

	return utils.ExecQueryAndCheckUpdatedNotZero(ctx, w.qw, query, mboxID, messageID)
}

func (w writeOps) ClearRecentFlagsInMailbox(ctx context.Context, mboxID imap.InternalMailboxID) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = FALSE WHERE `%v` = ?",
		v0.UIDsTableName,
		v0.UIDsFieldRecent,
		v0.UIDsFieldMailboxID,
	)

	_, err := utils.ExecQuery(ctx, w.qw, query, mboxID)

	return err
}

func (w writeOps) CreateMailboxIfNotExists(ctx context.Context, mbox imap.Mailbox, delimiter string, uidValidity imap.UID) error {
	_, err := w.GetOrCreateMailboxAlt(ctx, mbox, delimiter, uidValidity)

	return err
}

func (w writeOps) SetMailboxMessagesDeletedFlag(ctx context.Context, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID, deleted bool) error {
	for _, chunk := range xslices.Chunk(messageIDs, db.ChunkLimit) {
		query := fmt.Sprintf("UPDATE %v SET `%v` =? WHERE `%v` IN (%v) AND `%v` =? ",
			v0.UIDsTableName,
			v0.UIDsFieldDeleted,
			v0.UIDsFieldMessageID,
			utils.GenSQLIn(len(chunk)),
			v0.UIDsFieldMailboxID,
		)

		args := make([]any, 0, len(chunk)+2)
		args = append(args, deleted)
		args = append(args, utils.MapSliceToAny(chunk)...)
		args = append(args, mboxID)

		if _, err := utils.ExecQuery(ctx, w.qw, query, args...); err != nil {
			return err
		}
	}

	return nil
}

func (w writeOps) SetMailboxSubscribed(ctx context.Context, mboxID imap.InternalMailboxID, subscribed bool) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = ? WHERE `%v` = ?",
		v0.MailboxesTableName,
		v0.MailboxesFieldSubscribed,
		v0.MailboxesFieldID,
	)

	_, err := utils.ExecQuery(ctx, w.qw, query, subscribed, mboxID)

	return err
}

func (w writeOps) UpdateRemoteMailboxID(ctx context.Context, mboxID imap.InternalMailboxID, remoteID imap.MailboxID) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = ? WHERE `%v` = ?",
		v0.MailboxesTableName,
		v0.MailboxesFieldRemoteID,
		v0.MailboxesFieldID,
	)

	return utils.ExecQueryAndCheckUpdatedNotZero(ctx, w.qw, query, remoteID, mboxID)
}

func (w writeOps) SetMailboxUIDValidity(ctx context.Context, mboxID imap.InternalMailboxID, uidValidity imap.UID) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = ? WHERE `%v` = ?",
		v0.MailboxesTableName,
		v0.MailboxesFieldUIDValidity,
		v0.MailboxesFieldID,
	)

	return utils.ExecQueryAndCheckUpdatedNotZero(ctx, w.qw, query, uidValidity, mboxID)
}

func (w writeOps) CreateMessages(ctx context.Context, reqs ...*db.CreateMessageReq) ([]*db.Message, error) {
	result := make([]*db.Message, 0, len(reqs))

	for _, chunk := range xslices.Chunk(reqs, db.ChunkLimit) {
		createMessageQuery := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`, `%v`, `%v`, `%v`, `%v`, `%v`) VALUES %v",
			v0.MessagesTableName,
			v0.MessagesFieldID,
			v0.MessagesFieldRemoteID,
			v0.MessagesFieldDate,
			v0.MessagesFieldSize,
			v0.MessagesFieldBody,
			v0.MessagesFieldBodyStructure,
			v0.MessagesFieldEnvelope,
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

			result = append(result, &db.Message{
				ID:            req.InternalID,
				RemoteID:      req.Message.ID,
				Date:          req.Message.Date,
				Size:          req.LiteralSize,
				Body:          req.Body,
				BodyStructure: req.Structure,
				Envelope:      req.Envelope,
				Deleted:       false,
				Flags:         db.MessageFlagsFromFlagSet(req.Message.Flags),
				UIDs:          nil,
			})
		}

		if _, err := utils.ExecQuery(ctx, w.qw, createMessageQuery, args...); err != nil {
			return nil, err
		}

		for _, chunk := range xslices.Chunk(flagArgs, db.ChunkLimit) {
			createFlagsQuery := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES %v",
				v0.MessageFlagsTableName,
				v0.MessageFlagsFieldMessageID,
				v0.MessageFlagsFieldValue,
				strings.Join(xslices.Repeat("(?,?)", len(chunk)/2), ","),
			)

			if _, err := utils.ExecQuery(ctx, w.qw, createFlagsQuery, chunk...); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func (w writeOps) CreateMessageAndAddToMailbox(ctx context.Context, mbox imap.InternalMailboxID, req *db.CreateMessageReq) (imap.UID, imap.FlagSet, error) {
	mboxUID, err := w.GetMailboxUID(ctx, mbox)
	if err != nil {
		return 0, imap.FlagSet{}, err
	}

	createMessageQuery := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`, `%v`, `%v`, `%v`, `%v`, `%v`) VALUES (?,?,?,?,?,?,?)",
		v0.MessagesTableName,
		v0.MessagesFieldID,
		v0.MessagesFieldRemoteID,
		v0.MessagesFieldDate,
		v0.MessagesFieldSize,
		v0.MessagesFieldBody,
		v0.MessagesFieldBodyStructure,
		v0.MessagesFieldEnvelope,
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
			v0.MessageFlagsTableName,
			v0.MessageFlagsFieldMessageID,
			v0.MessageFlagsFieldValue,
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

	addToMboxQuery := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`, `%v`) VALUES (?,?,?)",
		v0.UIDsTableName,
		v0.UIDsFieldUID,
		v0.UIDsFieldMessageID,
		v0.UIDsFieldMailboxID,
	)

	if _, err := utils.ExecQuery(ctx, w.qw, addToMboxQuery, mboxUID, req.InternalID, mbox); err != nil {
		return 0, imap.FlagSet{}, err
	}

	if err := w.BumpMailboxUIDNext(ctx, mbox, 1); err != nil {
		return 0, imap.FlagSet{}, err
	}

	flags := req.Message.Flags.Add(imap.FlagRecent)

	return mboxUID, flags, nil
}

func (w writeOps) MarkMessageAsDeleted(ctx context.Context, id imap.InternalMessageID) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = TRUE WHERE `%v` = ?",
		v0.MessagesTableName,
		v0.MessagesFieldDeleted,
		v0.MessagesFieldID,
	)

	_, err := utils.ExecQuery(ctx, w.qw, query, id)

	return err
}

func (w writeOps) MarkMessageAsDeletedAndAssignRandomRemoteID(ctx context.Context, id imap.InternalMessageID) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = TRUE, `%v` = ? WHERE `%v` = ?",
		v0.MessagesTableName,
		v0.MessagesFieldDeleted,
		v0.MessagesFieldRemoteID,
		v0.MessagesFieldID,
	)

	randomID := imap.MessageID(fmt.Sprintf("DELETED-%v", imap.NewInternalMessageID()))

	_, err := utils.ExecQuery(ctx, w.qw, query, randomID, id)

	return err
}

func (w writeOps) MarkMessageAsDeletedWithRemoteID(ctx context.Context, id imap.MessageID) error {
	query := fmt.Sprintf("UPDATE %v SET `%v` = TRUE WHERE `%v` = ?",
		v0.MessagesTableName,
		v0.MessagesFieldDeleted,
		v0.MessagesFieldRemoteID,
	)

	_, err := utils.ExecQuery(ctx, w.qw, query, id)

	return err
}

func (w writeOps) DeleteMessages(ctx context.Context, ids []imap.InternalMessageID) error {
	for _, chunk := range xslices.Chunk(ids, db.ChunkLimit) {
		query := fmt.Sprintf("DELETE FROM %v WHERE `%v` IN (%v)",
			v0.MessagesTableName,
			v0.MessagesFieldID,
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
		v0.MessagesFieldID,
		v0.MessagesFieldRemoteID,
		v0.MessagesFieldID,
	)

	return utils.ExecQueryAndCheckUpdatedNotZero(ctx, w.qw, query, remoteID, internalID)
}

func (w writeOps) AddFlagToMessages(ctx context.Context, ids []imap.InternalMessageID, flag string) error {
	for _, chunk := range xslices.Chunk(ids, db.ChunkLimit) {
		query := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES %v",
			v0.MessageFlagsTableName,
			v0.MessageFlagsFieldMessageID,
			v0.MessageFlagsFieldValue,
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
			v0.MessageFlagsTableName,
			v0.MessageFlagsFieldMessageID,
			utils.GenSQLIn(len(chunk)),
			v0.MessageFlagsFieldValue,
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
			v0.MessageFlagsTableName,
			v0.MessageFlagsFieldMessageID,
			utils.GenSQLIn(len(chunk)),
			v0.MessageFlagsFieldValue,
			flagsSQLIn,
		)

		insertQuery := fmt.Sprintf("INSERT OR REPLACE INTO %v (`%v`, `%v`) VALUES %v",
			v0.MessageFlagsTableName,
			v0.MessageFlagsFieldMessageID,
			v0.MessageFlagsFieldValue,
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
		v0.DeletedSubscriptionsTableName,
		v0.DeletedSubscriptionsFieldRemoteID,
		v0.DeletedSubscriptionsFieldName,
	)

	count, err := utils.ExecQuery(ctx, w.qw, updateQuery, mboxID, mboxName)
	if err != nil {
		return err
	}

	if count == 0 {
		createQuery := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES (?, ?)",
			v0.DeletedSubscriptionsTableName,
			v0.DeletedSubscriptionsFieldName,
			v0.DeletedSubscriptionsFieldRemoteID,
		)

		if _, err := utils.ExecQuery(ctx, w.qw, createQuery, mboxName, mboxID); err != nil {
			return err
		}
	}

	return nil
}

func (w writeOps) RemoveDeletedSubscriptionWithName(ctx context.Context, mboxName string) (int, error) {
	query := fmt.Sprintf("DELETE FROM %v WHERE `%v` = ?",
		v0.DeletedSubscriptionsTableName,
		v0.DeletedSubscriptionsFieldName,
	)

	return utils.ExecQuery(ctx, w.qw, query, mboxName)
}
