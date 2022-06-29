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
	"github.com/ProtonMail/gluon/internal"
	"github.com/ProtonMail/gluon/store"
	"github.com/emersion/go-imap/client"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
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

var (
	TestServerVersionInfo = internal.VersionInfo{
		Name:       "gluon-test-server",
		Version:    internal.Version{Major: 1, Minor: 1, Patch: 1},
		Vendor:     "Proton",
		SupportURL: "",
	}
)

type pathGenerator interface {
	GenerateBackendPath() string
	GenerateStorePath() string
	GenerateUserPath(user string) string
}

type defaultPathGenerator struct {
	tb testing.TB
}

func (dpg *defaultPathGenerator) GenerateBackendPath() string {
	return dpg.tb.TempDir()
}

func (dpg *defaultPathGenerator) GenerateStorePath() string {
	return dpg.tb.TempDir()
}

func (dpg *defaultPathGenerator) GenerateUserPath(user string) string {
	tmpDir := dpg.tb.TempDir()
	return getEntPath(tmpDir)
}

type fixedPathGenerator struct {
	backendPath string
	storePath   string
	userPath    string
	userPaths   map[string]string
}

func newFixedPathGenerator(backendPath, storePath, userPath string) pathGenerator {
	return &fixedPathGenerator{
		backendPath: backendPath,
		userPath:    userPath,
		storePath:   storePath,
		userPaths:   make(map[string]string),
	}
}

func (fpg *fixedPathGenerator) GenerateBackendPath() string {
	return fpg.storePath
}

func (fpg *fixedPathGenerator) GenerateStorePath() string {
	return fpg.storePath
}

func (fpg *fixedPathGenerator) GenerateUserPath(user string) string {
	if v, ok := fpg.userPaths[user]; ok {
		return v
	}

	newUserPath := getEntPath(fpg.userPath)
	fpg.userPaths[user] = newUserPath

	return newUserPath
}

// runServer initializes and starts the mailserver.
func runServer(tb testing.TB, creds []credentials, delim string, tests func(*testSession)) {
	runServerWithPaths(tb, creds, delim, &defaultPathGenerator{tb: tb}, tests)
}

// runServerWithPaths initializes and starts the mailserver using a pathGenerator.
func runServerWithPaths(tb testing.TB, creds []credentials, delim string, pathGenerator pathGenerator, tests func(*testSession)) {
	loggerIn := logrus.StandardLogger().WriterLevel(logrus.TraceLevel)
	loggerOut := logrus.StandardLogger().WriterLevel(logrus.TraceLevel)

	// Setup goroutine leak detector here so that it doesn't report the goroutines created by logrus.
	defer goleak.VerifyNone(tb, goleak.IgnoreCurrent())

	gluonPath := pathGenerator.GenerateBackendPath()
	logrus.Tracef("Backend Path: %v", gluonPath)
	server, err := gluon.New(
		gluonPath,
		gluon.WithDelimiter(delim),
		gluon.WithTLS(&tls.Config{
			Certificates: []tls.Certificate{testCert},
			MinVersion:   tls.VersionTLS13,
		}),
		gluon.WithLogger(
			loggerIn,
			loggerOut,
		),
		gluon.WithVersionInfo(TestServerVersionInfo.Version.Major, TestServerVersionInfo.Version.Minor, TestServerVersionInfo.Version.Patch,
			TestServerVersionInfo.Name, TestServerVersionInfo.Vendor, TestServerVersionInfo.SupportURL),
	)
	require.NoError(tb, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userIDs := make(map[string]string)
	conns := make(map[string]Connector)
	dbPaths := make(map[string]string)

	for _, creds := range creds {
		conn := connector.NewDummy(
			creds.usernames,
			creds.password,
			defaultPeriod,
			defaultFlags,
			defaultPermanentFlags,
			defaultAttributes,
		)

		storePath := pathGenerator.GenerateStorePath()
		logrus.Tracef("User Store Path: %v=%v", creds.usernames[0], storePath)
		store, err := store.NewOnDiskStore(storePath, []byte(creds.password))
		require.NoError(tb, err)

		entPath := pathGenerator.GenerateUserPath(creds.usernames[0])
		logrus.Tracef("User DB path: %v=%v", creds.usernames[0], entPath)
		userID, err := server.AddUser(ctx, conn, store, dialect.SQLite, entPath)
		require.NoError(tb, err)

		require.NoError(tb, conn.Sync(ctx))

		for _, username := range creds.usernames {
			userIDs[username] = userID
		}

		conns[userID] = conn
		dbPaths[userID] = entPath
	}

	listener, err := net.Listen("tcp", net.JoinHostPort("localhost", "0"))
	require.NoError(tb, err)

	errCh := server.Serve(ctx, listener)

	// Run the test against the server.
	tests(newTestSession(tb, listener, server, userIDs, conns, dbPaths))

	// Flush and remove user before shutdown.
	for userID, conn := range conns {
		conn.Flush()
		require.NoError(tb, server.RemoveUser(ctx, userID))
	}

	// Expect the server to shut down successfully when closed.
	require.NoError(tb, server.Close(ctx))
	require.NoError(tb, <-errCh)
}

// runServerWithPaths initializes and starts the mailserver.

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
