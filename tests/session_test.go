package tests

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/utils"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-mbox"
	"github.com/stretchr/testify/require"
)

type Connector interface {
	connector.Connector

	SetFolderPrefix(string)
	SetLabelPrefix(string)

	MailboxCreated(imap.Mailbox) error
	MailboxDeleted(imap.LabelID) error

	MessageCreated(imap.Message, []byte, []imap.LabelID) error
	MessageAdded(imap.MessageID, imap.LabelID) error
	MessageRemoved(imap.MessageID, imap.LabelID) error
	MessageSeen(imap.MessageID, bool) error
	MessageFlagged(imap.MessageID, bool) error
	MessageDeleted(imap.MessageID) error

	Sync(context.Context) error
	Flush()

	GetLastRecordedIMAPID() imap.ID
}

type testSession struct {
	tb testing.TB

	listener      net.Listener
	server        *gluon.Server
	userIDs       map[string]string
	conns         map[string]Connector
	userDBPaths   map[string]string
	serverOptions *serverOptions
}

func newTestSession(
	tb testing.TB,
	listener net.Listener,
	server *gluon.Server,
	userIDs map[string]string,
	conns map[string]Connector,
	userDBPaths map[string]string,
	options *serverOptions,
) *testSession {
	return &testSession{
		tb:            tb,
		listener:      listener,
		server:        server,
		userIDs:       userIDs,
		conns:         conns,
		userDBPaths:   userDBPaths,
		serverOptions: options,
	}
}

func (s *testSession) newConnection() *testConnection {
	conn, err := net.Dial(s.listener.Addr().Network(), s.listener.Addr().String())
	require.NoError(s.tb, err)

	return newTestConnection(s.tb, conn).Sx(`\* OK.*`)
}

func (s *testSession) newClient() *client.Client {
	client, err := client.Dial(s.listener.Addr().String())
	require.NoError(s.tb, err)

	return client
}

func (s *testSession) withUserDB(user string, fn func(client *ent.Client, ctx context.Context)) error {
	path, ok := s.userDBPaths[s.userIDs[user]]
	if !ok {
		return fmt.Errorf("User not found")
	}

	client, err := ent.Open(dialect.SQLite, path)
	if err != nil {
		return err
	}

	fn(client, context.Background())

	return client.Close()
}

func (s *testSession) setFolderPrefix(user, prefix string) {
	s.conns[s.userIDs[user]].SetFolderPrefix(prefix)
}

func (s *testSession) setLabelPrefix(user, prefix string) {
	s.conns[s.userIDs[user]].SetLabelPrefix(prefix)
}

func (s *testSession) mailboxCreated(user string, name []string, withData ...string) imap.LabelID {
	mboxID := imap.LabelID(utils.NewRandomLabelID())

	require.NoError(s.tb, s.conns[s.userIDs[user]].MailboxCreated(imap.Mailbox{
		ID:             mboxID,
		Name:           name,
		Flags:          defaultFlags,
		PermanentFlags: defaultPermanentFlags,
		Attributes:     defaultAttributes,
	}))

	for _, data := range withData {
		s.messagesCreatedFromMBox(user, mboxID, data)
	}

	s.conns[s.userIDs[user]].Flush()

	return mboxID
}

func (s *testSession) batchMailboxCreated(user string, count int, mailboxNameGen func(number int) string) []imap.LabelID {
	var mboxIDs []imap.LabelID

	for i := 0; i < count; i++ {
		mboxID := imap.LabelID(utils.NewRandomLabelID())

		require.NoError(s.tb, s.conns[s.userIDs[user]].MailboxCreated(imap.Mailbox{
			ID:             mboxID,
			Name:           []string{mailboxNameGen(i)},
			Flags:          defaultFlags,
			PermanentFlags: defaultPermanentFlags,
			Attributes:     defaultAttributes,
		}))

		mboxIDs = append(mboxIDs, mboxID)
	}

	s.conns[s.userIDs[user]].Flush()

	return mboxIDs
}

func (s *testSession) mailboxCreatedCustom(user string, name []string, flags, permFlags, attrs imap.FlagSet) imap.LabelID {
	mboxID := imap.LabelID(utils.NewRandomLabelID())

	require.NoError(s.tb, s.conns[s.userIDs[user]].MailboxCreated(imap.Mailbox{
		ID:             mboxID,
		Name:           name,
		Flags:          flags,
		PermanentFlags: permFlags,
		Attributes:     attrs,
	}))

	s.conns[s.userIDs[user]].Flush()

	return mboxID
}

func (s *testSession) messageCreated(user string, mailboxID imap.LabelID, literal []byte, flags ...string) imap.MessageID {
	messageID := imap.MessageID(utils.NewRandomMessageID())

	require.NoError(s.tb, s.conns[s.userIDs[user]].MessageCreated(
		imap.Message{
			ID:    messageID,
			Flags: imap.NewFlagSetFromSlice(flags),
			Date:  time.Now(),
		},
		literal,
		[]imap.LabelID{mailboxID},
	))

	s.conns[s.userIDs[user]].Flush()

	return messageID
}

func (s *testSession) batchMessageCreated(user string, mailboxID imap.LabelID, count int, createMessage func(int) ([]byte, []string)) []imap.MessageID {
	var messageIDs []imap.MessageID

	for i := 0; i < count; i++ {
		messageID := imap.MessageID(utils.NewRandomMessageID())

		literal, flags := createMessage(i)

		require.NoError(s.tb, s.conns[s.userIDs[user]].MessageCreated(
			imap.Message{
				ID:    messageID,
				Flags: imap.NewFlagSetFromSlice(flags),
				Date:  time.Now(),
			},
			literal,
			[]imap.LabelID{mailboxID},
		))

		messageIDs = append(messageIDs, messageID)
	}
	s.conns[s.userIDs[user]].Flush()

	return messageIDs
}

func (s *testSession) messageCreatedFromFile(user string, mailboxID imap.LabelID, path string, flags ...string) imap.MessageID {
	literal, err := os.ReadFile(path)
	require.NoError(s.tb, err)

	return s.messageCreated(user, mailboxID, literal, flags...)
}

func (s *testSession) messagesCreatedFromMBox(user string, mailboxID imap.LabelID, path string, flags ...string) {
	f, err := os.Open(path)
	require.NoError(s.tb, err)

	require.NoError(s.tb, forMessageInMBox(f, func(literal []byte) {
		s.messageCreated(user, mailboxID, literal, flags...)
	}))

	require.NoError(s.tb, f.Close())
}

func (s *testSession) messageAdded(user string, messageID imap.MessageID, mailboxID imap.LabelID) {
	require.NoError(s.tb, s.conns[s.userIDs[user]].MessageAdded(messageID, mailboxID))

	s.conns[s.userIDs[user]].Flush()
}

func (s *testSession) messageRemoved(user string, messageID imap.MessageID, mailboxID imap.LabelID) {
	require.NoError(s.tb, s.conns[s.userIDs[user]].MessageRemoved(messageID, mailboxID))

	s.conns[s.userIDs[user]].Flush()
}

func (s *testSession) messageDeleted(user string, messageID imap.MessageID) {
	require.NoError(s.tb, s.conns[s.userIDs[user]].MessageDeleted(messageID))

	s.conns[s.userIDs[user]].Flush()
}

func (s *testSession) messageSeen(user string, messageID imap.MessageID, seen bool) {
	require.NoError(s.tb, s.conns[s.userIDs[user]].MessageSeen(messageID, seen))

	s.conns[s.userIDs[user]].Flush()
}

func (s *testSession) messageFlagged(user string, messageID imap.MessageID, flagged bool) {
	require.NoError(s.tb, s.conns[s.userIDs[user]].MessageFlagged(messageID, flagged))

	s.conns[s.userIDs[user]].Flush()
}

func (s *testSession) flush(user string) {
	s.conns[s.userIDs[user]].Flush()
}

func forMessageInMBox(rr io.Reader, fn func([]byte)) error {
	mr := mbox.NewReader(rr)

	var (
		r   io.Reader
		err error
	)

	for r, err = mr.NextMessage(); err == nil; r, err = mr.NextMessage() {
		literal, err := io.ReadAll(r)
		if err != nil {
			return err
		}

		fn(literal)
	}

	if !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}
