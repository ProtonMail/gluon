package tests

import (
	"testing"
)

func TestCapability(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("A001 Capability")
		c.S(`* CAPABILITY AUTH=PLAIN ID IDLE IMAP4rev1 STARTTLS`)
		c.S("A001 OK CAPABILITY")

		c.C(`A002 login "user" "pass"`)
		c.S(`A002 OK [CAPABILITY AUTH=PLAIN ID IDLE IMAP4rev1 MOVE STARTTLS UIDPLUS UNSELECT] Logged in`)

		c.C("A003 Capability")
		c.S(`* CAPABILITY AUTH=PLAIN ID IDLE IMAP4rev1 MOVE STARTTLS UIDPLUS UNSELECT`)
		c.S("A003 OK CAPABILITY")
	})
}

func TestCapabilityAuthenticateDisabled(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t, withDisableIMAPAuthenticate()), func(c *testConnection, _ *testSession) {
		c.C("A001 Capability")
		c.S(`* CAPABILITY ID IDLE IMAP4rev1 STARTTLS`)
		c.S("A001 OK CAPABILITY")

		c.C(`A002 login "user" "pass"`)
		c.S(`A002 OK [CAPABILITY ID IDLE IMAP4rev1 MOVE STARTTLS UIDPLUS UNSELECT] Logged in`)

		c.C("A003 Capability")
		c.S(`* CAPABILITY ID IDLE IMAP4rev1 MOVE STARTTLS UIDPLUS UNSELECT`)
		c.S("A003 OK CAPABILITY")
	})
}

// With the Gmail extension enabled, X-GM-EXT-1 is advertised (both before and after auth).
func TestCapabilityGmailExtensionEnabled(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t, withGmailExtension()), func(c *testConnection, _ *testSession) {
		c.C("A001 Capability")
		c.S(`* CAPABILITY AUTH=PLAIN ID IDLE IMAP4rev1 STARTTLS X-GM-EXT-1`)
		c.S("A001 OK CAPABILITY")

		c.C(`A002 login "user" "pass"`)
		c.S(`A002 OK [CAPABILITY AUTH=PLAIN ID IDLE IMAP4rev1 MOVE STARTTLS UIDPLUS UNSELECT X-GM-EXT-1] Logged in`)

		c.C("A003 Capability")
		c.S(`* CAPABILITY AUTH=PLAIN ID IDLE IMAP4rev1 MOVE STARTTLS UIDPLUS UNSELECT X-GM-EXT-1`)
		c.S("A003 OK CAPABILITY")
	})
}
