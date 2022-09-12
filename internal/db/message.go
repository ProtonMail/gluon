package db

import (
	"context"
	"entgo.io/ent/dialect/sql"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/db/ent/mailbox"
	"github.com/ProtonMail/gluon/internal/db/ent/message"
	"github.com/ProtonMail/gluon/internal/db/ent/messageflag"
	"github.com/ProtonMail/gluon/internal/db/ent/uid"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/bradenaw/juniper/xslices"
)

type CreateMessageReq struct {
	Message    imap.Message
	InternalID imap.InternalMessageID
	Literal    []byte
	Body       string
	Structure  string
	Envelope   string
}

func CreateMessages(ctx context.Context, tx *ent.Tx, reqs ...*CreateMessageReq) ([]*ent.Message, error) {
	flags := make(map[imap.InternalMessageID][]*ent.MessageFlag)

	for _, req := range reqs {
		builders := xslices.Map(req.Message.Flags.ToSlice(), func(flag string) *ent.MessageFlagCreate {
			return tx.MessageFlag.Create().SetValue(flag)
		})

		entFlags, err := tx.MessageFlag.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return nil, err
		}

		flags[req.InternalID] = entFlags
	}

	return tx.Message.CreateBulk(xslices.Map(reqs, func(req *CreateMessageReq) *ent.MessageCreate {
		msgCreate := tx.Message.Create().
			SetID(req.InternalID).
			SetDate(req.Message.Date).
			SetBody(req.Body).
			SetBodyStructure(req.Structure).
			SetEnvelope(req.Envelope).
			SetSize(len(req.Literal)).
			AddFlags(flags[req.InternalID]...)

		if len(req.Message.ID) != 0 {
			msgCreate = msgCreate.SetRemoteID(req.Message.ID)
		}

		return msgCreate
	})...).Save(ctx)
}

func AddMessagesToMailbox(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, mboxID imap.InternalMailboxID) (map[imap.InternalMessageID]*ent.UID, error) {
	messageUIDs := make(map[imap.InternalMessageID]imap.UID)

	mbox, err := tx.Mailbox.Query().Where(mailbox.ID(mboxID)).Select(mailbox.FieldUIDNext).Only(ctx)
	if err != nil {
		return nil, err
	}

	var builders []*ent.UIDCreate

	for idx, messageID := range messageIDs {
		nextUID := mbox.UIDNext.Add(uint32(idx))
		messageUIDs[messageID] = nextUID

		builders = append(builders, tx.UID.Create().
			SetMailboxID(mboxID).
			SetMessageID(messageID).
			SetUID(nextUID),
		)
	}

	if _, err := tx.UID.CreateBulk(builders...).Save(ctx); err != nil {
		return nil, err
	}

	if err := BumpMailboxUIDNext(ctx, tx, mbox, len(messageIDs)); err != nil {
		return nil, err
	}

	return GetMessageUIDsWithFlags(ctx, tx.Client(), mboxID, messageIDs)
}

func CreateAndAddMessageToMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, req *CreateMessageReq) (imap.UID, imap.FlagSet, error) {
	mbox, err := tx.Mailbox.Query().Where(mailbox.ID(mboxID)).Select(mailbox.FieldID, mailbox.FieldUIDNext).Only(ctx)
	if err != nil {
		return 0, imap.FlagSet{}, err
	}

	var flags []*ent.MessageFlag

	{
		builders := xslices.Map(req.Message.Flags.ToSlice(), func(flag string) *ent.MessageFlagCreate {
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
		SetSize(len(req.Literal)).
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

type CreateAndAddMessagesResult struct {
	UID       imap.UID
	Flags     imap.FlagSet
	MessageID ids.MessageIDPair
}

func CreateAndAddMessagesToMailbox(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, requests []*CreateMessageReq) ([]CreateAndAddMessagesResult, error) {
	mbox, err := tx.Mailbox.Query().Where(mailbox.ID(mboxID)).Select(mailbox.FieldID, mailbox.FieldUIDNext).Only(ctx)
	if err != nil {
		return nil, err
	}

	msgBuilders := make([]*ent.MessageCreate, 0, len(requests))
	flags := make([][]*ent.MessageFlag, 0, len(requests))

	for _, request := range requests {
		builders := xslices.Map(request.Message.Flags.ToSlice(), func(flag string) *ent.MessageFlagCreate {
			return tx.MessageFlag.Create().SetValue(flag)
		})

		entFlags, err := tx.MessageFlag.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return nil, err
		}

		flags = append(flags, entFlags)

		msgCreate := tx.Message.Create().
			SetID(request.InternalID).
			SetDate(request.Message.Date).
			SetBody(request.Body).
			SetBodyStructure(request.Structure).
			SetEnvelope(request.Envelope).
			SetSize(len(request.Literal)).
			AddFlags(entFlags...)

		if len(request.Message.ID) != 0 {
			msgCreate = msgCreate.SetRemoteID(request.Message.ID)
		}

		msgBuilders = append(msgBuilders, msgCreate)
	}

	messages, err := tx.Message.CreateBulk(msgBuilders...).Save(ctx)
	if err != nil {
		return nil, err
	}

	uidBuilders := make([]*ent.UIDCreate, 0, len(requests))

	for i, message := range messages {
		uidBuilders = append(uidBuilders, tx.UID.Create().
			SetMailboxID(mbox.ID).
			SetMessageID(message.ID).
			SetUID(mbox.UIDNext.Add(uint32(i))),
		)
	}

	uids, err := tx.UID.CreateBulk(uidBuilders...).Save(ctx)
	if err != nil {
		return nil, err
	}

	if err := BumpMailboxUIDNext(ctx, tx, mbox, len(requests)); err != nil {
		return nil, err
	}

	result := make([]CreateAndAddMessagesResult, 0, len(requests))

	for i := 0; i < len(requests); i++ {
		if uids[i].UID != mbox.UIDNext.Add(uint32(i)) {
			panic("Invalid UID ")
		}

		result = append(result, CreateAndAddMessagesResult{
			MessageID: ids.MessageIDPair{
				InternalID: messages[i].ID,
				RemoteID:   messages[i].RemoteID,
			},
			UID:   uids[i].UID,
			Flags: NewFlagSet(uids[i], flags[i]),
		})
	}

	return result, err
}

func BumpMailboxUIDsForMessage(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, mboxID imap.InternalMailboxID) (map[imap.InternalMessageID]*ent.UID, error) {
	messageUIDs := make(map[imap.InternalMessageID]imap.UID)

	mbox, err := tx.Mailbox.Query().Where(mailbox.ID(mboxID)).Only(ctx)
	if err != nil {
		return nil, err
	}

	var builders []*ent.UIDUpdate

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

	return GetMessageUIDsWithFlags(ctx, tx.Client(), mboxID, messageIDs)
}

func RemoveMessagesFromMailbox(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, mboxID imap.InternalMailboxID) error {
	if _, err := tx.UID.Delete().
		Where(func(s *sql.Selector) {
			s.Where(sql.And(sql.In(uid.MessageColumn, xslices.Map(messageIDs, func(t imap.InternalMessageID) interface{} {
				return interface{}(t)
			})...), sql.EQ(uid.MailboxColumn, mboxID)))
		}).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func MessageExistsWithRemoteID(ctx context.Context, client *ent.Client, messageID imap.MessageID) (bool, error) {
	count, err := client.Message.Query().Where(message.RemoteID(messageID)).Count(ctx)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func GetMessage(ctx context.Context, client *ent.Client, messageID imap.InternalMessageID) (*ent.Message, error) {
	return client.Message.Query().Where(message.ID(messageID)).Only(ctx)
}

func GetMessageMailboxIDs(ctx context.Context, client *ent.Client, messageID imap.InternalMessageID) ([]imap.InternalMailboxID, error) {
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

func GetMessageUIDsWithFlags(ctx context.Context, client *ent.Client, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) (map[imap.InternalMessageID]*ent.UID, error) {
	messageUIDs, err := client.UID.Query().
		Where(
			uid.HasMailboxWith(mailbox.ID(mboxID)),
			uid.HasMessageWith(message.IDIn(messageIDs...)),
		).
		WithMessage(func(query *ent.MessageQuery) {
			query.Select(message.FieldID, message.FieldRemoteID)
			query.WithFlags(func(query *ent.MessageFlagQuery) {
				query.Select(messageflag.FieldValue)
			})
		}).
		All(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[imap.InternalMessageID]*ent.UID)

	for _, messageUID := range messageUIDs {
		res[messageUID.Edges.Message.ID] = messageUID
	}

	return res, nil
}

// GetMessageFlags returns the flags of the given messages.
// It does not include per-mailbox flags (\Deleted, \Recent)!
func GetMessageFlags(ctx context.Context, client *ent.Client, messageIDs []imap.InternalMessageID) (map[imap.InternalMessageID]imap.FlagSet, error) {
	messages, err := client.Message.Query().
		Where(message.IDIn(messageIDs...)).
		WithFlags(func(query *ent.MessageFlagQuery) {
			query.Select(messageflag.FieldValue)
		}).
		Select(message.FieldID).
		All(ctx)
	if err != nil {
		return nil, err
	}

	curFlags := make(map[imap.InternalMessageID]imap.FlagSet)

	for _, message := range messages {
		curFlags[message.ID] = imap.NewFlagSetFromSlice(xslices.Map(message.Edges.Flags, func(flag *ent.MessageFlag) string {
			return flag.Value
		}))
	}

	return curFlags, nil
}

func GetMessageDeleted(ctx context.Context, client *ent.Client, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID) (map[imap.InternalMessageID]bool, error) {
	type tmp struct {
		MsgID   imap.InternalMessageID `json:"uid_message"`
		Deleted bool                   `json:"deleted"`
	}

	var result []tmp

	if err := client.UID.Query().Where(
		func(s *sql.Selector) {
			s.Where(sql.And(sql.In(uid.MessageColumn, xslices.Map(messageIDs, func(t imap.InternalMessageID) interface{} {
				return interface{}(t)
			})...), sql.EQ(uid.MailboxColumn, mboxID)))
		}).
		Select(uid.MessageColumn, uid.FieldDeleted).
		Scan(ctx, &result); err != nil {
		return nil, err
	}

	res := make(map[imap.InternalMessageID]bool)

	for _, r := range result {
		res[r.MsgID] = r.Deleted
	}

	return res, nil
}

func AddMessageFlag(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, addFlag string) error {
	builders := xslices.Map(messageIDs, func(imap.InternalMessageID) *ent.MessageFlagCreate {
		return tx.MessageFlag.Create().SetValue(addFlag)
	})

	flags, err := tx.MessageFlag.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return err
	}

	for idx, msg := range messageIDs {
		if _, err := tx.Message.Update().Where(message.ID(msg)).AddFlags(flags[idx]).Save(ctx); err != nil {
			return err
		}
	}

	return nil
}

func RemoveMessageFlag(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, remFlag string) error {
	messages, err := tx.Message.Query().
		Where(message.IDIn(messageIDs...)).
		Select(message.FieldID).
		WithFlags().
		All(ctx)
	if err != nil {
		return err
	}

	flags := xslices.Map(messages, func(message *ent.Message) *ent.MessageFlag {
		return message.Edges.Flags[xslices.IndexFunc(message.Edges.Flags, func(flag *ent.MessageFlag) bool {
			return imap.NewFlagSet(remFlag).Contains(flag.Value)
		})]
	})

	for idx, message := range messages {
		if _, err := message.Update().RemoveFlags(flags[idx]).Save(ctx); err != nil {
			return err
		}
	}

	return nil
}

func SetMessageFlags(ctx context.Context, tx *ent.Tx, messageIDs []imap.InternalMessageID, setFlags imap.FlagSet) error {
	messages, err := tx.Message.Query().
		Where(message.IDIn(messageIDs...)).
		Select(message.FieldID).
		WithFlags().
		All(ctx)
	if err != nil {
		return err
	}

	for _, message := range messages {
		curFlagSet := imap.NewFlagSetFromSlice(xslices.Map(message.Edges.Flags, func(flag *ent.MessageFlag) string {
			return flag.Value
		}))

		addFlags := xslices.Filter(setFlags.ToSlice(), func(flag string) bool {
			return !curFlagSet.Contains(flag)
		})

		builders := xslices.Map(addFlags, func(flag string) *ent.MessageFlagCreate {
			return tx.MessageFlag.Create().SetValue(flag)
		})

		addEntFlags, err := tx.MessageFlag.CreateBulk(builders...).Save(ctx)
		if err != nil {
			return err
		}

		remEntFlags := xslices.Filter(message.Edges.Flags, func(flag *ent.MessageFlag) bool {
			return !setFlags.Contains(flag.Value)
		})

		if _, err := message.Update().
			AddFlags(addEntFlags...).
			RemoveFlags(remEntFlags...).
			Save(ctx); err != nil {
			return err
		}
	}

	return nil
}

func SetDeletedFlag(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, messageIDs []imap.InternalMessageID, deleted bool) error {
	if _, err := tx.UID.Update().
		Where(
			func(s *sql.Selector) {
				s.Where(sql.And(sql.In(uid.MessageColumn, xslices.Map(messageIDs, func(t imap.InternalMessageID) interface{} {
					return interface{}(t)
				})...), sql.EQ(uid.MailboxColumn, mboxID)))
			}).
		SetDeleted(deleted).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func ClearRecentFlag(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID, messageID imap.InternalMessageID) error {
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

func ClearRecentFlags(ctx context.Context, tx *ent.Tx, mboxID imap.InternalMailboxID) error {
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

func UpdateRemoteMessageID(ctx context.Context, tx *ent.Tx, internalID imap.InternalMessageID, remoteID imap.MessageID) error {
	if _, err := tx.Message.Update().
		Where(message.ID(internalID)).
		SetRemoteID(remoteID).
		Save(ctx); err != nil {
		return err
	}

	return nil
}

func MarkMessageAsDeletedWithRemoteID(ctx context.Context, tx *ent.Tx, messageID imap.MessageID) error {
	if _, err := tx.Message.Update().Where(message.RemoteID(messageID)).SetDeleted(true).Save(ctx); err != nil {
		return err
	}

	return nil
}

func DeleteMessages(ctx context.Context, tx *ent.Tx, messageIDs ...imap.InternalMessageID) error {
	if _, err := tx.Message.Delete().Where(message.IDIn(messageIDs...)).Exec(ctx); err != nil {
		return err
	}

	return nil
}

func GetMessageIDsMarkedDeleted(ctx context.Context, client *ent.Client) ([]imap.InternalMessageID, error) {
	messages, err := client.Message.Query().Where(message.Deleted(true)).Select(message.FieldID).All(ctx)
	if err != nil {
		return nil, err
	}

	return xslices.Map(messages, func(t *ent.Message) imap.InternalMessageID {
		return t.ID
	}), nil
}

func HasMessageWithID(ctx context.Context, client *ent.Client, id imap.InternalMessageID) (bool, error) {
	count, err := client.Message.Query().Where(message.ID(id)).Count(ctx)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func GetMessageIDFromRemoteID(ctx context.Context, client *ent.Client, id imap.MessageID) (imap.InternalMessageID, error) {
	message, err := client.Message.Query().Where(message.RemoteID(id)).Select(message.FieldID).Only(ctx)
	if err != nil {
		return "", err
	}

	return message.ID, nil
}

func NewFlagSet(msgUID *ent.UID, flags []*ent.MessageFlag) imap.FlagSet {
	flagSet := imap.NewFlagSetFromSlice(xslices.Map(flags, func(flag *ent.MessageFlag) string {
		return flag.Value
	}))

	if msgUID.Deleted {
		flagSet = flagSet.Add(imap.FlagDeleted)
	}

	if msgUID.Recent {
		flagSet = flagSet.Add(imap.FlagRecent)
	}

	return flagSet
}
