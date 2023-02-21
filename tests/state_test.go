package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/ProtonMail/gluon"
	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/session"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/stretchr/testify/require"
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
			c.Sx(fmt.Sprintf("%d NO %v", i, state.ErrSessionNotSelected))
		}
	})
}

func TestErrNoSuchMessage(t *testing.T) {
	connBuilder := &updateInjectorConnectorBuilder{
		updateCh: make(chan imap.Update),
	}

	runOneToOneTestWithAuth(t, defaultServerOptions(t, withConnectorBuilder(connBuilder)), func(c *testConnection, s *testSession) {
		update := imap.NewMessageMailboxesUpdated("this is not the message you are looking for", []imap.MailboxID{}, false, false, false)

		connBuilder.updateCh <- update

		err, ok := update.Wait()
		require.True(t, ok)
		require.True(t, gluon.IsNoSuchMessage(err))
	})
}

type updateInjectorConnector struct {
	*connector.Dummy

	updateCh chan imap.Update
}

func (conn *updateInjectorConnector) GetUpdates() <-chan imap.Update {
	return conn.updateCh
}

type updateInjectorConnectorBuilder struct {
	updateCh chan imap.Update
}

func (builder updateInjectorConnectorBuilder) New(usernames []string, password []byte, period time.Duration, flags, permFlags, attrs imap.FlagSet) Connector {
	dummy := connector.NewDummy(usernames, password, period, flags, permFlags, attrs)

	go func() {
		for event := range dummy.GetUpdates() {
			builder.updateCh <- event
		}
	}()

	return &updateInjectorConnector{
		Dummy:    dummy,
		updateCh: builder.updateCh,
	}
}
