package tests

import (
	"testing"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/events"
	"github.com/stretchr/testify/require"
)

func TestLoginSuccess(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("A001 login user pass").OK("A001")
	})
}

func TestLoginQuoted(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C(`A001 login "user" "pass"`).OK("A001")
	})
}

func TestLoginLiteral(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C(`A001 login {4}`)
		c.S(`+ Ready`)
		c.C(`user {4}`)
		c.S(`+ Ready`)
		c.C(`pass`)
		c.OK(`A001`)
	})
}

func TestLoginMultiple(t *testing.T) {
	runTest(t, defaultServerOptions(t, withCredentials([]credentials{
		{usernames: []string{"user1"}, password: "pass1"},
		{usernames: []string{"user2"}, password: "pass2"},
	})), []int{1, 2}, func(c map[int]*testConnection, _ *testSession) {
		// Login as the first user.
		c[1].C("A001 login user1 pass1").OK("A001")

		// Logout the first user.
		c[1].C("A002 logout").OK("A002")

		// Login as the second user.
		c[2].C("B001 login user2 pass2").OK("B001")

		// Logout the second user.
		c[2].C("B002 logout").OK("B002")
	})
}

func TestLoginAlias(t *testing.T) {
	runTest(t, defaultServerOptions(t, withCredentials([]credentials{{
		usernames: []string{"alias1", "alias2"},
		password:  "pass",
	}})), []int{1, 2}, func(c map[int]*testConnection, _ *testSession) {
		// Login as each alias.
		c[1].C("tag1 login alias1 pass").OK("tag1")
		c[2].C("tag2 login alias2 pass").OK("tag2")

		// Create a message with each alias.
		c[1].C("tag3 append inbox {11}\r\nTo: 1@pm.me").OK("tag3")
		c[2].C("tag4 append inbox {11}\r\nTo: 2@pm.me").OK("tag4")

		// Both messages should be visible to both clients.
		c[1].C("tag5 status inbox (messages)").Sx("MESSAGES 2").OK("tag5")
		c[2].C("tag6 status inbox (messages)").Sx("MESSAGES 2").OK("tag6")
	})
}

func TestLoginFailure(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("A001 login baduser badpass").NO("A001")
	})
}

func TestLoginLiteralFailure(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C(`A001 login {7}`)
		c.S(`+ Ready`)
		c.C(`baduser {7}`)
		c.S(`+ Ready`)
		c.C(`badpass`)
		c.NO(`A001`)
	})
}

func TestLoginCapabilities(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("A001 login user pass")
		c.S(`A001 OK [CAPABILITY IDLE IMAP4rev1 MOVE STARTTLS UIDPLUS UNSELECT] Logged in`)
	})
}

func TestLoginTooManyAttemps(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		// 3 attempts.
		c.C("A001 login user badpass").NO("A001")
		c.C("A001 login user badpass").NO("A001")
		c.C("A001 login user badpass").NO("A001")

		// The client should be jailed for 1 sec.
		require.Greater(t, timeFunc(func() {
			c.C("A001 login user badpass").NO("A001")
		}), time.Second)

		// After unjailed, get direct answer.
		require.Less(t, timeFunc(func() {
			c.C("A001 login user pass").OK("A001")
		}), time.Second)
	})
}

func TestLoginTooManyAttemptsMany(t *testing.T) {
	runManyToOneTest(t, defaultServerOptions(t), []int{1, 2, 3}, func(c map[int]*testConnection, s *testSession) {
		// 3 attempts.
		c[1].C("A001 login user badpass").NO("A001")
		c[2].C("A002 login user badpass").NO("A002")
		c[3].C("A003 login user badpass").NO("A003")

		wg := async.MakeWaitGroup(async.NoopPanicHandler{})

		// All clients should be jailed for 1 sec.
		for _, i := range []int{1, 2, 3} {
			i := i

			wg.Go(func() {
				require.Greater(t, timeFunc(func() {
					c[i].C("A001 login user badpass").NO("A001")
				}), time.Second)
			})
		}

		wg.Wait()
	})
}

func TestLoginEvents(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		require.IsType(t, events.UserAdded{}, <-s.eventCh)
		require.IsType(t, events.ListenerAdded{}, <-s.eventCh)
		require.IsType(t, events.SessionAdded{}, <-s.eventCh)

		c.C("A001 login baduser badpass").NO("A001")
		failedEvent, ok := (<-s.eventCh).(events.LoginFailed)
		require.True(t, ok)
		require.Equal(t, failedEvent.Username, "baduser")

		c.C("A002 login user pass").OK("A002")
		loginEvent, ok := (<-s.eventCh).(events.Login)
		require.True(t, ok)
		require.Equal(t, loginEvent.UserID, s.userIDs["user"])
	})
}

func timeFunc(fn func()) time.Duration {
	start := time.Now()

	fn()

	return time.Since(start)
}
