package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStartTLS(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		c.C("A001 starttls")
		c.S("A001 OK Begin TLS negotiation now")

		c.upgradeConnection()

		c.C("A002 noop")
		c.OK("A002")
	})
}

func TestAutoUpgradeTLS(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(_ *testConnection, s *testSession) {
		c := s.newClientTLS()

		require.NoError(t, c.Login(s.options.defaultUsername(), s.options.defaultPassword()))
	})
}
