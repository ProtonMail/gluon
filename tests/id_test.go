package tests

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/stretchr/testify/require"
)

func TestIdNilNoData(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		idResponse := response.ID(imap.NewIMAPIDFromVersionInfo(&TestServerVersionInfo))
		c.C(`A001 ID NIL`)
		c.S(idResponse.String())
		c.OK(`A001`)
	})
}

func TestIdContextLookup(t *testing.T) {
	runOneToOneTest(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		idResponse := response.ID(imap.NewIMAPIDFromVersionInfo(&TestServerVersionInfo))
		// Store new ID
		c.C(`A001 ID ("foo" "bar")`)
		c.S(idResponse.String())
		c.OK(`A001`)

		c.C("A003 LOGIN user pass").OK("A003")

		// NOTE: We are only recording this with APPEND since it was the easiest command to verify the data has been
		// record properly in the context, as APPEND will always require a communication with the remote connector.
		c.C("A004 APPEND INBOX (\\Seen) {26}")
		c.S("+ Ready")
		c.C("To: 00010203-0405-4607-880").OK("A004")

		s.flush("user")

		expectedID := imap.NewID()
		expectedID.Other["foo"] = "bar"

		require.Equal(t, expectedID, s.conns[s.userIDs["user"]].GetLastRecordedIMAPID())
	})
}
