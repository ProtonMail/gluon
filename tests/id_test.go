package tests

import (
	"fmt"
	"testing"

	"github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/stretchr/testify/require"
)

func TestIdNilNoData(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		wantResponse := response.ID(imap.NewIMAPIDFromVersionInfo(testServerVersionInfo))
		c.C(`A001 ID NIL`)
		c.S(wantResponse.String())
		c.OK(`A001`)
	})
}

func TestIdContextLookup(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		wantResponse := response.ID(imap.NewIMAPIDFromVersionInfo(testServerVersionInfo))
		// Store new ID
		c.C(`A001 ID ("foo" "bar")`)
		c.S(wantResponse.String())
		c.OK(`A001`)

		c.C("A003 LOGIN user pass").OK("A003")

		// NOTE: We are only recording this with APPEND since it was the easiest command to verify the data has been
		// record properly in the context, as APPEND will always require a communication with the remote connector.
		literal := buildRFC5322TestLiteral("To: 00010203-0405-4607-880")
		literalLen := len(literal)
		c.C(fmt.Sprintf("A004 APPEND INBOX (\\Seen) {%v}", literalLen))
		c.S("+ Ready")
		c.C(literal).OK("A004")

		s.flush("user")

		wantID := imap.NewIMAPIDFromKeyMap(map[string]string{"foo": "bar"})

		require.Equal(t, wantID, s.conns[s.userIDs["user"]].GetLastRecordedIMAPID())
	})
}

func TestIdEvent(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		c.C(`A001 ID ("name" "foo" "version" "bar")`).OK(`A001`)

		imapID := getEvent[events.IMAPID](s.eventCh).IMAPID
		require.Equal(t, "foo", imapID.Name)
		require.Equal(t, "bar", imapID.Version)
	})
}
