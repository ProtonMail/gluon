// Package tests defines "integration" tests of the library that run on top of a dummy remote.
package tests

import (
	"testing"

	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// runOneToOneTest runs a test with one account and one connection.
func runOneToOneTest(tb testing.TB, username, password, delimiter string, tests func(*testConnection, *testSession)) {
	runTest(tb, []credentials{{
		usernames: []string{username},
		password:  password,
	}}, delimiter, []int{1}, func(c map[int]*testConnection, s *testSession) {
		tests(c[1], s)
	})
}

// runOneToOneTestWithAuth runs a test with one account and one connection. The connection is logged in.
func runOneToOneTestWithAuth(tb testing.TB, username, password, delimiter string, tests func(*testConnection, *testSession)) {
	runOneToOneTest(tb, username, password, delimiter, func(c *testConnection, s *testSession) {
		withTag(func(tag string) { c.Cf("%v login %v %v", tag, username, password).OK(tag) })

		tests(c, s)
	})
}

// runOneToOneTestWithData runs a test with one account and one connection. A mailbox is created with test data.
func runOneToOneTestWithData(tb testing.TB, username, password, delimiter string, tests func(*testConnection, *testSession, string, string)) {
	runOneToOneTestWithAuth(tb, username, password, delimiter, func(c *testConnection, s *testSession) {
		withData(s, username, func(mbox, mboxID string) {
			withTag(func(tag string) { c.Cf("%v select %v", tag, mbox).OK(tag) })

			tests(c, s, mbox, mboxID)
		})
	})
}

// runManyToOneTest runs a test with one account and multiple connections.
func runManyToOneTest(tb testing.TB, username, password, delimiter string, connIDs []int, tests func(map[int]*testConnection, *testSession)) {
	runTest(tb, []credentials{{
		usernames: []string{username},
		password:  password,
	}}, delimiter, connIDs, func(c map[int]*testConnection, s *testSession) {
		tests(c, s)
	})
}

// runManyToOneTestWithAuth runs a test with one account and multiple connections. Each connection is logged in.
func runManyToOneTestWithAuth(tb testing.TB, username, password, delimiter string, connIDs []int, tests func(map[int]*testConnection, *testSession)) {
	runManyToOneTest(tb, username, password, delimiter, connIDs, func(c map[int]*testConnection, s *testSession) {
		for _, c := range c {
			withTag(func(tag string) { c.Cf("%v login %v %v", tag, username, password).OK(tag) })
		}

		tests(c, s)
	})
}

// runManyToOneTestWithData runs a test with one account and multiple connections. A mailbox is created with test data.
func runManyToOneTestWithData(tb testing.TB, username, password, delimiter string, connIDs []int, tests func(map[int]*testConnection, *testSession, string, string)) {
	runManyToOneTestWithAuth(tb, username, password, delimiter, connIDs, func(c map[int]*testConnection, s *testSession) {
		withData(s, username, func(mbox, mboxID string) {
			for _, c := range c {
				withTag(func(tag string) { c.Cf("%v select %v", tag, mbox).OK(tag) })
			}

			tests(c, s, mbox, mboxID)
		})
	})
}

// runTest runs the mailserver and creates test connections to it.
func runTest(tb testing.TB, creds []credentials, delimiter string, connIDs []int, tests func(map[int]*testConnection, *testSession)) {
	runServer(tb, creds, delimiter, func(s *testSession) {
		withConnections(tb, s, connIDs, func(c map[int]*testConnection) {
			tests(c, s)
		})
	})
}

// -- IMap client test helpers

// runTestClient runs the mailserver and creates test connections to it using an imap client.
func runTestClient(tb testing.TB, creds []credentials, delimiter string, connIDs []int, tests func(map[int]*client.Client, *testSession)) {
	runServer(tb, creds, delimiter, func(s *testSession) {
		withClients(tb, s, connIDs, func(clientMap map[int]*client.Client) {
			tests(clientMap, s)
		})
	})
}

// runOneToOneTestClient runs a test with one account and one connection using an imap client.
func runOneToOneTestClient(tb testing.TB, username, password, delimiter string, test func(*client.Client, *testSession)) {
	runTestClient(tb, []credentials{{
		usernames: []string{username},
		password:  password,
	}}, delimiter, []int{1}, func(c map[int]*client.Client, s *testSession) {
		test(c[1], s)
	})
}

// runOneToOneTestClientWithAuth runs a test with one account and one connection using an imap client. The connection is logged in.
func runOneToOneTestClientWithAuth(tb testing.TB, username, password, delimiter string, test func(*client.Client, *testSession)) {
	runOneToOneTestClient(tb, username, password, delimiter, func(client *client.Client, s *testSession) {
		require.NoError(tb, client.Login(username, password))
		test(client, s)
	})
}

// runOneToOneTestClientWithData runs a test with one account and one connection using an imap client. A mailbox is created with test data.
func runOneToOneTestClientWithData(tb testing.TB, username, password, delimiter string, test func(*client.Client, *testSession, string, string)) {
	runOneToOneTestClientWithAuth(tb, username, password, delimiter, func(client *client.Client, s *testSession) {
		withData(s, username, func(mbox, mboxID string) {
			_, err := client.Select(mbox, false)
			require.NoError(s.tb, err)
			test(client, s, mbox, mboxID)
		})
	})
}
