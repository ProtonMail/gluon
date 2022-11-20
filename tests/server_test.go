package tests

import (
	"context"
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
	"github.com/ProtonMail/gluon/internal/hash"
	"github.com/ProtonMail/gluon/logging"
	"github.com/ProtonMail/gluon/store"
	"github.com/ProtonMail/gluon/version"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-imap/client"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
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

var testServerVersionInfo = version.Info{
	Name:       "gluon-test-server",
	Version:    version.Version{Major: 1, Minor: 1, Patch: 1},
	Vendor:     "Proton",
	SupportURL: "",
}

type connectorBuilder interface {
	New(usernames []string, password []byte, period time.Duration, flags, permFlags, attrs imap.FlagSet) Connector
}

type dummyConnectorBuilder struct{}

func (*dummyConnectorBuilder) New(usernames []string, password []byte, period time.Duration, flags, permFlags, attrs imap.FlagSet) Connector {
	return connector.NewDummy(
		usernames,
		password,
		period,
		flags,
		permFlags,
		attrs,
	)
}

type serverOptions struct {
	credentials        []credentials
	delimiter          string
	loginJailTime      time.Duration
	dataDir            string
	idleBulkTime       time.Duration
	storeBuilder       store.Builder
	connectorBuilder   connectorBuilder
	disableParallelism bool
}

func (s *serverOptions) defaultUsername() string {
	return s.credentials[0].usernames[0]
}

func (s *serverOptions) defaultPassword() string {
	return s.credentials[0].password
}

func (s *serverOptions) password(username string) string {
	return s.credentials[xslices.IndexFunc(s.credentials, func(cred credentials) bool {
		return slices.Contains(cred.usernames, username)
	})].password
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

type idleBulkTimeOption struct {
	idleBulkTime time.Duration
}

func (d *idleBulkTimeOption) apply(options *serverOptions) {
	options.idleBulkTime = d.idleBulkTime
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

type storeBuilderOption struct {
	builder store.Builder
}

func (s *storeBuilderOption) apply(options *serverOptions) {
	options.storeBuilder = s.builder
}

type connectorBuilderOption struct {
	builder connectorBuilder
}

func (c *connectorBuilderOption) apply(options *serverOptions) {
	options.connectorBuilder = c.builder
}

type disableParallelism struct{}

func (disableParallelism) apply(options *serverOptions) {
	options.disableParallelism = true
}

func withIdleBulkTime(idleBulkTime time.Duration) serverOption {
	return &idleBulkTimeOption{idleBulkTime: idleBulkTime}
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

func withStoreBuilder(builder store.Builder) serverOption {
	return &storeBuilderOption{builder: builder}
}

func withConnectorBuilder(builder connectorBuilder) serverOption {
	return &connectorBuilderOption{builder: builder}
}

func withDisableParallelism() serverOption {
	return &disableParallelism{}
}

func defaultServerOptions(tb testing.TB, modifiers ...serverOption) *serverOptions {
	options := &serverOptions{
		credentials: []credentials{{
			usernames: []string{"user"},
			password:  "pass",
		}},
		delimiter:        "/",
		loginJailTime:    time.Second,
		dataDir:          tb.TempDir(),
		idleBulkTime:     time.Duration(500 * time.Millisecond),
		storeBuilder:     &store.OnDiskStoreBuilder{},
		connectorBuilder: &dummyConnectorBuilder{},
	}

	for _, op := range modifiers {
		op.apply(options)
	}

	return options
}

// runServer initializes and starts the mailserver.
func runServer(tb testing.TB, options *serverOptions, tests func(session *testSession)) {
	loggerIn := logrus.StandardLogger().WriterLevel(logrus.TraceLevel)
	defer loggerIn.Close()

	loggerOut := logrus.StandardLogger().WriterLevel(logrus.TraceLevel)
	defer loggerOut.Close()

	// Log the (temporary?) directory to store gluon data.
	logrus.Tracef("Gluon Data Dir: %v", options.dataDir)

	gluonOptions := []gluon.Option{
		gluon.WithDataDir(options.dataDir),
		gluon.WithDelimiter(options.delimiter),
		gluon.WithLoginJailTime(options.loginJailTime),
		gluon.WithTLS(&tls.Config{
			Certificates: []tls.Certificate{testCert},
			MinVersion:   tls.VersionTLS13,
		}),
		gluon.WithLogger(
			loggerIn,
			loggerOut,
		),
		gluon.WithVersionInfo(
			testServerVersionInfo.Version.Major,
			testServerVersionInfo.Version.Minor,
			testServerVersionInfo.Version.Patch,
			testServerVersionInfo.Name,
			testServerVersionInfo.Vendor,
			testServerVersionInfo.SupportURL,
		),
		gluon.WithIdleBulkTime(options.idleBulkTime),
		gluon.WithStoreBuilder(options.storeBuilder),
	}

	if options.disableParallelism {
		gluonOptions = append(gluonOptions, gluon.WithDisableParallelism())
	}

	// Create a new gluon server.
	server, err := gluon.New(gluonOptions...)
	require.NoError(tb, err)

	// Watch server events.
	eventCh := server.AddWatcher()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userIDs := make(map[string]string)
	conns := make(map[string]Connector)
	dbPaths := make(map[string]string)

	for _, creds := range options.credentials {
		conn := options.connectorBuilder.New(
			creds.usernames,
			[]byte(creds.password),
			defaultPeriod,
			defaultFlags,
			defaultPermanentFlags,
			defaultAttributes,
		)

		// Force USER ID to be consistent.
		userID := hex.EncodeToString(hash.SHA256([]byte(creds.usernames[0])))

		// Load the user.
		require.NoError(tb, server.LoadUser(ctx, conn, userID, []byte(creds.password)))

		// Trigger a sync of the user's data.
		require.NoError(tb, conn.Sync(ctx))

		for _, username := range creds.usernames {
			userIDs[username] = userID
		}

		conns[userID] = conn
		dbPaths[userID] = filepath.Join(server.GetDataPath(), "backend", "db", fmt.Sprintf("%v.db", userID))
	}

	listener, err := net.Listen("tcp", net.JoinHostPort("localhost", "0"))
	require.NoError(tb, err)

	// Start the server.
	require.NoError(tb, server.Serve(ctx, listener))

	// Run the test against the server.
	logging.DoAnnotated(ctx, func(context.Context) {
		tests(newTestSession(tb, listener, server, eventCh, userIDs, conns, dbPaths, options))
	}, logging.Labels{
		"Action": "Running gluon tests",
	})

	// Flush and remove user before shutdown.
	for userID, conn := range conns {
		conn.Flush()
		require.NoError(tb, server.RemoveUser(ctx, userID, false))
	}

	// Expect the server to shut down successfully when closed.
	require.NoError(tb, server.Close(ctx))
	require.NoError(tb, <-server.GetErrorCh())
	require.NoError(tb, listener.Close())
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

func withData(s *testSession, username string, tests func(string, imap.MailboxID)) {
	mbox := uuid.NewString()

	mboxID := s.mailboxCreated(username, []string{mbox}, "testdata/dovecot-crlf")

	tests(mbox, mboxID)
}
