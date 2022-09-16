package imap_benchmarks

import (
	"fmt"
	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"time"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/utils"
	"github.com/emersion/go-imap/client"
)

// BuildMailbox creates a mailbox of name `mailbox` and fills it up with `messageCount` random messages.
func BuildMailbox(cl *client.Client, mailbox string, messageCount int) error {
	var messages []string

	if len(*flags.IMAPMailboxMessageDir) != 0 {
		msgs, err := utils.LoadEMLFilesFromDirectory(*flags.IMAPMailboxMessageDir)
		if err != nil {
			return err
		}

		if *flags.Verbose {
			fmt.Printf("Loaded %v messages from '%v'\n", len(msgs), *flags.IMAPMailboxMessageDir)
		}

		messages = msgs
	} else {
		messages = []string{utils.MessageAfterNoonMeeting, utils.MessageMultiPartMixed, utils.MessageEmbedded}
	}

	return BuildMailboxWithMessages(cl, mailbox, messageCount, messages)
}

func BuildMailboxWithMessages(cl *client.Client, mailbox string, messageCount int, messages []string) error {
	messagesLen := len(messages)

	for i := 0; i < messageCount; i++ {
		if err := AppendToMailbox(cl, mailbox, messages[i%messagesLen], time.Now()); err != nil {
			return err
		}
	}

	return nil
}
