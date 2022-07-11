package tests

import "testing"

func TestStartTLS(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		c.C("A001 starttls")
		c.S("A001 OK Begin TLS negotiation now")

		c.upgradeConnection()

		c.C("A002 noop")
		c.OK("A002")
	})
}
