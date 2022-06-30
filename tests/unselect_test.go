package tests

import (
	"testing"
)

func TestUnselect(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("b001 CREATE saved-messages")
		c.S("b001 OK (^_^)")

		c.C(`A002 SELECT INBOX`)
		c.Se(`A002 OK [READ-WRITE] (^_^)`)

		c.C(`A202 UNSELECT`)
		c.S("A202 OK (^_^)")
	})
}
