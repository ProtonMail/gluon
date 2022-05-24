package tests

import (
	"testing"
)

func TestLoginSuccess(t *testing.T) {
	runOneToOneTest(t, "user", "pass", "/", func(c *testConnection, _ *testSession) {
		c.C("A001 login user pass").OK("A001")
	})
}

func TestLoginQuoted(t *testing.T) {
	runOneToOneTest(t, "user", "pass", "/", func(c *testConnection, _ *testSession) {
		c.C(`A001 login "user" "pass"`).OK("A001")
	})
}

func TestLoginLiteral(t *testing.T) {
	runOneToOneTest(t, "user", "pass", "/", func(c *testConnection, _ *testSession) {
		c.C(`A001 login {4}`)
		c.S(`+ (*_*)`)
		c.C(`user {4}`)
		c.S(`+ (*_*)`)
		c.C(`pass`)
		c.OK(`A001`)
	})
}

func TestLoginMultiple(t *testing.T) {
	runTest(t, map[string]string{"user1": "pass1", "user2": "pass2"}, "/", []int{1, 2}, func(c map[int]*testConnection, _ *testSession) {
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

func TestLoginFailure(t *testing.T) {
	runOneToOneTest(t, "user", "pass", "/", func(c *testConnection, _ *testSession) {
		c.C("A001 login baduser badpass").NO("A001")
	})
}

func TestLoginLiteralFailure(t *testing.T) {
	runOneToOneTest(t, "user", "pass", "/", func(c *testConnection, _ *testSession) {
		c.C(`A001 login {7}`)
		c.S(`+ (*_*)`)
		c.C(`baduser {7}`)
		c.S(`+ (*_*)`)
		c.C(`badpass`)
		c.NO(`A001`)
	})
}

func TestLoginCapabilities(t *testing.T) {
	runOneToOneTest(t, "user", "pass", "/", func(c *testConnection, _ *testSession) {
		c.C("A001 login user pass")
		c.S(`A001 OK [CAPABILITY IDLE IMAP4rev1 MOVE STARTTLS UIDPLUS UNSELECT] (^_^)`)
	})
}
