package connector

import "github.com/ProtonMail/gluon/imap"

func (conn *Dummy) SetFolderPrefix(pfx string) {
	defer conn.Flush()

	conn.pfxFolder = pfx

	mbox := conn.state.createLabel([]string{pfx}, true)

	mbox.Attributes = mbox.Attributes.Add(imap.AttrNoSelect)

	conn.pushUpdate(imap.NewMailboxCreated(mbox))
}

func (conn *Dummy) SetLabelPrefix(pfx string) {
	defer conn.Flush()

	conn.pfxLabel = pfx

	mbox := conn.state.createLabel([]string{pfx}, false)

	mbox.Attributes = mbox.Attributes.Add(imap.AttrNoSelect)

	conn.pushUpdate(imap.NewMailboxCreated(mbox))
}

func (conn *Dummy) MailboxCreated(mbox imap.Mailbox) error {
	conn.state.lock.Lock()
	defer conn.state.lock.Unlock()

	exclusive, err := conn.validateName(mbox.Name)
	if err != nil {
		return err
	}

	conn.state.labels[mbox.ID] = &dummyLabel{
		labelName: mbox.Name,
		exclusive: exclusive,
	}

	conn.pushUpdate(imap.NewMailboxCreated(mbox))

	return nil
}

func (conn *Dummy) MailboxDeleted(labelID imap.LabelID) error {
	conn.state.deleteLabel(labelID)

	conn.pushUpdate(imap.NewMailboxDeleted(labelID))

	return nil
}

func (conn *Dummy) MessageCreated(message imap.Message, literal []byte, mboxIDs []imap.LabelID) error {
	conn.state.lock.Lock()
	defer conn.state.lock.Unlock()

	labelIDs := make(map[imap.LabelID]struct{})

	for _, mboxID := range mboxIDs {
		labelIDs[mboxID] = struct{}{}
	}

	conn.state.messages[message.ID] = &dummyMessage{
		literal:  literal,
		seen:     message.Flags.Contains(imap.FlagSeen),
		flagged:  message.Flags.Contains(imap.FlagFlagged),
		date:     message.Date,
		labelIDs: labelIDs,
	}

	update := imap.NewMessagesCreated()

	if err := update.Add(message, literal, mboxIDs...); err != nil {
		return err
	}

	conn.pushUpdate(update)

	return nil
}

func (conn *Dummy) MessageAdded(messageID imap.MessageID, labelID imap.LabelID) error {
	conn.state.labelMessage(messageID, labelID)

	conn.pushUpdate(imap.NewMessageLabelsUpdated(
		messageID,
		conn.state.getLabelIDs(messageID),
		conn.state.isSeen(messageID),
		conn.state.isFlagged(messageID),
	))

	return nil
}

func (conn *Dummy) MessageRemoved(messageID imap.MessageID, labelID imap.LabelID) error {
	conn.state.unlabelMessage(messageID, labelID)

	conn.pushUpdate(imap.NewMessageLabelsUpdated(
		messageID,
		conn.state.getLabelIDs(messageID),
		conn.state.isSeen(messageID),
		conn.state.isFlagged(messageID),
	))

	return nil
}

func (conn *Dummy) MessageSeen(messageID imap.MessageID, seen bool) error {
	conn.state.setSeen(messageID, seen)

	conn.pushUpdate(imap.NewMessageFlagsUpdated(
		messageID,
		conn.state.isSeen(messageID),
		conn.state.isFlagged(messageID),
	))

	return nil
}

func (conn *Dummy) MessageFlagged(messageID imap.MessageID, flagged bool) error {
	conn.state.setFlagged(messageID, flagged)

	conn.pushUpdate(imap.NewMessageFlagsUpdated(
		messageID,
		conn.state.isSeen(messageID),
		conn.state.isFlagged(messageID),
	))

	return nil
}

func (conn *Dummy) MessageDeleted(messageID imap.MessageID) error {
	conn.pushUpdate(imap.NewMessagesDeleted(messageID))

	return nil
}

func (conn *Dummy) Flush() {
	conn.ticker.Poll()
}
