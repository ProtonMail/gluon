package sqlite3

import (
	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3/utils"
)

func ScanMailbox(scanner utils.RowScanner) (*db.Mailbox, error) {
	mbox := new(db.Mailbox)

	if err := scanner.Scan(&mbox.ID, &mbox.RemoteID, &mbox.Name, &mbox.UIDValidity, &mbox.Subscribed); err != nil {
		return nil, err
	}

	return mbox, nil
}

func ScanMailboxWithAttr(scanner utils.RowScanner) (*db.MailboxWithAttr, error) {
	mbox := new(db.MailboxWithAttr)

	if err := scanner.Scan(&mbox.ID, &mbox.RemoteID, &mbox.Name, &mbox.UIDValidity, &mbox.Subscribed); err != nil {
		return nil, err
	}

	return mbox, nil
}

func ScanMessage(scanner utils.RowScanner) (*db.Message, error) {
	msg := new(db.Message)

	if err := scanner.Scan(&msg.ID, &msg.RemoteID, &msg.Date, &msg.Size, &msg.Body, &msg.BodyStructure, &msg.Envelope, &msg.Deleted); err != nil {
		return nil, err
	}

	return msg, nil
}

func ScanMessageWithFlags(scanner utils.RowScanner) (*db.MessageWithFlags, error) {
	msg := new(db.MessageWithFlags)

	if err := scanner.Scan(&msg.ID, &msg.RemoteID, &msg.Date, &msg.Size, &msg.Body, &msg.BodyStructure, &msg.Envelope, &msg.Deleted); err != nil {
		return nil, err
	}

	return msg, nil
}
