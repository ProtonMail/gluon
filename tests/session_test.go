package tests

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/utils"
	"github.com/ProtonMail/go-mbox"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

type Connector interface {
	connector.Connector

	SetFolderPrefix(string)
	SetLabelsPrefix(string)

	MailboxCreated(imap.Mailbox) error
	MailboxDeleted(imap.MailboxID) error
	SetMailboxVisible(imap.MailboxID, bool)

	SetAllowMessageCreateWithUnknownMailboxID(value bool)

	MessageCreated(imap.Message, []byte, []imap.MailboxID) error
	MessagesCreated([]imap.Message, [][]byte, [][]imap.MailboxID) error
	MessageUpdated(imap.Message, []byte, []imap.MailboxID) error
	MessageAdded(imap.MessageID, imap.MailboxID) error
	MessageRemoved(imap.MessageID, imap.MailboxID) error
	MessageSeen(imap.MessageID, bool) error
	MessageFlagged(imap.MessageID, bool) error
	MessageDeleted(imap.MessageID) error

	UIDValidityBumped()

	GetLastRecordedIMAPID() imap.IMAPID

	Sync(context.Context) error

	Flush()
}

type testSession struct {
	tb testing.TB

	listener    net.Listener
	server      *gluon.Server
	eventCh     <-chan events.Event
	reporter    *testReporter
	userIDs     map[string]string
	conns       map[string]Connector
	userDBPaths map[string]string
	options     *serverOptions
}

func newTestSession(
	tb testing.TB,
	listener net.Listener,
	server *gluon.Server,
	eventCh <-chan events.Event,
	reporter *testReporter,
	userIDs map[string]string,
	conns map[string]Connector,
	userDBPaths map[string]string,
	options *serverOptions,
) *testSession {
	return &testSession{
		tb:          tb,
		listener:    listener,
		server:      server,
		eventCh:     eventCh,
		reporter:    reporter,
		userIDs:     userIDs,
		conns:       conns,
		userDBPaths: userDBPaths,
		options:     options,
	}
}

func (s *testSession) newConnection() *testConnection {
	conn, err := net.Dial(s.listener.Addr().Network(), s.listener.Addr().String())
	require.NoError(s.tb, err)

	return newTestConnection(s.tb, conn).Sx(`\* OK.*`)
}

func (s *testSession) withConnection(user string, fn func(*testConnection)) {
	conn := s.newConnection()
	defer func() { require.NoError(s.tb, conn.disconnect()) }()

	fn(conn.Login(user, s.options.password(user)))
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

func (s *testSession) setLabelsPrefix(user, prefix string) {
	s.conns[s.userIDs[user]].SetLabelsPrefix(prefix)
}

func (s *testSession) mailboxCreated(user string, name []string, withData ...string) imap.MailboxID {
	return s.mailboxCreatedWithAttributes(user, name, defaultAttributes, withData...)
}

func (s *testSession) setAllowMessageCreateWithUnknownMailboxID(user string, value bool) {
	s.conns[s.userIDs[user]].SetAllowMessageCreateWithUnknownMailboxID(value)
}

func (s *testSession) mailboxDeleted(user string, id imap.MailboxID) {
	require.NoError(s.tb, s.conns[s.userIDs[user]].MailboxDeleted(id))
}

func (s *testSession) mailboxCreatedWithAttributes(user string, name []string, attributes imap.FlagSet, withData ...string) imap.MailboxID {
	mboxID := imap.MailboxID(utils.NewRandomMailboxID())

	require.NoError(s.tb, s.conns[s.userIDs[user]].MailboxCreated(imap.Mailbox{
		ID:             mboxID,
		Name:           name,
		Flags:          defaultFlags,
		PermanentFlags: defaultPermanentFlags,
		Attributes:     attributes,
	}))

	for _, data := range withData {
		s.messagesCreatedFromMBox(user, mboxID, data)
	}

	s.conns[s.userIDs[user]].Flush()

	return mboxID
}

func (s *testSession) batchMailboxCreated(user string, count int, mailboxNameGen func(number int) string) []imap.MailboxID {
	var mboxIDs []imap.MailboxID

	for i := 0; i < count; i++ {
		mboxID := imap.MailboxID(utils.NewRandomMailboxID())

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

func (s *testSession) mailboxCreatedCustom(user string, name []string, flags, permFlags, attrs imap.FlagSet) imap.MailboxID {
	mboxID := imap.MailboxID(utils.NewRandomMailboxID())

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

func (s *testSession) messageCreatedWithMailboxes(user string, mailboxIDs []imap.MailboxID, literal []byte, internalDate time.Time, flags ...string) imap.MessageID {
	messageID := imap.MessageID(utils.NewRandomMessageID())

	require.NoError(s.tb, s.conns[s.userIDs[user]].MessageCreated(
		imap.Message{
			ID:    messageID,
			Flags: imap.NewFlagSetFromSlice(flags),
			Date:  internalDate,
		},
		literal,
		mailboxIDs,
	))

	s.conns[s.userIDs[user]].Flush()

	return messageID
}

func (s *testSession) messageCreated(user string, mailboxID imap.MailboxID, literal []byte, internalDate time.Time, flags ...string) imap.MessageID {
	messageID := imap.MessageID(utils.NewRandomMessageID())

	s.messageCreatedWithID(user, messageID, mailboxID, literal, internalDate, flags...)

	return messageID
}

func (s *testSession) messageCreatedWithID(user string, messageID imap.MessageID, mailboxID imap.MailboxID, literal []byte, internalDate time.Time, flags ...string) {
	require.NoError(s.tb, s.conns[s.userIDs[user]].MessageCreated(
		imap.Message{
			ID:    messageID,
			Flags: imap.NewFlagSetFromSlice(flags),
			Date:  internalDate,
		},
		literal,
		[]imap.MailboxID{mailboxID},
	))

	s.conns[s.userIDs[user]].Flush()
}

func (s *testSession) messageUpdatedWithID(user string, messageID imap.MessageID, mailboxID imap.MailboxID, literal []byte, internalDate time.Time, flags ...string) {
	require.NoError(s.tb, s.conns[s.userIDs[user]].MessageUpdated(
		imap.Message{
			ID:    messageID,
			Flags: imap.NewFlagSetFromSlice(flags),
			Date:  internalDate,
		},
		literal,
		[]imap.MailboxID{mailboxID},
	))

	s.conns[s.userIDs[user]].Flush()
}

func (s *testSession) batchMessageCreated(user string, mailboxID imap.MailboxID, count int, createMessage func(int) ([]byte, []string)) []imap.MessageID {
	return s.batchMessageCreatedWithID(user, mailboxID, count, func(i int) (imap.MessageID, []byte, []string) {
		messageID := imap.MessageID(utils.NewRandomMessageID())
		literal, flags := createMessage(i)

		return messageID, literal, flags
	})
}

func (s *testSession) batchMessageCreatedWithID(user string, mailboxID imap.MailboxID, count int, createMessage func(int) (imap.MessageID, []byte, []string)) []imap.MessageID {
	var messageIDs []imap.MessageID

	messages := make([]imap.Message, 0, count)
	literals := make([][]byte, 0, count)
	mailboxes := make([][]imap.MailboxID, 0, count)

	for i := 0; i < count; i++ {
		messageID, literal, flags := createMessage(i)

		messages = append(messages, imap.Message{
			ID:    messageID,
			Flags: imap.NewFlagSetFromSlice(flags),
			Date:  time.Now(),
		})

		literals = append(literals, literal)

		mailboxes = append(mailboxes, []imap.MailboxID{mailboxID})

		messageIDs = append(messageIDs, messageID)
	}

	require.NoError(s.tb, s.conns[s.userIDs[user]].MessagesCreated(messages, literals, mailboxes))

	s.conns[s.userIDs[user]].Flush()

	return messageIDs
}

func (s *testSession) messageCreatedFromFile(user string, mailboxID imap.MailboxID, path string, flags ...string) imap.MessageID {
	literal, err := os.ReadFile(path)
	require.NoError(s.tb, err)

	return s.messageCreated(user, mailboxID, literal, time.Now(), flags...)
}

func (s *testSession) messagesCreatedFromMBox(user string, mailboxID imap.MailboxID, path string, flags ...string) {
	f, err := os.Open(path)
	require.NoError(s.tb, err)

	require.NoError(s.tb, forMessageInMBox(f, func(messageDelimiter, literal []byte) {
		// If possible use mbox delimiter time as internal date to able to
		// test cases where header and internal date are different.
		internalDate, err := parseDateFromDelimiter(string(messageDelimiter))
		if err != nil {
			internalDate = time.Now()
		}

		s.messageCreated(user, mailboxID, literal, internalDate, flags...)
	}))

	require.NoError(s.tb, f.Close())
}

func (s *testSession) messageAdded(user string, messageID imap.MessageID, mailboxID imap.MailboxID) {
	require.NoError(s.tb, s.conns[s.userIDs[user]].MessageAdded(messageID, mailboxID))

	s.conns[s.userIDs[user]].Flush()
}

func (s *testSession) messageRemoved(user string, messageID imap.MessageID, mailboxID imap.MailboxID) {
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

func (s *testSession) uidValidityBumped(user string) {
	s.conns[s.userIDs[user]].UIDValidityBumped()
}

func (s *testSession) flush(user string) {
	s.conns[s.userIDs[user]].Flush()
}

func forMessageInMBox(rr io.Reader, fn func(messageDelimiter, literal []byte)) error {
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

		fn(mr.GetMessageDelimiter(), literal)
	}

	if !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}

func parseDateFromDelimiter(messageDelimiter string) (t time.Time, err error) {
	split := strings.Split(messageDelimiter, " ")
	if len(split) <= 3 {
		return t, errors.New("not enough arguments in delimiter")
	}

	return time.Parse("Mon Jan _2 15:04:05 2006", strings.TrimSpace(strings.Join(split[2:], " ")))
}

func TestTooManyInvalidCommands(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		for i := 1; i <= 19; i++ {
			c.Cf("%d FOO", i).BAD(fmt.Sprintf("%d", i))
		}

		// The next command should fail; the server should disconnect the client.
		c.Cf("100 FOO").BAD("100")

		// The client should be disconnected.
		_, err := c.conn.Read(make([]byte, 1))
		require.Error(t, err)
	})
}

func TestResetTooManyInvalidCommands(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		for i := 1; i <= 19; i++ {
			c.Cf("%d FOO", i).BAD(fmt.Sprintf("%d", i))
		}

		// The next command should succeed; the counter should be reset.
		c.C("100 NOOP").OK("100")

		for i := 1; i <= 19; i++ {
			c.Cf("%d FOO", i).BAD(fmt.Sprintf("%d", i))
		}

		// The next command should fail; the server should disconnect the client.
		c.Cf("100 FOO").BAD("100")

		// The client should be disconnected.
		_, err := c.conn.Read(make([]byte, 1))
		require.Error(t, err)
	})
}
