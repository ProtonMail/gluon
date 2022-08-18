package remote

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
)

func (user *User) refresh(ctx context.Context, messageIDs []imap.MessageID, mboxIDs ...imap.LabelID) error {
	for _, messageID := range messageIDs {
		message, mboxIDs, err := user.conn.GetMessage(ctx, messageID)
		if err != nil {
			return err
		}

		user.send(imap.NewMessageLabelsUpdated(
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
