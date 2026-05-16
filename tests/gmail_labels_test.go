package tests

import (
	"testing"
)

// These integration tests exercise the X-GM-EXT-1 Gmail label extension
// end-to-end (STORE / FETCH / SEARCH) through the dummy connector, which now
// stores labels as non-exclusive label mailboxes. This is the path Paperless-NGX
// uses to tag mail over Proton Bridge IMAP.

func TestGmailLabelsStoreFetchRoundTrip(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("b001 CREATE saved-messages")
		c.S("b001 OK CREATE")

		c.doAppend(`saved-messages`, buildRFC5322TestLiteral(`To: 1@pm.me`)).expect("OK")

		c.C(`A001 SELECT saved-messages`)
		c.Se(`A001 OK [READ-WRITE] SELECT`)

		// No labels initially.
		c.C(`A002 FETCH 1 (X-GM-LABELS)`)
		c.S(`* 1 FETCH (X-GM-LABELS ())`)
		c.OK(`A002`)

		// Apply two labels. STORE X-GM-LABELS emits no untagged FETCH (unlike
		// STORE FLAGS) — only the tagged OK.
		c.C(`A003 STORE 1 +X-GM-LABELS ("Paperless" "Invoices")`)
		c.OK(`A003`)

		// FETCH returns them sorted.
		c.C(`A004 FETCH 1 (X-GM-LABELS)`)
		c.S(`* 1 FETCH (X-GM-LABELS ("Invoices" "Paperless"))`)
		c.OK(`A004`)

		// Remove one label.
		c.C(`A005 STORE 1 -X-GM-LABELS ("Invoices")`)
		c.OK(`A005`)
		c.C(`A006 FETCH 1 (X-GM-LABELS)`)
		c.S(`* 1 FETCH (X-GM-LABELS ("Paperless"))`)
		c.OK(`A006`)

		// Remove the last label -> back to empty.
		c.C(`A007 STORE 1 -X-GM-LABELS ("Paperless")`)
		c.OK(`A007`)
		c.C(`A008 FETCH 1 (X-GM-LABELS)`)
		c.S(`* 1 FETCH (X-GM-LABELS ())`)
		c.OK(`A008`)
	})
}

// Python's imaplib (used by Paperless-NGX) sends the label without parentheses.
// This is the exact wire form Paperless emits, so it must round-trip.
func TestGmailLabelsBareLabelForm(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("b001 CREATE saved-messages")
		c.S("b001 OK CREATE")

		c.doAppend(`saved-messages`, buildRFC5322TestLiteral(`To: 1@pm.me`)).expect("OK")

		c.C(`A001 SELECT saved-messages`)
		c.Se(`A001 OK [READ-WRITE] SELECT`)

		c.C(`A002 STORE 1 +X-GM-LABELS Paperless`)
		c.OK(`A002`)

		c.C(`A003 FETCH 1 (X-GM-LABELS)`)
		c.S(`* 1 FETCH (X-GM-LABELS ("Paperless"))`)
		c.OK(`A003`)
	})
}

func TestGmailLabelsWithSpaces(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("b001 CREATE saved-messages")
		c.S("b001 OK CREATE")

		c.doAppend(`saved-messages`, buildRFC5322TestLiteral(`To: 1@pm.me`)).expect("OK")

		c.C(`A001 SELECT saved-messages`)
		c.Se(`A001 OK [READ-WRITE] SELECT`)

		c.C(`A002 STORE 1 +X-GM-LABELS ("Label With Spaces")`)
		c.OK(`A002`)

		c.C(`A003 FETCH 1 (X-GM-LABELS)`)
		c.S(`* 1 FETCH (X-GM-LABELS ("Label With Spaces"))`)
		c.OK(`A003`)
	})
}

func TestGmailLabelsMultipleMessages(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.C("b001 CREATE saved-messages")
		c.S("b001 OK CREATE")

		c.doAppend(`saved-messages`, buildRFC5322TestLiteral(`To: 1@pm.me`)).expect("OK")
		c.doAppend(`saved-messages`, buildRFC5322TestLiteral(`To: 2@pm.me`)).expect("OK")
		c.doAppend(`saved-messages`, buildRFC5322TestLiteral(`To: 3@pm.me`)).expect("OK")

		c.C(`A001 SELECT saved-messages`)
		c.Se(`A001 OK [READ-WRITE] SELECT`)

		// Label a range of messages in one STORE.
		c.C(`A002 STORE 1:3 +X-GM-LABELS ("Paperless")`)
		c.OK(`A002`)

		c.C(`A003 FETCH 1:3 (X-GM-LABELS)`)
		c.S(
			`* 1 FETCH (X-GM-LABELS ("Paperless"))`,
			`* 2 FETCH (X-GM-LABELS ("Paperless"))`,
			`* 3 FETCH (X-GM-LABELS ("Paperless"))`,
		)
		c.OK(`A003`)
	})
}

// SEARCH X-GM-LABELS and its negation are how Paperless-NGX finds untagged mail
// (NOT X-GM-LABELS "<tag>") and verifies tagging. The connector pushes the label
// mailbox + membership asynchronously, so we flush before searching.
func TestGmailLabelsSearchRoundTrip(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		c.C("b001 CREATE saved-messages")
		c.S("b001 OK CREATE")

		c.doAppend(`saved-messages`, buildRFC5322TestLiteral(`To: 1@pm.me`)).expect("OK")
		c.doAppend(`saved-messages`, buildRFC5322TestLiteral(`To: 2@pm.me`)).expect("OK")

		c.C(`A001 SELECT saved-messages`)
		c.Se(`A001 OK [READ-WRITE] SELECT`)

		// Before tagging, nothing matches and everything is "not labelled".
		c.C(`A002 SEARCH X-GM-LABELS "Paperless"`)
		c.S(`* SEARCH`)
		c.OK(`A002`)

		// Tag only the first message.
		c.C(`A003 STORE 1 +X-GM-LABELS ("Paperless")`)
		c.OK(`A003`)

		// Make the connector-pushed label mailbox + membership visible.
		s.flush("user")

		c.C(`A004 SEARCH X-GM-LABELS "Paperless"`)
		c.S(`* SEARCH 1`)
		c.OK(`A004`)

		// The dedupe query Paperless-NGX issues: everything NOT yet tagged.
		c.C(`A005 SEARCH NOT X-GM-LABELS "Paperless"`)
		c.S(`* SEARCH 2`)
		c.OK(`A005`)
	})
}
