package tests

import (
	"fmt"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/ProtonMail/gluon/imap"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/htmlindex"
)

func TestSearchCharSetUTF8(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		c.C(`tag select inbox`).OK("tag")

		// Encode "ééé" as UTF-8.
		b := enc("ééé", "UTF-8")

		// Append a message with that as the body.
		c.doAppend("inbox", buildRFC5322TestLiteral("To: 1@pm.me\r\n\r\nééé")).expect("OK")

		// Search for it with UTF-8 encoding.
		c.Cf(`TAG SEARCH CHARSET UTF-8 BODY {%v}`, len(b)).Continue().Cb(b).S("* SEARCH 1").OK("TAG")
	})
}

func TestSearchCharSetISO88591(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		c.C(`tag select inbox`).OK("tag")

		// Encode "ééé" as ISO-8859-1.
		b := enc("ééé", "ISO-8859-1")

		// Assert that b is no longer valid UTF-8.
		require.False(t, utf8.Valid(b))

		// Append a message with that as the body.
		c.doAppend("inbox", buildRFC5322TestLiteral("To: 1@pm.me\r\n\r\nééé")).expect("OK")

		// Search for it with ISO-8859-1 encoding (literal).
		c.Cf(`TAG SEARCH CHARSET ISO-8859-1 BODY {%v}`, len(b)).Continue().Cb(b).S("* SEARCH 1").OK("TAG")

		// Search for it with ISO-8859-1 encoding (direct).
		c.Cf(`TAG SEARCH CHARSET ISO-8859-1 BODY ` + string(b)).S("* SEARCH 1").OK("TAG")
	})
}

func TestSearchCharSetASCII(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C("A001 search CHARSET US-ASCII TEXT foo")
		c.S("* SEARCH 75")
		c.OK("A001")
	})
}

func TestSearchCharSetInvalid(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C("A001 search CHARSET invalid-charset TEXT foo").NO("A001", "BADCHARSET")
	})
}

func TestSearchAll(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C("A001 search all")
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A001")
	})
}

func TestSearchAnswered(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Set some messages as answered.
		c.C(`A001 STORE 10,20,30,40,50 +FLAGS (\Answered)`).OK("A001")

		// They should show up in search.
		c.C("A002 search answered")
		c.S("* SEARCH 10 20 30 40 50")
		c.OK("A002")
	})
}

func TestSearchBcc(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C(`A001 search bcc "dovecot@procontrol.fi"`)
		c.S("* SEARCH 49 50")
		c.OK("A001")

		// Search is also case-insensitive.
		c.C(`A001 search bcc "dovecot@PROcontrol.FI"`)
		c.S("* SEARCH 49 50")
		c.OK("A001")
	})
}

func TestSearchBefore(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// None
		c.C(`A001 search before 13-Aug-1982`)
		c.S("* SEARCH")
		c.OK("A001")

		// All
		c.C(`A001 search before 13-Aug-2200`)
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A001")

		// Earliest messages
		c.C(`A001 search before 30-Jul-2002`)
		c.S("* SEARCH 1 2")
		c.OK("A001")
	})
}

func TestSearchBody(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C(`A001 search body "Content-Length saves just the size of mail body"`)
		c.S("* SEARCH 50")
		c.OK("A001")

		// Search is also case-insensitive.
		c.C(`A001 search body "Content-LenGTH sAvEs just the size of MaiL body"`)
		c.S("* SEARCH 50")
		c.OK("A001")
	})
}

func TestSearchCc(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C(`A001 search cc "Dovecot Mailinglist <dovecot@procontrol.fi>"`)
		c.S("* SEARCH 53 55 60")
		c.OK("A001")

		// Search is also case-insensitive.
		c.C(`A001 search cc "DoVeCot Mailinglist <doveCOT@proconTROl.fi>"`)
		c.S("* SEARCH 53 55 60")
		c.OK("A001")
	})
}

func TestSearchDeleted(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Set some messages as deleted.
		c.C(`A001 STORE 10,20,30,40,50 +FLAGS (\Deleted)`).OK("A001")

		// They should show up in search.
		c.C("A002 search deleted")
		c.S("* SEARCH 10 20 30 40 50")
		c.OK("A002")
	})
}

func TestSearchDraft(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Set some messages as draft.
		c.C(`A001 STORE 10,20,30,40,50 +FLAGS (\Draft)`).OK("A001")

		// They should show up in search.
		c.C("A002 search draft")
		c.S("* SEARCH 10 20 30 40 50")
		c.OK("A002")
	})
}

func TestSearchFlagged(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Set some messages as flagged.
		c.C(`A001 STORE 10,20,30,40,50 +FLAGS (\Flagged)`).OK("A001")

		// They should show up in search.
		c.C("A002 search flagged")
		c.S("* SEARCH 10 20 30 40 50")

		c.OK("A002")
	})
}

func TestSearchFrom(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C(`A001 search from "\"Vanessa Lintner\" <reply@seekercenter.net>"`)
		c.S("* SEARCH 5")
		c.OK("A001")

		c.C(`A001 search from "reply@seekercenter.net"`)
		c.S("* SEARCH 5")
		c.OK("A001")

		c.C(`A001 search from "reply@seeKERcenTER.net"`)
		c.S("* SEARCH 5")
		c.OK("A001")
	})
}

func TestSearchHeader(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C(`A001 search header Message-ID "<20020723193923.J22431@irccrew.org>"`)
		c.S("* SEARCH 1")
		c.OK("A001")

		c.C(`A002 search header Message-ID "<20020807065824.C16470@irccrew.org>"`)
		c.S("* SEARCH 10")
		c.OK("A002")

		c.C(`A003 search header Message-ID "<006701c24183$c7f10d60$0200a8c0@eero>"`)
		c.S("* SEARCH 20")
		c.OK("A003")
	})
}

func TestSearchKeyword(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Set some messages as with a custom flag.
		c.C(`A001 STORE 10,20,30,40,50 +FLAGS (my-special-flag)`).OK("A001")

		// They should show up in search.
		c.C("A002 search keyword my-special-flag")
		c.S("* SEARCH 10 20 30 40 50")
		c.OK("A002")
	})
}

func TestSearchLarger(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C("A001 search larger 9500")
		c.S("* SEARCH 47 48")
		c.OK("A001")
	})
}

func TestSearchNew(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Initially all messages are recent and none are seen. Thus they are all new.
		c.C("A001 search new")
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A001")

		// Mark most as seen.
		c.C(`A002 STORE 11:* +FLAGS (\Seen)`).OK("A002")

		// Only the unseen ones are still new.
		c.C("A003 search new")
		c.S("* SEARCH " + seq(1, 10))
		c.OK("A003")
	})
}

func TestSearchNot(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Set some messages as deleted.
		c.C(`A001 STORE 50:* +FLAGS (\Deleted)`).OK("A001")

		// The first half are deleted.
		c.C("A002 search deleted")
		c.S("* SEARCH " + seq(50, 100))
		c.OK("A002")

		// The second half are not deleted.
		c.C("A003 search not deleted")
		c.S("* SEARCH " + seq(1, 49))
		c.OK("A003")
	})
}

func TestSearchOld(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Initially all messages are recent so none are old.
		c.C("A001 search old")
		c.S("* SEARCH")
		c.OK("A001")

		// Re-select the mailbox.
		c.Cf("A002 select %v", mbox).OK("A002")

		// Now no messages are recent; all of them are old.
		c.C("A003 search old")
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A003")

		// Create a new message; it will be recent and thus not old.
		c.doAppend(mbox, buildRFC5322TestLiteral(`To: 1@pm.me`)).expect("OK")

		// It will be returned in the search result.
		c.C("A004 search old")
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A004")

		c.C("A005 search not recent")
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A005")
	})
}

func TestSearchOn(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// None
		c.C(`A001 search on 13-Aug-1982`)
		c.S("* SEARCH")
		c.OK("A001")

		// One
		c.C(`A001 search on 23-Jul-2002`)
		c.S("* SEARCH 1")
		c.OK("A001")

		// More
		c.C(`A001 search on 26-Nov-2002`)
		c.S("* SEARCH 99 100")
		c.OK("A001")
	})
}

func TestSearchSentOnAndOn(t *testing.T) {
	// Test search senton/on when internal date (e.g. 11-Aug-2002) and
	// header date (10-Aug-2002) are different.
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C(`A001 search on 10-Aug-2002`)
		c.S("* SEARCH")
		c.OK("A001")

		c.C(`A001 search senton 10-Aug-2002`)
		c.S("* SEARCH 19")
		c.OK("A001")

		c.C(`A001 search on 11-Aug-2002`)
		c.S("* SEARCH 19")
		c.OK("A001")

		c.C(`A001 search senton 11-Aug-2002`)
		c.S("* SEARCH")
		c.OK("A001")
	})
}

func TestSearchOr(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Set some messages as deleted.
		c.C(`A001 STORE 1:30 +FLAGS (\Deleted)`).OK("A001")

		// Set some messages as seen.
		c.C(`A002 STORE 10:40 +FLAGS (\Seen)`).OK("A002")

		// Set some messages as draft.
		c.C(`A003 STORE 20:50 +FLAGS (\Draft)`).OK("A003")

		// 1:40 are deleted or seen.
		c.C("A004 search or deleted seen")
		c.S("* SEARCH " + seq(1, 40))
		c.OK("A004")

		// 10:50 are seen or draft.
		c.C("A004 search or seen draft")
		c.S("* SEARCH " + seq(10, 50))
		c.OK("A004")

		// 1:50 are deleted, seen or draft.
		c.C("A004 search or deleted seen draft")
		c.S("* SEARCH " + seq(1, 50))
		c.OK("A004")

		// Alternative way of writing deleted, seen or draft.
		c.C("A004 search or deleted or seen draft")
		c.S("* SEARCH " + seq(1, 50))
		c.OK("A004")

		// 1:9,41:50 are not seen but either deleted or draft.
		c.C("A004 search not seen or deleted draft")
		c.S("* SEARCH " + seq(1, 9) + " " + seq(41, 50))
		c.OK("A004")
	})
}

func TestSearchRecent(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Initially all messages are recent.
		c.C("A001 search recent")
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A001")

		// Re-select the mailbox.
		c.Cf("A002 select %v", mbox).OK("A002")

		// Now no messages are recent.
		c.C("A003 search recent")
		c.S("* SEARCH")
		c.OK("A003")

		// Create a new message; it will be recent.
		c.doAppend(mbox, buildRFC5322TestLiteral(`To: 1@pm.me`)).expect("OK")

		// It will be returned in the search result.
		c.C("A004 search recent")
		c.S("* SEARCH 101")
		c.OK("A004")
	})
}

func TestSearchSeen(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Initially no messages are seen.
		c.C("A001 search seen")
		c.S("* SEARCH")
		c.OK("A001")

		// Set some messages as seen.
		c.C(`A002 STORE 10,20,30,40,50 +FLAGS (\Seen)`).OK("A002")

		// They should show up in search.
		c.C("A003 search seen")
		c.S("* SEARCH 10 20 30 40 50")
		c.OK("A003")
	})
}

func TestSearchSentBefore(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C(`A001 search sentbefore 10-Aug-2002`)
		c.S("* SEARCH " + seq(1, 18))
		c.OK("A001")
	})
}

func TestSearchSentSinceAndSentBefore(t *testing.T) {
	// The result of this test should be no messages in the search result. Due to timezone adjustments, by
	// mail.ParseDate the date was being converted to 17 Feb 2003 22:29:37 +000, causing the search to pass
	// rather than fail.
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.doAppend("INBOX", buildRFC5322TestLiteral("Date: 18 Feb 2003 00:29:37 +0200\n\nTo: foo@foo.com\r\n"))
		c.C(`A002 SELECT INBOX`)
		c.Se(`A002 OK [READ-WRITE] SELECT`)

		t := time.Now()

		c.C(fmt.Sprintf(`A001 search since %s sentbefore 18-Feb-2003`, t.Format("02-Jan-2006")))
		c.S("* SEARCH")
		c.OK("A001")
	})
}

func TestSearchSentOn(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C(`A001 search senton 22-Aug-2002`)
		c.S("* SEARCH 21")
		c.OK("A001")
	})
}

func TestSearchSentSince(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C(`A001 search sentsince 13-Nov-2002`)
		c.S("* SEARCH " + seq(93, 100))
		c.OK("A001")
	})
}

func TestSearchSince(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// None
		c.C(`A001 search since 13-Aug-2200`)
		c.S("* SEARCH")
		c.OK("A001")

		// All
		c.C(`A001 search since 13-Aug-1982`)
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A001")

		// Latest messages
		c.C(`A001 search since 26-Nov-2002`)
		c.S("* SEARCH 99 100")
		c.OK("A001")
	})
}

func TestSearchSmaller(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C("A001 search smaller 1250")
		c.S("* SEARCH 1 8 83")
		c.OK("A001")
	})
}

func TestSearchSubject(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C(`A001 search subject "first test mail"`)
		c.S("* SEARCH 1")
		c.OK("A001")

		c.C(`A002 search subject "v0.96 released"`)
		c.S("* SEARCH 16")
		c.OK("A002")

		c.C(`A003 search subject "mbox problems"`)
		c.S("* SEARCH 100")
		c.OK("A003")

		// Subject search is case-insensitive.
		c.C(`A003 search subject "MBOX PROBLEMS"`)
		c.S("* SEARCH 100")
		c.OK("A003")
	})
}

func TestSearchText(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Search something from the header.
		c.C(`A001 search text "Message-ID: <006701c24183$c7f10d60$0200a8c0@eero>"`)
		c.S("* SEARCH 20")
		c.OK("A001")

		// Search something from the body.
		c.C(`A002 search text "Content-Length saves just the size of mail body"`)
		c.S("* SEARCH 50")
		c.OK("A002")

		// Text search is case-insensitive.
		c.C(`A002 search text "ContenT-LeNgTh saveS jUst the Size of mail body"`)
		c.S("* SEARCH 50")
		c.OK("A002")
	})
}

func TestSearchTo(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C(`A001 search to "Timo Sirainen <tss@iki.fi>"`)
		c.S("* SEARCH 49")
		c.OK("A001")

		c.C(`A001 search to "Timo SirAINEN <tSS@ikI.FI>"`)
		c.S("* SEARCH 49")
		c.OK("A001")
	})
}

func TestUIDSearchTo(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Remove some messages to ensure sequence doesn't match uid
		c.C(`A002 STORE 10,20,30,40,50 +FLAGS (\DELETED)`).OK("A002")
		c.C(`A003 EXPUNGE`).OK("A003")

		c.C(`A001 UID search to "Timo Sirainen <tss@iki.fi>"`)
		c.S("* SEARCH 49")
		c.OK("A001")
	})
}

func TestSearchUID(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C(`A001 search uid 1`)
		c.S("* SEARCH 1")
		c.OK("A001")

		c.C(`A002 search uid 1:10`)
		c.S("* SEARCH " + seq(1, 10))
		c.OK("A002")

		c.C(`A003 search uid 1:10,2:20,3:30,4:40,5:50`)
		c.S("* SEARCH " + seq(1, 50))
		c.OK("A003")

		c.C(`A003 search uid *:1`)
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A003")
	})
}

func TestSearchUIDAfterDelete(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Remove some messages to ensure sequence doesn't match uid
		c.C(`A002 STORE 5,6,7 +FLAGS (\DELETED)`).OK("A002")
		c.C(`A003 EXPUNGE`).OK("A003")

		c.C(`A002 search uid 1:10`)
		c.S("* SEARCH " + seq(1, 7))
		c.OK("A002")
	})
}

func TestSearchUnanswered(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Initially, nothing is answered; everything is unanswered.
		c.C("A001 search unanswered")
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A001")

		// Set some messages as answered.
		c.C(`A002 STORE 50:* +FLAGS (\Answered)`).OK("A002")

		// The rest should now show up in the search.
		c.C("A003 search unanswered")
		c.S("* SEARCH " + seq(1, 49))
		c.OK("A003")

		// Negated query should return the same results.
		c.C("A004 search not answered")
		c.S("* SEARCH " + seq(1, 49))
		c.OK("A004")
	})
}

func TestSearchUndeleted(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Initially, nothing is deleted; everything is undeleted.
		c.C("A001 search undeleted")
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A001")

		// Set some messages as deleted.
		c.C(`A002 STORE 50:* +FLAGS (\Deleted)`).OK("A002")

		// The rest should now show up in the search.
		c.C("A003 search undeleted")
		c.S("* SEARCH " + seq(1, 49))
		c.OK("A003")

		// Negated query should return the same results.
		c.C("A004 search not deleted")
		c.S("* SEARCH " + seq(1, 49))
		c.OK("A004")
	})
}

func TestSearchUndraft(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Initially, nothing is draft; everything is undraft.
		c.C("A001 search undraft")
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A001")

		// Set some messages as draft.
		c.C(`A002 STORE 50:* +FLAGS (\Draft)`).OK("A002")

		// The rest should now show up in the search.
		c.C("A003 search undraft")
		c.S("* SEARCH " + seq(1, 49))
		c.OK("A003")

		// Negated query should return the same results.
		c.C("A004 search not draft")
		c.S("* SEARCH " + seq(1, 49))
		c.OK("A004")
	})
}

func TestSearchUnflagged(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Initially, nothing is flagged; everything is unflagged.
		c.C("A001 search unflagged")
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A001")

		// Set some messages as flagged.
		c.C(`A002 STORE 50:* +FLAGS (\Flagged)`).OK("A002")

		// The rest should now show up in the search.
		c.C("A003 search unflagged")
		c.S("* SEARCH " + seq(1, 49))
		c.OK("A003")

		// Negated query should return the same results.
		c.C("A004 search not flagged")
		c.S("* SEARCH " + seq(1, 49))
		c.OK("A004")
	})
}

func TestSearchUnkeyword(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Initially, nothing has a custom flag; everything is "unkeyword".
		c.C("A001 search unkeyword my-special-flag")
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A001")

		// Set some messages with the custom flag.
		c.C(`A002 STORE 50:* +FLAGS (my-special-flag)`).OK("A002")

		// The rest should now show up in the search.
		c.C("A003 search unkeyword my-special-flag")
		c.S("* SEARCH " + seq(1, 49))
		c.OK("A003")

		// Negated query should return the same results.
		c.C("A004 search not keyword my-special-flag")
		c.S("* SEARCH " + seq(1, 49))
		c.OK("A004")
	})
}

func TestSearchUnseen(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Initially, nothing is seen; everything is unseen.
		c.C("A001 search unseen")
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A001")

		// Set some messages as seen.
		c.C(`A002 STORE 50:* +FLAGS (\Seen)`).OK("A002")

		// The rest should now show up in the search.
		c.C("A003 search unseen")
		c.S("* SEARCH " + seq(1, 49))
		c.OK("A003")

		// Negated query should return the same results.
		c.C("A004 search not seen")
		c.S("* SEARCH " + seq(1, 49))
		c.OK("A004")
	})
}

func TestSearchSeqSet(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C(`A001 search 1`)
		c.S("* SEARCH 1")
		c.OK("A001")

		c.C(`A002 search 1:10`)
		c.S("* SEARCH " + seq(1, 10))
		c.OK("A002")

		c.C(`A003 search 1:10,2:20,3:30,4:40,5:50`)
		c.S("* SEARCH " + seq(1, 50))
		c.OK("A003")

		c.C(`A003 search *:1`)
		c.S("* SEARCH " + seq(1, 100))
		c.OK("A003")
	})
}

func TestSearchSeqSetFrom(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C(`A001 search 1:2,4:6,30:31 from "\"Vanessa Lintner\" <reply@seekercenter.net>"`)
		c.S("* SEARCH 5")
		c.OK("A001")
	})
}

func TestUIDSearchSeqSetFrom(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		c.C(`A002 STORE 1:4,10:80 +FLAGS (\DELETED)`).OK("A002")
		c.C(`A003 EXPUNGE`).OK("A003")

		c.C(`A001 UID search 1:* UID 1:10 from "\"Vanessa Lintner\" <reply@seekercenter.net>"`)
		c.S("* SEARCH 5")
		c.OK("A001")
	})
}

func TestSearchList(t *testing.T) {
	runOneToOneTestWithData(t, defaultServerOptions(t), func(c *testConnection, s *testSession, mbox string, mboxID imap.MailboxID) {
		// Set some messages as deleted.
		c.C(`A001 STORE 1:30 +FLAGS (\Deleted)`).OK("A001")

		// Set some messages as seen.
		c.C(`A002 STORE 10:40 +FLAGS (\Seen)`).OK("A002")

		// Set some messages as seen.
		c.C(`A003 STORE 20:50 +FLAGS (\Draft)`).OK("A003")

		// 10:30 are deleted and seen.
		c.C("A004 search (deleted seen)")
		c.S("* SEARCH " + seq(10, 30))
		c.OK("A004")

		// 20:40 are seen and draft.
		c.C("A004 search (seen draft)")
		c.S("* SEARCH " + seq(20, 40))
		c.OK("A004")

		// 20:30 are deleted, seen and draft.
		c.C("A004 search (deleted seen draft)")
		c.S("* SEARCH " + seq(20, 30))
		c.OK("A004")
	})
}

func enc(text, encoding string) []byte {
	enc, err := htmlindex.Get(encoding)
	if err != nil {
		panic(err)
	}

	b, err := enc.NewEncoder().Bytes([]byte(text))
	if err != nil {
		panic(err)
	}

	return b
}
