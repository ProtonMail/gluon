package tests

import (
	"testing"
)

func TestStatus(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, _ *testSession) {
		c.C("B001 CREATE blurdybloop")
		c.S("B001 OK CREATE")

		c.doAppend(`blurdybloop`, `To: 1@pm.me`, `\Seen`).expect("OK")
		c.doAppend(`blurdybloop`, `To: 2@pm.me`).expect("OK")
		c.doAppend(`blurdybloop`, `To: 3@pm.me`).expect("OK")

		c.C("A042 STATUS blurdybloop (MESSAGES UNSEEN)")
		c.S(`* STATUS "blurdybloop" (MESSAGES 3 UNSEEN 2)`)
		c.S("A042 OK STATUS")
	})
}

func TestStatusWithUtf8MailboxNames(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withDelimiter(".")), func(c *testConnection, s *testSession) {
		s.mailboxCreated("user", []string{"mbox-öüäëæøå"})
		s.flush("user")
		c.doAppend(`mbox-&APYA,ADkAOsA5gD4AOU-`, `To: 1@pm.me`).expect("OK")
		c.C(`a STATUS mbox-&APYA,ADkAOsA5gD4AOU- (MESSAGES)`)
		c.S(`* STATUS "mbox-&APYA,ADkAOsA5gD4AOU-" (MESSAGES 1)`)
		c.OK(`a`)
	})
}
