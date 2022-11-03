package connector

import (
	"github.com/ProtonMail/gluon/imap"
)

func (conn *Dummy) SetFolderPrefix(pfx string) {
	defer conn.Flush()

	conn.pfxFolder = pfx

	mbox := conn.state.createMailbox([]string{pfx}, true)

	mbox.Attributes = mbox.Attributes.Add(imap.AttrNoSelect)

	conn.pushUpdate(imap.NewMailboxCreated(mbox))
}

func (conn *Dummy) SetLabelsPrefix(pfx string) {
	defer conn.Flush()

	conn.pfxLabel = pfx

	mbox := conn.state.createMailbox([]string{pfx}, false)

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

	conn.state.mailboxes[mbox.ID] = &dummyMailbox{
		mboxName:  mbox.Name,
		exclusive: exclusive,
	}

	conn.pushUpdate(imap.NewMailboxCreated(mbox))

	return nil
}

func (conn *Dummy) MailboxDeleted(mboxID imap.MailboxID) error {
	conn.state.deleteMailbox(mboxID)

	conn.pushUpdate(imap.NewMailboxDeleted(mboxID))

	return nil
}

func (conn *Dummy) MessageCreated(message imap.Message, literal []byte, mboxIDs []imap.MailboxID) error {
	parsedMessage, err := imap.NewParsedMessage(literal)
	if err != nil {
		return err
	}

	conn.state.lock.Lock()
	defer conn.state.lock.Unlock()

	mboxIDMap := make(map[imap.MailboxID]struct{})

	for _, mboxID := range mboxIDs {
		mboxIDMap[mboxID] = struct{}{}
	}

	conn.state.messages[message.ID] = &dummyMessage{
		literal: literal,
		seen:    message.Flags.Contains(imap.FlagSeen),
		flagged: message.Flags.Contains(imap.FlagFlagged),
		parsed:  parsedMessage,
		date:    message.Date,
		mboxIDs: mboxIDMap,
	}

	update := imap.NewMessagesCreated(&imap.MessageCreated{
		Message:       message,
		Literal:       literal,
		MailboxIDs:    mboxIDs,
		ParsedMessage: parsedMessage,
	})

	conn.pushUpdate(update)

	return nil
}

func (conn *Dummy) MessagesCreated(messages []imap.Message, literals [][]byte, mboxIDs [][]imap.MailboxID) error {
	conn.state.lock.Lock()
	defer conn.state.lock.Unlock()

	var updates []*imap.MessageCreated

	for i := 0; i < len(messages); i++ {
		parsedMessage, err := imap.NewParsedMessage(literals[i])
		if err != nil {
			return err
		}

		mboxIDMap := make(map[imap.MailboxID]struct{})

		for _, mboxID := range mboxIDs[i] {
			mboxIDMap[mboxID] = struct{}{}
		}

		conn.state.messages[messages[i].ID] = &dummyMessage{
			literal: literals[i],
			seen:    messages[i].Flags.Contains(imap.FlagSeen),
			flagged: messages[i].Flags.Contains(imap.FlagFlagged),
			parsed:  parsedMessage,
			date:    messages[i].Date,
			mboxIDs: mboxIDMap,
		}

		updates = append(updates, &imap.MessageCreated{
			Message:       messages[i],
			Literal:       literals[i],
			MailboxIDs:    mboxIDs[i],
			ParsedMessage: parsedMessage,
		})
	}

	conn.pushUpdate(imap.NewMessagesCreated(updates...))

	return nil
}

func (conn *Dummy) MessageUpdated(message imap.Message, literal []byte, mboxIDs []imap.MailboxID) error {
	conn.state.lock.Lock()
	defer conn.state.lock.Unlock()

	parsedMessage, err := imap.NewParsedMessage(literal)
	if err != nil {
		return err
	}

	mboxIDMap := make(map[imap.MailboxID]struct{})

	for _, mboxID := range mboxIDs {
		mboxIDMap[mboxID] = struct{}{}
	}

	conn.state.messages[message.ID] = &dummyMessage{
		literal: literal,
		seen:    message.Flags.Contains(imap.FlagSeen),
		flagged: message.Flags.Contains(imap.FlagFlagged),
		parsed:  parsedMessage,
		date:    message.Date,
		mboxIDs: mboxIDMap,
	}

	conn.pushUpdate(imap.NewMessageUpdated(message, literal, mboxIDs, parsedMessage))

	return nil
}

func (conn *Dummy) MessageAdded(messageID imap.MessageID, mboxID imap.MailboxID) error {
	conn.state.addMessageToMailbox(messageID, mboxID)

	conn.pushUpdate(imap.NewMessageMailboxesUpdated(
		messageID,
		conn.state.getMailboxIDs(messageID),
		conn.state.isSeen(messageID),
		conn.state.isFlagged(messageID),
	))

	return nil
}

func (conn *Dummy) MessageRemoved(messageID imap.MessageID, mboxID imap.MailboxID) error {
	conn.state.removeMessageFromMailbox(messageID, mboxID)

	conn.pushUpdate(imap.NewMessageMailboxesUpdated(
		messageID,
		conn.state.getMailboxIDs(messageID),
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
