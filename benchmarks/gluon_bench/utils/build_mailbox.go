package utils

import (
	"time"

	"github.com/emersion/go-imap/client"
	"golang.org/x/exp/rand"
)

// BuildMailbox creates a mailbox of name `mailbox` and fills it up with `messageCount` random messages.
func BuildMailbox(cl *client.Client, mailbox string, messageCount int) error {
	messages := []string{MessageAfterNoonMeeting, MessageMultiPartMixed, MessageEmbedded}
	messagesLen := len(messages)

	for i := 0; i < messageCount; i++ {
		if err := AppendToMailbox(cl, mailbox, messages[rand.Intn(messagesLen)], time.Now()); err != nil {
			return err
		}
	}

	return nil
}
