// Package tests defines "integration" tests of the library that run on top of a dummy remote.
package tests

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/emersion/go-imap/client"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

// runOneToOneTest runs a test with one account and one connection.
func runOneToOneTest(tb testing.TB, options *serverOptions, tests func(*testConnection, *testSession)) {
	runTest(tb, options, []int{1}, func(c map[int]*testConnection, s *testSession) {
		tests(c[1], s)
	})
}

// runOneToOneTestWithAuth runs a test with one account and one connection. The connection is logged in.
func runOneToOneTestWithAuth(tb testing.TB, options *serverOptions, tests func(*testConnection, *testSession)) {
	runOneToOneTest(tb, options, func(c *testConnection, s *testSession) {
		withTag(func(tag string) {
			c.Cf("%v login %v %v", tag, options.defaultUsername(), options.defaultUserPassword()).OK(tag)
		})

		tests(c, s)
	})
}

// runOneToOneTestWithData runs a test with one account and one connection. Apply mailbox is created with test data.
func runOneToOneTestWithData(tb testing.TB, options *serverOptions,
	tests func(*testConnection, *testSession, string, imap.LabelID),
) {
	runOneToOneTestWithAuth(tb, options, func(c *testConnection, s *testSession) {
		withData(s, options.defaultUsername(), func(mbox string, mboxID imap.LabelID) {
			withTag(func(tag string) { c.Cf("%v select %v", tag, mbox).OK(tag) })

			tests(c, s, mbox, mboxID)
		})
	})
}

// runManyToOneTest runs a test with one account and multiple connections.
func runManyToOneTest(tb testing.TB, options *serverOptions, connIDs []int, tests func(map[int]*testConnection, *testSession)) {
	runTest(tb, options, connIDs, func(c map[int]*testConnection, s *testSession) {
		tests(c, s)
	})
}

// runManyToOneTestWithAuth runs a test with one account and multiple connections. Each connection is logged in.
func runManyToOneTestWithAuth(tb testing.TB, options *serverOptions,
	connIDs []int, tests func(map[int]*testConnection, *testSession),
) {
	runManyToOneTest(tb, options, connIDs, func(c map[int]*testConnection, s *testSession) {
		for _, c := range c {
			withTag(func(tag string) {
				c.Cf("%v login %v %v", tag, options.defaultUsername(), options.defaultUserPassword()).OK(tag)
			})
		}

		tests(c, s)
	})
}

// runManyToOneTestWithData runs a test with one account and multiple connections. Apply mailbox is created with test data.
func runManyToOneTestWithData(tb testing.TB, options *serverOptions,
	connIDs []int, tests func(map[int]*testConnection, *testSession, string, imap.LabelID),
) {
	runManyToOneTestWithAuth(tb, options, connIDs, func(c map[int]*testConnection, s *testSession) {
		withData(s, options.defaultUsername(), func(mbox string, mboxID imap.LabelID) {
			for _, c := range c {
				withTag(func(tag string) { c.Cf("%v select %v", tag, mbox).OK(tag) })
			}

			tests(c, s, mbox, mboxID)
		})
	})
}

// runTest runs the mailserver and creates test connections to it.
func runTest(tb testing.TB, options *serverOptions, connIDs []int, tests func(map[int]*testConnection, *testSession)) {
	runServer(tb, options, func(s *testSession) {
		withConnections(tb, s, connIDs, func(c map[int]*testConnection) {
			tests(c, s)
		})
	})
}

// -- IMAP client test helpers

// runTestClient runs the mailserver and creates test connections to it using an imap client.
func runTestClient(tb testing.TB, options *serverOptions, connIDs []int, tests func(map[int]*client.Client, *testSession)) {
	runServer(tb, options, func(s *testSession) {
		withClients(tb, s, connIDs, func(clientMap map[int]*client.Client) {
			tests(clientMap, s)
		})
	})
}

// runOneToOneTestClient runs a test with one account and one connection using an imap client.
func runOneToOneTestClient(tb testing.TB, options *serverOptions, test func(*client.Client, *testSession)) {
	runTestClient(tb, options, []int{1}, func(c map[int]*client.Client, s *testSession) {
		test(c[1], s)
	})
}

// runOneToOneTestClientWithAuth runs a test with one account and one connection using an imap client. The connection is logged in.
func runOneToOneTestClientWithAuth(tb testing.TB, options *serverOptions, test func(*client.Client, *testSession)) {
	runOneToOneTestClient(tb, options, func(client *client.Client, s *testSession) {
		require.NoError(tb, client.Login(options.defaultUsername(), options.defaultUserPassword()))
		test(client, s)
	})
}

// runOneToOneTestClientWithData runs a test with one account and one connection using an imap client. Apply mailbox is created with test data.
func runOneToOneTestClientWithData(tb testing.TB, options *serverOptions, test func(*client.Client, *testSession, string, imap.LabelID)) {
	runOneToOneTestClientWithAuth(tb, options, func(client *client.Client, s *testSession) {
		withData(s, options.defaultUsername(), func(mbox string, mboxID imap.LabelID) {
			_, err := client.Select(mbox, false)
			require.NoError(s.tb, err)
			test(client, s, mbox, mboxID)
		})
	})
}
