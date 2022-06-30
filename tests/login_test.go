package tests

import (
	"testing"
)

func TestLoginSuccess(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("A001 login user pass").OK("A001")
	})
}

func TestLoginQuoted(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C(`A001 login "user" "pass"`).OK("A001")
	})
}

func TestLoginLiteral(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C(`A001 login {4}`)
		c.S(`+ (*_*)`)
		c.C(`user {4}`)
		c.S(`+ (*_*)`)
		c.C(`pass`)
		c.OK(`A001`)
	})
}

func TestLoginMultiple(t *testing.T) {
	runTest(t, defaultServerOptions(t, withCredentials([]credentials{
		{usernames: []string{"user1"}, password: "pass1"},
		{usernames: []string{"user2"}, password: "pass2"},
	})), []int{1, 2}, func(c map[int]*testConnection, _ *testSession) {
		// Login as the first user.
		c[1].C("A001 login user1 pass1").OK("A001")

		// Logout the first user.
		c[1].C("A002 logout").OK("A002")

		// Login as the second user.
		c[2].C("B001 login user2 pass2").OK("B001")

		// Logout the second user.
		c[2].C("B002 logout").OK("B002")
	})
}

func TestLoginAlias(t *testing.T) {
	runTest(t, defaultServerOptions(t, withCredentials([]credentials{{
		usernames: []string{"alias1", "alias2"},
		password:  "pass",
	}})), []int{1, 2}, func(c map[int]*testConnection, _ *testSession) {
		// Login as each alias.
		c[1].C("tag1 login alias1 pass").OK("tag1")
		c[2].C("tag2 login alias2 pass").OK("tag2")

		// Create a message with each alias.
		c[1].C("tag3 append inbox {11}\r\nTo: 1@pm.me").OK("tag3")
		c[2].C("tag4 append inbox {11}\r\nTo: 2@pm.me").OK("tag4")

		// Both messages should be visible to both clients.
		c[1].C("tag5 status inbox (messages)").Sx("MESSAGES 2").OK("tag5")
		c[2].C("tag6 status inbox (messages)").Sx("MESSAGES 2").OK("tag6")
	})
}

func TestLoginFailure(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("A001 login baduser badpass").NO("A001")
	})
}

func TestLoginLiteralFailure(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C(`A001 login {7}`)
		c.S(`+ (*_*)`)
		c.C(`baduser {7}`)
		c.S(`+ (*_*)`)
		c.C(`badpass`)
		c.NO(`A001`)
	})
}

func TestLoginCapabilities(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("A001 login user pass")
		c.S(`A001 OK [CAPABILITY IDLE IMAP4rev1 MOVE STARTTLS UIDPLUS UNSELECT] (^_^)`)
	})
}
