package tests

import (
	"testing"
)

func TestMultiUser(t *testing.T) {
	runTest(t, defaultServerOptions(t, withCredentials([]credentials{
		{usernames: []string{"user1"}, password: "pass"},
		{usernames: []string{"user2"}, password: "pass"},
	})), []int{1, 2}, func(c map[int]*testConnection, s *testSession) {
		c[1].C(`A001 login user1 pass`).OK(`A001`)
		c[2].C(`B001 login user2 pass`).OK(`B001`)

		c[1].C(`A002 select inbox`).OK(`A002`)
		c[2].C(`B002 select inbox`).OK(`B002`)
	})
}
