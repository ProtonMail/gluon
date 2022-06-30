package tests

import (
	"testing"
)

func TestInvalidIMAPCommandBadTag(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		c.C("A006 RANDOMGIBBERISHTHATDOESNOTMAKEAVALIDIMAPCOMMAND").BAD("A006")
	})
}
