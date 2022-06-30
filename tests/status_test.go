package tests

import (
	"testing"
)

func TestStatus(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, _ *testSession) {
		c.C("B001 CREATE blurdybloop")
		c.S("B001 OK (^_^)")

		c.doAppend(`blurdybloop`, `To: 1@pm.me`, `\Seen`).expect("OK")
		c.doAppend(`blurdybloop`, `To: 2@pm.me`).expect("OK")
		c.doAppend(`blurdybloop`, `To: 3@pm.me`).expect("OK")

		c.C("A042 STATUS blurdybloop (MESSAGES UNSEEN)")
		c.S(`* STATUS "blurdybloop" (MESSAGES 3 UNSEEN 2)`)
		c.S("A042 OK (^_^)")
	})
}
