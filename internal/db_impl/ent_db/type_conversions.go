package ent_db

import (
	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/internal/db_impl/ent_db/internal"
	"github.com/bradenaw/juniper/xslices"
)

func entMailboxFlagToDB(flag *internal.MailboxFlag) *db.MailboxFlag {
	if flag == nil {
		return nil
	}

	return &db.MailboxFlag{
		ID:    flag.ID,
		Value: flag.Value,
	}
}

func entMailboxPermFlagsToDB(flag *internal.MailboxPermFlag) *db.MailboxFlag {
	if flag == nil {
		return nil
	}

	return &db.MailboxFlag{
		ID:    flag.ID,
		Value: flag.Value,
	}
}

func entMailboxAttrToDB(flag *internal.MailboxAttr) *db.MailboxAttr {
	if flag == nil {
		return nil
	}

	return &db.MailboxAttr{
		ID:    flag.ID,
		Value: flag.Value,
	}
}

func entMBoxToDB(mbox *internal.Mailbox) *db.Mailbox {
	if mbox == nil {
		return nil
	}

	return &db.Mailbox{
		ID:             mbox.ID,
		RemoteID:       mbox.RemoteID,
		Name:           mbox.Name,
		UIDNext:        mbox.UIDNext,
		UIDValidity:    mbox.UIDValidity,
		Subscribed:     mbox.Subscribed,
		Flags:          xslices.Map(mbox.Edges.Flags, entMailboxFlagToDB),
		PermanentFlags: xslices.Map(mbox.Edges.PermanentFlags, entMailboxPermFlagsToDB),
		Attributes:     xslices.Map(mbox.Edges.Attributes, entMailboxAttrToDB),
	}
}

func entMessageFlagsToDB(flag *internal.MessageFlag) *db.MessageFlag {
	return &db.MessageFlag{
		ID:    flag.ID,
		Value: flag.Value,
	}
}

func entMessageUIDToDB(uid *internal.UID) *db.UID {
	return &db.UID{
		UID:     uid.UID,
		Deleted: uid.Deleted,
		Recent:  uid.Recent,
	}
}

func entMessageToDB(msg *internal.Message) *db.Message {
	if msg == nil {
		return nil
	}

	return &db.Message{
		ID:            msg.ID,
		RemoteID:      msg.RemoteID,
		Date:          msg.Date,
		Size:          msg.Size,
		Body:          msg.Body,
		BodyStructure: msg.BodyStructure,
		Envelope:      msg.Envelope,
		Deleted:       msg.Deleted,
		Flags:         xslices.Map(msg.Edges.Flags, entMessageFlagsToDB),
		UIDs:          xslices.Map(msg.Edges.UIDs, entMessageUIDToDB),
	}
}

func entSubscriptionToDB(s *internal.DeletedSubscription) *db.DeletedSubscription {
	if s == nil {
		return nil
	}

	return &db.DeletedSubscription{
		Name:     s.Name,
		RemoteID: s.RemoteID,
	}
}
