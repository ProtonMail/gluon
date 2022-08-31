package tests

import (
	"fmt"
	"testing"

	"github.com/ProtonMail/gluon/internal/errors"
	"github.com/ProtonMail/gluon/internal/session"
)

var (
	notAuthenticatedCommands = []string{
		`LOGIN user pass`,
		// `STARTTLS`, we allow upgrade to TLS after login.
		//`AUTHENTICATE GSSAPI`, not yet supported.
	}

	authenticatedCommands = []string{
		`SELECT INBOX`,
		`EXAMINE INBOX`,
		`CREATE mbox`,
		`DELETE mbox`,
		`RENAME mbox mbox2`,
		`SUBSCRIBE mbox`,
		`UNSUBSCRIBE mbox`,
		`LIST "" *`,
		`LSUB "#news." "comp.mail.*"`,
		`STATUS INBOX (messages)`,
		`IDLE`,
	}

	selectedCommands = []string{
		`CHECK`,
		`CLOSE`,
		`EXPUNGE`,
		`UID EXPUNGE 1`,
		`UNSELECT`,
		`SEARCH FLAGGED`,
		`FETCH 1 (UID)`,
		`STORE 1 +FLAGS (\Deleted)`,
		`COPY 1 INBOX`,
		`MOVE 1 INBOX`,
		`UID COPY 1 INBOX`,
		`UID MOVE 1 INBOX`,
	}
)

func TestErrorsWhenAuthenticated(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		for i, command := range notAuthenticatedCommands {
			c.C(fmt.Sprintf("%d %v", i, command))
			c.Sx(fmt.Sprintf("%d BAD %v", i, session.ErrAlreadyAuthenticated))
		}
	})
}

func TestErrorsWhenNotAuthenticated(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		for i, command := range append(authenticatedCommands, selectedCommands...) {
			c.C(fmt.Sprintf("%d %v", i, command))
			c.Sx(fmt.Sprintf("%d NO %v", i, session.ErrNotAuthenticated))
		}

		// Currently, the parser requires to read the message content before the error can be reported.
		c.C(`A001 APPEND INBOX {2}`)
		c.Sx(`\+ `)
		c.C(`12`)
		c.Sx(fmt.Sprintf("NO %v", session.ErrNotAuthenticated))
	})
}

func TestErrorsWhenNotSelected(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		for i, command := range selectedCommands {
			c.C(fmt.Sprintf("%d %v", i, command))
			c.Sx(fmt.Sprintf("%d NO %v", i, errors.ErrSessionNotSelected))
		}
	})
}
