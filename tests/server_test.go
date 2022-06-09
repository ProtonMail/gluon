package tests

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/store"
	"github.com/emersion/go-imap/client"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const defaultPeriod = time.Second

var (
	defaultFlags          = imap.NewFlagSet(imap.FlagSeen, imap.FlagFlagged, imap.FlagDeleted)
	defaultPermanentFlags = imap.NewFlagSet(imap.FlagSeen, imap.FlagFlagged, imap.FlagDeleted)
	defaultAttributes     = imap.NewFlagSet()
)

type credentials struct {
	usernames []string
	password  string
}

// runServer initializes and starts the mailserver.
func runServer(tb testing.TB, creds []credentials, delim string, tests func(*testSession)) {
	server, err := gluon.New(
		tb.TempDir(),
		gluon.WithDelimiter(delim),
		gluon.WithTLS(&tls.Config{
			Certificates: []tls.Certificate{testCert},
			MinVersion:   tls.VersionTLS13,
		}),
		gluon.WithLogger(
			logrus.StandardLogger().WriterLevel(logrus.TraceLevel),
			logrus.StandardLogger().WriterLevel(logrus.TraceLevel),
		),
	)
	require.NoError(tb, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userIDs := make(map[string]string)
	conns := make(map[string]Connector)

	for _, creds := range creds {
		conn := connector.NewDummy(
			creds.usernames,
			creds.password,
			defaultPeriod,
			defaultFlags,
			defaultPermanentFlags,
			defaultAttributes,
		)

		store, err := store.NewOnDiskStore(tb.TempDir(), []byte(creds.password))
		require.NoError(tb, err)

		userID, err := server.AddUser(conn, store, dialect.SQLite, getEntPath(tb.TempDir()))
		require.NoError(tb, err)

		require.NoError(tb, conn.Sync(ctx))

		for _, username := range creds.usernames {
			userIDs[username] = userID
		}

		conns[userID] = conn
	}

	listener, err := net.Listen("tcp", net.JoinHostPort("localhost", "0"))
	require.NoError(tb, err)

	errCh := server.Serve(ctx, listener)

	// Run the test against the server.
	tests(newTestSession(tb, listener, server, userIDs, conns))

	// Flush and remove user before shutdown.
	for userID, conn := range conns {
		conn.Flush()
		require.NoError(tb, server.RemoveUser(ctx, userID))
	}

	// Expect the server to shut down successfully when closed.
	require.NoError(tb, server.Close(ctx))
	require.NoError(tb, <-errCh)
}

func withConnections(tb testing.TB, s *testSession, connIDs []int, tests func(map[int]*testConnection)) {
	conns := make(map[int]*testConnection)

	for _, connID := range connIDs {
		conns[connID] = s.newConnection()
	}

	tests(conns)

	for _, connection := range conns {
		require.NoError(tb, connection.disconnect())
	}
}

func withClients(tb testing.TB, s *testSession, connIDs []int, tests func(map[int]*client.Client)) {
	clients := make(map[int]*client.Client)

	for _, connID := range connIDs {
		clients[connID] = s.newClient()
	}

	tests(clients)

	for _, client := range clients {
		require.NoError(tb, client.Logout())
	}
}

func withData(s *testSession, username string, tests func(string, string)) {
	mbox := uuid.NewString()

	mboxID := s.mailboxCreated(username, []string{mbox}, "testdata/dovecot-crlf")

	tests(mbox, mboxID)
}

func getEntPath(dir string) string {
	return fmt.Sprintf("file:%v?cache=shared&_fk=1", filepath.Join(dir, fmt.Sprintf("%v.db", uuid.NewString())))
}
