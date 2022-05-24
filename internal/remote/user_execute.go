package remote

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gluon/imap"
)

func (user *User) execute(ctx context.Context, op operation) error {
	switch op := op.(type) {
	case *OpMailboxCreate:
		return user.executeMailboxCreate(ctx, op)

	case *OpMailboxDelete:
		return user.executeMailboxDelete(ctx, op)

	case *OpMailboxUpdate:
		return user.executeMailboxUpdate(ctx, op)

	case *OpMessageCreate:
		return user.executeMessageCreate(ctx, op)

	case *OpMessageAdd:
		return user.executeMessageAdd(ctx, op)

	case *OpMessageRemove:
		return user.executeMessageRemove(ctx, op)

	case *OpMessageSeen:
		return user.executeMessageSeen(ctx, op)

	case *OpMessageFlagged:
		return user.executeMessageFlagged(ctx, op)

	default:
		panic(fmt.Sprintf("bad operation: %v", op))
	}
}

func (user *User) executeMailboxCreate(ctx context.Context, op *OpMailboxCreate) error {
	mbox, err := user.conn.CreateLabel(ctx, op.Name)
	if err != nil {
		return err
	}

	user.setMailboxID(op.TempID, mbox.ID)

	user.send(imap.NewMailboxIDChanged(op.TempID, mbox.ID), true)

	return nil
}

func (user *User) executeMailboxDelete(ctx context.Context, op *OpMailboxDelete) error {
	return user.conn.DeleteLabel(ctx, op.MBoxID)
}

func (user *User) executeMailboxUpdate(ctx context.Context, op *OpMailboxUpdate) error {
	return user.conn.UpdateLabel(ctx, op.MBoxID, op.Name)
}

func (user *User) executeMessageCreate(ctx context.Context, op *OpMessageCreate) error {
	msg, err := user.conn.CreateMessage(ctx, op.MBoxID, op.Literal, op.Flags, op.Date)
	if err != nil {
		return err
	}

	user.setMessageID(op.TempID, msg.ID)

	user.send(imap.NewMessageIDChanged(op.TempID, msg.ID), true)

	return nil
}

func (user *User) executeMessageAdd(ctx context.Context, op *OpMessageAdd) error {
	if err := user.conn.LabelMessages(ctx, op.MessageIDs, op.MBoxID); err != nil {
		return user.refresh(ctx, op.MessageIDs, op.MBoxID)
	}

	return nil
}

func (user *User) executeMessageRemove(ctx context.Context, op *OpMessageRemove) error {
	if err := user.conn.UnlabelMessages(ctx, op.MessageIDs, op.MBoxID); err != nil {
		return user.refresh(ctx, op.MessageIDs, op.MBoxID)
	}

	return nil
}

func (user *User) executeMessageSeen(ctx context.Context, op *OpMessageSeen) error {
	if err := user.conn.MarkMessagesSeen(ctx, op.MessageIDs, op.Seen); err != nil {
		return user.refresh(ctx, op.MessageIDs)
	}

	return nil
}

func (user *User) executeMessageFlagged(ctx context.Context, op *OpMessageFlagged) error {
	if err := user.conn.MarkMessagesFlagged(ctx, op.MessageIDs, op.Flagged); err != nil {
		return user.refresh(ctx, op.MessageIDs)
	}

	return nil
}

func (user *User) refresh(ctx context.Context, messageIDs []string, mboxIDs ...string) error {
	for _, messageID := range messageIDs {
		message, mboxIDs, err := user.conn.GetMessage(ctx, messageID)
		if err != nil {
			return err
		}

		user.send(imap.NewMessageUpdated(
			message.ID,
			mboxIDs,
			message.Flags.Contains(imap.FlagSeen),
			message.Flags.Contains(imap.FlagFlagged),
		), true)
	}

	for _, mboxID := range mboxIDs {
		mailbox, err := user.conn.GetLabel(ctx, mboxID)
		if err != nil {
			return err
		}

		user.send(imap.NewMailboxUpdated(
			mailbox.ID,
			mailbox.Name,
		), true)
	}

	return nil
}
