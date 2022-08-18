package tests

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

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

type serverOptions struct {
	credentials []credentials
	delimiter   string
	dataDir     string
}

func (s *serverOptions) defaultUsername() string {
	return s.credentials[0].usernames[0]
}

func (s *serverOptions) defaultUserPassword() string {
	return s.credentials[0].password
}

type serverOption interface {
	apply(options *serverOptions)
}

type delimiterServerOption struct {
	delimiter string
}

func (d *delimiterServerOption) apply(options *serverOptions) {
	options.delimiter = d.delimiter
}

type dataDirOption struct {
	dir string
}

func (opt *dataDirOption) apply(options *serverOptions) {
	options.dataDir = opt.dir
}

type credentialsSeverOption struct {
	credentials []credentials
}

func (c *credentialsSeverOption) apply(options *serverOptions) {
	options.credentials = c.credentials
}

func withDelimiter(delimiter string) serverOption {
	return &delimiterServerOption{delimiter: delimiter}
}

func withDataDir(dir string) serverOption {
	return &dataDirOption{dir: dir}
}

func withCredentials(credentials []credentials) serverOption {
	return &credentialsSeverOption{credentials: credentials}
}

func defaultServerOptions(tb testing.TB, modifiers ...serverOption) *serverOptions {
	options := &serverOptions{
		credentials: []credentials{{
			usernames: []string{"user"},
			password:  "pass",
		}},
		delimiter: "/",
		dataDir:   tb.TempDir(),
	}

	for _, op := range modifiers {
		op.apply(options)
	}

	return options
}

// Wrapper to ensure we always pass 32bytes worth of encryption key to the tests.
type testBadgerStoreBuilder struct{}

func (*testBadgerStoreBuilder) New(directory, userID string, encryptionPassphrase []byte) (store.Store, error) {
	encryptionBytes := sha256.Sum256(encryptionPassphrase)
	return store.NewTestBadgerStore(directory, userID, encryptionBytes[:])
}

// runServerWithPaths initializes and starts the mailserver using a pathGenerator.
func runServer(tb testing.TB, options *serverOptions, tests func(*testSession)) {
	loggerIn := logrus.StandardLogger().WriterLevel(logrus.TraceLevel)
	loggerOut := logrus.StandardLogger().WriterLevel(logrus.TraceLevel)

	// Setup goroutine leak detector here so that it doesn't report the goroutines created by logrus.
	defer goleak.VerifyNone(tb, goleak.IgnoreCurrent())

	// Log the (temporary?) directory to store gluon data.
	logrus.Tracef("Gluon Data Dir: %v", options.dataDir)

	// Create a new gluon server.
	server, err := gluon.New(
		gluon.WithDataDir(options.dataDir),
		gluon.WithDelimiter(options.delimiter),
		gluon.WithTLS(&tls.Config{
			Certificates: []tls.Certificate{testCert},
			MinVersion:   tls.VersionTLS13,
		}),
		gluon.WithLogger(
			loggerIn,
			loggerOut,
		),
		gluon.WithVersionInfo(
			TestServerVersionInfo.Version.Major,
			TestServerVersionInfo.Version.Minor,
			TestServerVersionInfo.Version.Patch,
			TestServerVersionInfo.Name,
			TestServerVersionInfo.Vendor,
			TestServerVersionInfo.SupportURL,
		),
		gluon.WithStoreBuilder(&testBadgerStoreBuilder{}),
	)
	require.NoError(tb, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userIDs := make(map[string]string)
	conns := make(map[string]Connector)
	dbPaths := make(map[string]string)

	for _, creds := range options.credentials {
		conn := connector.NewDummy(
			creds.usernames,
			creds.password,
			defaultPeriod,
			defaultFlags,
			defaultPermanentFlags,
			defaultAttributes,
		)

		// Force USER ID to be consistent.
		hash := sha256.Sum256([]byte(creds.usernames[0]))
		userID := hex.EncodeToString(hash[:])

		err := server.LoadUser(ctx, conn, userID, []byte(creds.password))
		require.NoError(tb, err)

		require.NoError(tb, conn.Sync(ctx))

		for _, username := range creds.usernames {
			userIDs[username] = userID
		}

		conns[userID] = conn
		dbPaths[userID] = filepath.Join(server.GetDataPath(), "backend", "db", fmt.Sprintf("%v.db", userID))
	}

	listener, err := net.Listen("tcp", net.JoinHostPort("localhost", "0"))
	require.NoError(tb, err)

	errCh := server.Serve(ctx, listener)

	// Run the test against the server.
	tests(newTestSession(tb, listener, server, userIDs, conns, dbPaths, options))

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

func withData(s *testSession, username string, tests func(string, imap.LabelID)) {
	mbox := uuid.NewString()

	mboxID := s.mailboxCreated(username, []string{mbox}, "testdata/dovecot-crlf")

	tests(mbox, mboxID)
}

func getEntPath(dir string) string {
	return fmt.Sprintf("file:%v?cache=shared&_fk=1", filepath.Join(dir, fmt.Sprintf("%v.db", uuid.NewString())))
}
