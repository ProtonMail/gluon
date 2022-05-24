package tests

import (
	"testing"
)

func TestLogout(t *testing.T) {
	runOneToOneTest(t, "user", "pass", "/", func(c *testConnection, _ *testSession) {
		c.C("a001 logout")
		c.S("* BYE (^_^)/~")
		c.S("a001 OK (^_^)")
		c.expectClosed()
	})
}
