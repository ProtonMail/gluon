package tests

import (
	"testing"
)

func TestCapability(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("A001 Capability")
		c.S(`* CAPABILITY IDLE IMAP4rev1 STARTTLS`)
		c.S("A001 OK CAPABILITY")

		c.C(`A002 login "user" "pass"`)
		c.S(`A002 OK [CAPABILITY IDLE IMAP4rev1 MOVE STARTTLS UIDPLUS UNSELECT] Logged in`)

		c.C("A003 Capability")
		c.S(`* CAPABILITY IDLE IMAP4rev1 MOVE STARTTLS UIDPLUS UNSELECT`)
		c.S("A003 OK CAPABILITY")
	})
}
