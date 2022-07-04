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
	"github.com/ProtonMail/gluon/internal/backend/ent"
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
	MailboxDeleted(string) error

	MessageCreated(imap.Message, []byte, []string) error
	MessageAdded(string, string) error
	MessageRemoved(string, string) error
	MessageSeen(string, bool) error
	MessageFlagged(string, bool) error
	MessageDeleted(string) error

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

func (s *testSession) mailboxCreated(user string, name []string, withData ...string) string {
	mboxID := utils.NewRandomLabelID()

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

func (s *testSession) mailboxCreatedCustom(user string, name []string, flags, permFlags, attrs imap.FlagSet) string {
	mboxID := utils.NewRandomLabelID()

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

func (s *testSession) messageCreated(user, mailboxID string, literal []byte, flags ...string) string {
	messageID := utils.NewRandomMessageID()

	require.NoError(s.tb, s.conns[s.userIDs[user]].MessageCreated(
		imap.Message{
			ID:    messageID,
			Flags: imap.NewFlagSetFromSlice(flags),
			Date:  time.Now(),
		},
		literal,
		[]string{mailboxID},
	))

	s.conns[s.userIDs[user]].Flush()

	return messageID
}

func (s *testSession) batchMessageCreated(user string, mailboxID string, count int, createMessage func(int) ([]byte, []string)) []string {
	var messageIDs []string

	for i := 0; i < count; i++ {
		messageID := utils.NewRandomMessageID()

		literal, flags := createMessage(i)

		require.NoError(s.tb, s.conns[s.userIDs[user]].MessageCreated(
			imap.Message{
				ID:    messageID,
				Flags: imap.NewFlagSetFromSlice(flags),
				Date:  time.Now(),
			},
			literal,
			[]string{mailboxID},
		))

		messageIDs = append(messageIDs, messageID)
	}
	s.conns[s.userIDs[user]].Flush()

	return messageIDs
}

func (s *testSession) messageCreatedFromFile(user, mailboxID, path string, flags ...string) string {
	literal, err := os.ReadFile(path)
	require.NoError(s.tb, err)

	return s.messageCreated(user, mailboxID, literal, flags...)
}

func (s *testSession) messagesCreatedFromMBox(user, mailboxID, path string, flags ...string) {
	f, err := os.Open(path)
	require.NoError(s.tb, err)

	require.NoError(s.tb, forMessageInMBox(f, func(literal []byte) {
		s.messageCreated(user, mailboxID, literal, flags...)
	}))

	require.NoError(s.tb, f.Close())
}

func (s *testSession) messageAdded(user, messageID, mailboxID string) {
	require.NoError(s.tb, s.conns[s.userIDs[user]].MessageAdded(messageID, mailboxID))

	s.conns[s.userIDs[user]].Flush()
}

func (s *testSession) messageRemoved(user, messageID, mailboxID string) {
	require.NoError(s.tb, s.conns[s.userIDs[user]].MessageRemoved(messageID, mailboxID))

	s.conns[s.userIDs[user]].Flush()
}

func (s *testSession) messageDeleted(user, messageID string) {
	require.NoError(s.tb, s.conns[s.userIDs[user]].MessageDeleted(messageID))

	s.conns[s.userIDs[user]].Flush()
}

func (s *testSession) messageSeen(user, messageID string, seen bool) {
	require.NoError(s.tb, s.conns[s.userIDs[user]].MessageSeen(messageID, seen))

	s.conns[s.userIDs[user]].Flush()
}

func (s *testSession) messageFlagged(user, messageID string, flagged bool) {
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
