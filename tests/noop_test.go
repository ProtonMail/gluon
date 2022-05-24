package tests

import (
	"testing"
)

func TestNoop(t *testing.T) {
	runOneToOneTest(t, "user", "pass", "/", func(c *testConnection, _ *testSession) {
		c.C("a001 noop")
		c.S("a001 OK (^_^)")
	})
}
