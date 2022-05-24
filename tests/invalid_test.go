package tests

import (
	"testing"
)

func TestInvalidIMAPCommandBadTag(t *testing.T) {
	runOneToOneTestWithAuth(t, "user", "pass", "/", func(c *testConnection, s *testSession) {
		c.C("A006 RANDOMGIBBERISHTHATDOESNOTMAKEAVALIDIMAPCOMMAND").BAD("A006")
	})
}
