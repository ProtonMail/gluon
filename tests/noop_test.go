package tests

import (
	"testing"
)

func TestNoop(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("a001 noop")
		c.S("a001 OK (^_^)")
	})
}
