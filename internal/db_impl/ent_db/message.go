package ent_db

import (
	"context"
	"fmt"
	"github.com/ProtonMail/gluon/db"

	"entgo.io/ent/dialect/sql"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db_impl/ent_db/internal"
	"github.com/ProtonMail/gluon/internal/db_impl/ent_db/internal/mailbox"
	"github.com/ProtonMail/gluon/internal/db_impl/ent_db/internal/message"
	"github.com/ProtonMail/gluon/internal/db_impl/ent_db/internal/messageflag"
	"github.com/ProtonMail/gluon/internal/db_impl/ent_db/internal/uid"
	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/slices"
)

const ChunkLimit = db.ChunkLimit

func CreateMessages(ctx context.Context, tx *internal.Tx, reqs ...*db.CreateMessageReq) ([]*internal.Message, error) {
	flags := make(map[imap.InternalMessageID][]*internal.MessageFlag)

	for _, req := range reqs {
		builders := xslices.Map(req.Message.Flags.ToSlice(), func(flag string) *internal.MessageFlagCreate {
			return tx.MessageFlag.Create().SetValue(flag)
		})

		entFlags, err := tx.MessageFlag.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return nil, err
		}

		flags[req.InternalID] = entFlags
	}

	builders := xslices.Map(reqs, func(req *db.CreateMessageReq) *internal.MessageCreate {
		msgCreate := tx.Message.Create().
			SetID(req.InternalID).
			SetDate(req.Message.Date).
			SetBody(req.Body).
			SetBodyStructure(req.Structure).
			SetEnvelope(req.Envelope).
			SetSize(req.LiteralSize).
			AddFlags(flags[req.InternalID]...)

		if len(req.Message.ID) != 0 {
			msgCreate = msgCreate.SetRemoteID(req.Message.ID)
		}

		return msgCreate
	})

	messages := make([]*internal.Message, 0, len(builders))

	// Avoid too many SQL variables error.
	for _, chunk := range xslices.Chunk(builders, ChunkLimit) {
		msgs, err := tx.Message.CreateBulk(chunk...).Save(ctx)
		if err != nil {
			return nil, err
		}

		messages = append(messages, msgs...)
	}

	return messages, nil
}

func AddMessagesToMailbox(ctx context.Context, tx *internal.Tx, messageIDs []imap.InternalMessageID, mboxID imap.InternalMailboxID) ([]db.UIDWithFlags, error) {
	if len(messageIDs) == 0 {
		return nil, nil
	}

	messageUIDs := make(map[imap.InternalMessageID]imap.UID)

	mbox, err := tx.Mailbox.Query().Where(mailbox.ID(mboxID)).Select(mailbox.FieldUIDNext).Only(ctx)
	if err != nil {
		return nil, err
	}

	var builders []*internal.UIDCreate

	for idx, messageID := range messageIDs {
		nextUID := mbox.UIDNext.Add(uint32(idx))
		messageUIDs[messageID] = nextUID

		builders = append(builders, tx.UID.Create().
			SetMailboxID(mboxID).
			SetMessageID(messageID).
			SetUID(nextUID),
		)
	}

	// Avoid too many SQL variables error.
	for _, chunk := range xslices.Chunk(builders, ChunkLimit) {
		if m, err := tx.UID.CreateBulk(chunk...).Save(ctx); err != nil {
			return nil, err
		} else if len(m) == 0 {
			return nil, fmt.Errorf("no messages were added to the mailbox")
		}
	}

	if err := BumpMailboxUIDNext(ctx, tx, mbox, len(messageIDs)); err != nil {
		return nil, err
	}

	return GetMessageUIDsWithFlagsAfterAddOrUIDBump(ctx, tx.Client(), mboxID, messageIDs)
}

func CreateAndAddMessageToMailbox(ctx context.Context, tx *internal.Tx, mboxID imap.InternalMailboxID, req *db.CreateMessageReq) (imap.UID, imap.FlagSet, error) {
	mbox, err := tx.Mailbox.Query().Where(mailbox.ID(mboxID)).Select(mailbox.FieldID, mailbox.FieldUIDNext).Only(ctx)
	if err != nil {
		return 0, imap.FlagSet{}, err
	}

	var flags []*internal.MessageFlag

	{
		builders := xslices.Map(req.Message.Flags.ToSlice(), func(flag string) *internal.MessageFlagCreate {
			return tx.MessageFlag.Create().SetValue(flag)
		})

		entFlags, err := tx.MessageFlag.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return 0, imap.FlagSet{}, err
		}

		flags = entFlags
	}

	msgCreate := tx.Message.Create().
		SetID(req.InternalID).
		SetDate(req.Message.Date).
		SetBody(req.Body).
		SetBodyStructure(req.Structure).
		SetEnvelope(req.Envelope).
		SetSize(req.LiteralSize).
		AddFlags(flags...)

	if len(req.Message.ID) != 0 {
		msgCreate = msgCreate.SetRemoteID(req.Message.ID)
	}

	message, err := msgCreate.Save(ctx)
	if err != nil {
		return 0, imap.FlagSet{}, err
	}

	uid, err := tx.UID.Create().
		SetMailbox(mbox).
		SetMessage(message).
		SetUID(mbox.UIDNext).
		Save(ctx)
	if err != nil {
		return 0, imap.FlagSet{}, err
	}

	if err := BumpMailboxUIDNext(ctx, tx, mbox, 1); err != nil {
		return 0, imap.FlagSet{}, err
	}

	return uid.UID, NewFlagSet(uid, flags), err
}

func BumpMailboxUIDsForMessage(ctx context.Context, tx *internal.Tx, messageIDs []imap.InternalMessageID, mboxID imap.InternalMailboxID) ([]db.UIDWithFlags, error) {
	messageUIDs := make(map[imap.InternalMessageID]imap.UID)

	mbox, err := tx.Mailbox.Query().Where(mailbox.ID(mboxID)).Only(ctx)
	if err != nil {
		return nil, err
	}

	var builders []*internal.UIDUpdate

	for idx, messageID := range messageIDs {
		uidNext := mbox.UIDNext.Add(uint32(idx))
		messageUIDs[messageID] = uidNext

		builders = append(builders, tx.UID.Update().
			SetUID(uidNext).
			Where(uid.HasMessageWith(message.ID(messageID))),
		)
	}

	for _, builder := range builders {
		if _, err := builder.Save(ctx); err != nil {
			return nil, err
		}
	}

	if err := BumpMailboxUIDNext(ctx, tx, mbox, len(messageIDs)); err != nil {
		return nil, err
	}

	return GetMessageUIDsWithFlagsAfterAddOrUIDBump(ctx, tx.Client(), mboxID, messageIDs)
}

func RemoveMessagesFromMailbox(ctx context.Context, tx *internal.Tx, messageIDs []imap.InternalMessageID, mboxID imap.InternalMailboxID) error {
	// Avoid too many SQL variables error.
	for _, chunk := range xslices.Chunk(messageIDs, ChunkLimit) {
		if _, err := tx.UID.Delete().
			Where(func(s *sql.Selector) {
				s.Where(sql.And(sql.In(uid.MessageColumn, xslices.Map(chunk, func(t imap.InternalMessageID) interface{} {
					return interface{}(t)
				})...), sql.EQ(uid.MailboxColumn, mboxID)))
			}).
			Exec(ctx); err != nil {
			return err
		}
	}

	return nil
}

func MessageExistsWithRemoteID(ctx context.Context, client *internal.Client, messageID imap.MessageID) (bool, error) {
	count, err := client.Message.Query().Where(message.RemoteID(messageID)).Count(ctx)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func GetMessage(ctx context.Context, client *internal.Client, messageID imap.InternalMessageID) (*internal.Message, error) {
	return client.Message.Query().Where(message.ID(messageID)).Only(ctx)
}

func GetImportedMessageData(ctx context.Context, client *internal.Client, messageID imap.InternalMessageID) (*internal.Message, error) {
	return client.Message.Query().Where(message.ID(messageID)).WithFlags().Select(message.FieldDate).Only(ctx)
}

func GetMessageDateAndSize(ctx context.Context, client *internal.Client, messageID imap.InternalMessageID) (*internal.Message, error) {
	return client.Message.Query().Where(message.ID(messageID)).Select(message.FieldSize, message.FieldDate).Only(ctx)
}

func GetMessageMailboxIDs(ctx context.Context, client *internal.Client, messageID imap.InternalMessageID) ([]imap.InternalMailboxID, error) {
	type tmp struct {
		MBoxID imap.InternalMailboxID `json:"mailbox_ui_ds"`
	}

	var messageUIDs []tmp

	if err := client.UID.Query().Where(func(s *sql.Selector) {
		s.Where(sql.EQ(uid.MessageColumn, messageID))
	}).Select(uid.MailboxColumn).Scan(ctx, &messageUIDs); err != nil {
		return nil, err
	}

	return xslices.Map(messageUIDs, func(t tmp) imap.InternalMailboxID {
		return t.MBoxID
	}), nil
}

// GetMessageUIDsWithFlagsAfterAddOrUIDBump exploits a property of adding a message to or bumping the UIDs of existing message in mailbox. It can only be
// used if you can guarantee that the messageID list contains only IDs that have recently added or bumped in the mailbox.
func GetMessageUIDsWithFlagsAfterAddOrUIDBump(ctx context.Context, client *internal.Client, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) ([]db.UIDWithFlags, error) {
	result := make([]db.UIDWithFlags, 0, len(messageIDs))

	// Hav to split this in chunks as this can trigger too many SQL Variables.
	for _, chunk := range xslices.Chunk(messageIDs, ChunkLimit) {
		// We can just sort the by UID as every addition and bump will be a new UID in ascending order.
		if err := client.UID.Query().Where(func(s *sql.Selector) {
			msgTable := sql.Table(message.Table)
			flagTable := sql.Table(messageflag.Table)
			s.Join(msgTable).On(s.C(uid.MessageColumn), msgTable.C(message.FieldID))
			s.LeftJoin(flagTable).On(s.C(uid.MessageColumn), flagTable.C(messageflag.MessagesColumn))
			s.Where(sql.And(sql.EQ(uid.MailboxColumn, mboxID), sql.In(s.C(uid.MessageColumn), xslices.Map(chunk, func(id imap.InternalMessageID) interface{} {
				return id
			})...)))
			s.Select(msgTable.C(message.FieldRemoteID), sql.As(fmt.Sprintf("GROUP_CONCAT(%v)", flagTable.C(messageflag.FieldValue)), "flags"), s.C(uid.FieldRecent), s.C(uid.FieldDeleted), s.C(uid.FieldUID), s.C(uid.MessageColumn))
			s.GroupBy(s.C(uid.MessageColumn))
		}).Select().Scan(ctx, &result); err != nil {
			return nil, err
		}
	}

	slices.SortFunc(result, func(v1 db.UIDWithFlags, v2 db.UIDWithFlags) bool {
		return v1.UID < v2.UID
	})

	if len(result) == 0 {
		return nil, fmt.Errorf("result can never be null after UID bump")
	}

	return result, nil
}

// GetMessageFlags returns the flags of the given messages.
// It does not include per-mailbox flags (\Deleted, \Recent)!
func GetMessageFlags(ctx context.Context, client *internal.Client, messageIDs []imap.InternalMessageID) ([]db.MessageFlagSet, error) {
	result := make([]db.MessageFlagSet, 0, len(messageIDs))

	for _, chunk := range xslices.Chunk(messageIDs, ChunkLimit) {
		chunkMessages, err := client.Message.Query().
			Where(message.IDIn(chunk...)).
			WithFlags(func(query *internal.MessageFlagQuery) {
				query.Select(messageflag.FieldValue)
			}).
			Select(message.FieldID, message.FieldRemoteID).
			All(ctx)
		if err != nil {
			return nil, err
		}

		for _, msg := range chunkMessages {
			mfs := db.MessageFlagSet{
				ID:       msg.ID,
				RemoteID: msg.RemoteID,
				FlagSet:  imap.NewFlagSetWithCapacity(len(msg.Edges.Flags)),
			}

			for _, v := range msg.Edges.Flags {
				mfs.FlagSet.AddToSelf(v.Value)
			}

			result = append(result, mfs)
		}
	}

	return result, nil
}

func GetMessageDeleted(ctx context.Context, client *internal.Client, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) (map[imap.InternalMessageID]bool, error) {
	type tmp struct {
		MsgID   imap.InternalMessageID `json:"uid_message"`
		Deleted bool                   `json:"deleted"`
	}

	var result []tmp

	for _, chunk := range xslices.Chunk(messageIDs, ChunkLimit) {
		if err := client.UID.Query().Where(
			func(s *sql.Selector) {
				s.Where(sql.And(sql.In(uid.MessageColumn, xslices.Map(chunk, func(t imap.InternalMessageID) interface{} {
					return interface{}(t)
				})...), sql.EQ(uid.MailboxColumn, mboxID)))
			}).
			Select(uid.MessageColumn, uid.FieldDeleted).
			Scan(ctx, &result); err != nil {
			return nil, err
		}
	}

	res := make(map[imap.InternalMessageID]bool)

	for _, r := range result {
		res[r.MsgID] = r.Deleted
	}

	return res, nil
}

func AddMessageFlag(ctx context.Context, tx *internal.Tx, messageIDs []imap.InternalMessageID, addFlag string) error {
	for _, chunk := range xslices.Chunk(messageIDs, ChunkLimit) {
		builders := xslices.Map(chunk, func(imap.InternalMessageID) *internal.MessageFlagCreate {
			return tx.MessageFlag.Create().SetValue(addFlag)
		})

		flags, err := tx.MessageFlag.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return err
		}

		for idx, msg := range chunk {
			if _, err := tx.Message.Update().Where(message.ID(msg)).AddFlags(flags[idx]).Save(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

func RemoveMessageFlag(ctx context.Context, tx *internal.Tx, messageIDs []imap.InternalMessageID, remFlag string) error {
	remFlagSet := imap.NewFlagSet(remFlag)

	for _, chunk := range xslices.Chunk(messageIDs, ChunkLimit) {
		messages, err := tx.Message.Query().
			Where(message.IDIn(chunk...)).
			Select(message.FieldID).
			WithFlags().
			All(ctx)
		if err != nil {
			return err
		}

		flags := xslices.Map(messages, func(message *internal.Message) *internal.MessageFlag {
			return message.Edges.Flags[xslices.IndexFunc(message.Edges.Flags, func(flag *internal.MessageFlag) bool {
				return remFlagSet.Contains(flag.Value)
			})]
		})

		for idx, message := range messages {
			if _, err := message.Update().RemoveFlags(flags[idx]).Save(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

func SetMessageFlags(ctx context.Context, tx *internal.Tx, messageIDs []imap.InternalMessageID, setFlags imap.FlagSet) error {
	for _, chunk := range xslices.Chunk(messageIDs, ChunkLimit) {
		messages, err := tx.Message.Query().
			Where(message.IDIn(chunk...)).
			Select(message.FieldID).
			WithFlags().
			All(ctx)
		if err != nil {
			return err
		}

		for _, message := range messages {
			curFlagSet := imap.NewFlagSetFromSlice(xslices.Map(message.Edges.Flags, func(flag *internal.MessageFlag) string {
				return flag.Value
			}))

			addFlags := xslices.Filter(setFlags.ToSlice(), func(flag string) bool {
				return !curFlagSet.Contains(flag)
			})

			builders := xslices.Map(addFlags, func(flag string) *internal.MessageFlagCreate {
				return tx.MessageFlag.Create().SetValue(flag)
			})

			addEntFlags, err := tx.MessageFlag.CreateBulk(builders...).Save(ctx)
			if err != nil {
				return err
			}

			remEntFlags := xslices.Filter(message.Edges.Flags, func(flag *internal.MessageFlag) bool {
				return !setFlags.Contains(flag.Value)
			})

			if _, err := message.Update().
				AddFlags(addEntFlags...).
				RemoveFlags(remEntFlags...).
				Save(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

func SetDeletedFlag(ctx context.Context, tx *internal.Tx, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID, deleted bool) error {
	for _, chunk := range xslices.Chunk(messageIDs, ChunkLimit) {
		if _, err := tx.UID.Update().
			Where(
				func(s *sql.Selector) {
					s.Where(sql.And(sql.In(uid.MessageColumn, xslices.Map(chunk, func(t imap.InternalMessageID) interface{} {
						return interface{}(t)
					})...), sql.EQ(uid.MailboxColumn, mboxID)))
				}).
			SetDeleted(deleted).
			Save(ctx); err != nil {
			return err
		}
	}

	return nil
}

func ClearRecentFlag(ctx context.Context, tx *internal.Tx, mboxID imap.InternalMailboxID, messageID imap.InternalMessageID) error {
	if _, err := tx.UID.Update().
		Where(
			func(s *sql.Selector) {
				s.Where(sql.And(sql.EQ(uid.MessageColumn, messageID), sql.EQ(uid.MailboxColumn, mboxID), sql.EQ(uid.FieldRecent, true)))
			}).
		SetRecent(false).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func ClearRecentFlags(ctx context.Context, tx *internal.Tx, mboxID imap.InternalMailboxID) error {
	if _, err := tx.UID.Update().
		Where(
			func(s *sql.Selector) {
				s.Where(sql.And(sql.EQ(uid.MailboxColumn, mboxID), sql.EQ(uid.FieldRecent, true)))
			}).
		SetRecent(false).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func UpdateRemoteMessageID(ctx context.Context, tx *internal.Tx, internalID imap.InternalMessageID, remoteID imap.MessageID) error {
	if _, err := tx.Message.Update().
		Where(message.ID(internalID)).
		SetRemoteID(remoteID).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func MarkMessageAsDeleted(ctx context.Context, tx *internal.Tx, messageID imap.InternalMessageID) error {
	if _, err := tx.Message.Update().Where(message.ID(messageID)).SetDeleted(true).Save(ctx); err != nil {
		return err
	}

	return nil
}

func MarkMessageAsDeletedAndAssignRandomRemoteID(ctx context.Context, tx *internal.Tx, messageID imap.InternalMessageID) error {
	randomID := imap.MessageID(fmt.Sprintf("DELETED-%v", imap.NewInternalMessageID()))
	if _, err := tx.Message.Update().Where(message.ID(messageID)).SetDeleted(true).SetRemoteID(randomID).Save(ctx); err != nil {
		return err
	}

	return nil
}

func MarkMessageAsDeletedWithRemoteID(ctx context.Context, tx *internal.Tx, messageID imap.MessageID) error {
	if _, err := tx.Message.Update().Where(message.RemoteID(messageID)).SetDeleted(true).Save(ctx); err != nil {
		return err
	}

	return nil
}

func DeleteMessages(ctx context.Context, tx *internal.Tx, messageIDs ...imap.InternalMessageID) error {
	for _, chunk := range xslices.Chunk(messageIDs, ChunkLimit) {
		if _, err := tx.Message.Delete().Where(message.IDIn(chunk...)).Exec(ctx); err != nil {
			return err
		}
	}

	return nil
}

func GetMessageIDsMarkedDeleted(ctx context.Context, client *internal.Client) ([]imap.InternalMessageID, error) {
	messages, err := client.Message.Query().Where(message.Deleted(true)).Select(message.FieldID).All(ctx)
	if err != nil {
		return nil, err
	}

	return xslices.Map(messages, func(t *internal.Message) imap.InternalMessageID {
		return t.ID
	}), nil
}

func HasMessageWithID(ctx context.Context, client *internal.Client, id imap.InternalMessageID) (bool, error) {
	return client.Message.Query().Where(message.ID(id)).Exist(ctx)
}

func HasMessageWithRemoteID(ctx context.Context, client *internal.Client, id imap.MessageID) (bool, error) {
	_, err := client.Message.Query().Where(message.RemoteID(id)).Select(message.FieldRemoteID).Only(ctx)
	if err != nil {
		if internal.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	// For whatever weird reason, this stopped working all together. No code changes were made, but reflection started
	// failing.
	// Now we get error="internal: check existence: sql/scan: missing struct field for column: id (id)".
	//return client.Message.Query().Where(message.RemoteID(id)).Exist(ctx)

	return true, nil
}

func GetMessageIDFromRemoteID(ctx context.Context, client *internal.Client, id imap.MessageID) (imap.InternalMessageID, error) {
	message, err := client.Message.Query().Where(message.RemoteID(id)).Select(message.FieldID).Only(ctx)
	if err != nil {
		return imap.InternalMessageID{}, err
	}

	return message.ID, nil
}

func GetMessageWithIDWithDeletedFlag(ctx context.Context, client *internal.Client, id imap.InternalMessageID) (*internal.Message, error) {
	message, err := client.Message.Query().Where(message.ID(id)).Select(message.FieldID, message.FieldDeleted).Only(ctx)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func GetMessageFromRemoteIDWithDeletedFlag(ctx context.Context, client *internal.Client, id imap.MessageID) (*internal.Message, error) {
	message, err := client.Message.Query().Where(message.RemoteID(id)).Select(message.FieldID, message.FieldDeleted).Only(ctx)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func GetMessageRemoteIDFromID(ctx context.Context, client *internal.Client, id imap.InternalMessageID) (imap.MessageID, error) {
	message, err := client.Message.Query().Where(message.ID(id)).Select(message.FieldRemoteID).Only(ctx)
	if err != nil {
		return "", err
	}

	return message.RemoteID, nil
}

func NewFlagSet(msgUID *internal.UID, flags []*internal.MessageFlag) imap.FlagSet {
	flagSet := imap.NewFlagSetFromSlice(xslices.Map(flags, func(flag *internal.MessageFlag) string {
		return flag.Value
	}))

	if msgUID.Deleted {
		flagSet.AddToSelf(imap.FlagDeleted)
	}

	if msgUID.Recent {
		flagSet.AddToSelf(imap.FlagRecent)
	}

	return flagSet
}

func GetHighestMessageID(ctx context.Context, client *internal.Client) (imap.InternalMessageID, error) {
	message, err := client.Message.Query().Select(message.FieldID).Order(internal.Desc(message.FieldID)).Limit(1).All(ctx)
	if err != nil {
		return imap.InternalMessageID{}, err
	}

	if len(message) == 0 {
		return imap.InternalMessageID{}, nil
	}

	return message[0].ID, nil
}

func GetAllMessagesIDsAsMap(ctx context.Context, client *internal.Client) (map[imap.InternalMessageID]struct{}, error) {
	messages, err := client.Message.Query().Select(message.FieldID).All(ctx)
	if err != nil {
		return nil, err
	}

	idMap := make(map[imap.InternalMessageID]struct{}, len(messages))
	for _, v := range messages {
		idMap[v.ID] = struct{}{}
	}

	return idMap, nil
}
