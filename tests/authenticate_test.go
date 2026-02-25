package tests

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/gluon/events"
	"github.com/stretchr/testify/require"
)

func base64AuthString(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("\x00%v\x00%v", username, password)))
}

func TestAuthenticateSuccess(t *testing.T) {
	authString := base64AuthString("user", "pass")

	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("A001 authenticate PLAIN")
		c.S("+")
		c.C(authString).OK("A001")
	})
}

func TestAuthenticateDisabled(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t, withDisableIMAPAuthenticate()), func(c *testConnection, _ *testSession) {
		c.C("A001 AUTHENTICATE PLAIN")
		c.BAD("A001")
	})
}

func TestAuthenticateFailure(t *testing.T) {
	authString := base64AuthString("user", "badPass")

	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("A001 AUTHENTICATE PLAIN")
		c.S("+")
		c.C(authString).NO("A001")
	})
}

func TestAuthenticateMultiple(t *testing.T) {
	authString1 := base64AuthString("user1", "pass1")
	authString2 := base64AuthString("user2", "pass2")

	runTest(t, defaultServerOptions(t, withCredentials([]credentials{
		{usernames: []string{"user1"}, password: "pass1"},
		{usernames: []string{"user2"}, password: "pass2"},
	})), []int{1, 2}, func(c map[int]*testConnection, _ *testSession) {
		// Login as the first user.
		c[1].C("A001 AUTHENTICATE plain")
		c[1].S("+")
		c[1].C(authString1).OK("A001")

		// Logout the first user.
		c[1].C("A002 logout").OK("A002")

		// Login as the second user.
		c[2].C("B001 AUTHENTICATE plain")
		c[2].S("+")
		c[2].C(authString2).OK("B001")

		// Logout the second user.
		c[2].C("B002 logout").OK("B002")
	})
}

func TestAuthenticateCapabilities(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("A001 AUTHENTICATE PLAIN")
		c.S("+")
		c.C(base64AuthString("user", "pass"))
		c.S(`A001 OK [CAPABILITY AUTH=PLAIN ID IDLE IMAP4rev1 MOVE QUOTA STARTTLS UIDPLUS UNSELECT] Logged in`)
	})
}

func TestAuthenticateTooManyAttemptsMany(t *testing.T) {
	runManyToOneTest(t, defaultServerOptions(t), []int{1, 2, 3}, func(c map[int]*testConnection, s *testSession) {
		authString := base64AuthString("user", "badPass")

		// 3 attempts.
		for _, i := range []int{1, 2, 3} {
			tag := fmt.Sprintf("A%03d", i)
			c[1].C(tag + " AUTHENTICATE PLAIN")
			c[1].S("+")
			c[1].C(authString).NO(tag)
		}

		wg := async.MakeWaitGroup(async.NoopPanicHandler{})

		// All clients should be jailed for 1 sec.
		for _, i := range []int{1, 2, 3} {
			i := i
			tag := fmt.Sprintf("A%03d", i)

			wg.Go(func() {
				require.Greater(t, timeFunc(func() {
					c[i].C(tag + " AUTHENTICATE PLAIN")
					c[i].S("+")
					c[i].C(authString).NO(tag)
				}), 990*time.Millisecond)
			})
		}

		wg.Wait()
	})
}

func TestAuthenticateEvents(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		require.IsType(t, events.UserAdded{}, <-s.eventCh)
		require.IsType(t, events.ListenerAdded{}, <-s.eventCh)
		require.IsType(t, events.SessionAdded{}, <-s.eventCh)

		c.C("A001 authenticate PLAIN")
		c.S("+")
		c.C(base64AuthString("badUser", "badPass")).NO("A001")

		failedEvent, ok := (<-s.eventCh).(events.LoginFailed)
		require.True(t, ok)
		require.Equal(t, "badUser", failedEvent.Username)

		c.C("A002 authenticate plain")
		c.S("+")
		c.C(base64AuthString("user", "pass")).OK("A002")

		loginEvent, ok := (<-s.eventCh).(events.Login)
		require.True(t, ok)
		require.Equal(t, s.userIDs["user"], loginEvent.UserID)
	})
}
