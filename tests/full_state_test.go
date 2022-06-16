package tests

import (
	"fmt"
	"testing"
	"time"

	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func _uidFetchFlags(t *testing.T, client *client.Client, first int, last int, flags []string) {
	const sectionStr = "FLAGS"
	nbRes := (last - first) + 1
	seqSet := fmt.Sprint(first)
	if first != last {
		seqSet += ":" + fmt.Sprint(last)
	}
	fetchResult := newFetchCommand(t, client).withItems(sectionStr).fetchUid(seqSet)
	fetchResult.checkAndRequireMessageCount(nbRes)
	for i := 1; i <= nbRes; i++ {
		fetchResult.forSeqNum(uint32(i), func(builder *validatorBuilder) {
			for _, flag := range flags {
				builder.wantFlags(flag)
			}
		})
	}
}

func _uidFetchMailHeader(t *testing.T, client *client.Client, first int, last int, expectAfternoon bool) {
	const sectionStr = "BODY.PEEK[HEADER.FIELDS (DATE FROM Subject)]"
	nbRes := (last - first) + 1
	seqSet := fmt.Sprint(first)
	if first != last {
		seqSet += ":" + fmt.Sprint(last)
	}
	fetchResult := newFetchCommand(t, client).withItems(sectionStr).fetchUid(seqSet)
	fetchResult.checkAndRequireMessageCount(nbRes)
	for i := 1; i <= nbRes; i++ {
		fetchResult.forSeqNum(uint32(i), func(builder *validatorBuilder) {
			builder.ignoreFlags()
			if expectAfternoon {
				builder.wantSection(sectionStr,
					`Date: Mon, 7 Feb 1994 21:52:25 -0800 (PST)`,
					`From: Fred Foobar <foobar@Blurdybloop.COM>`,
					`Subject: afternoon meeting`,
					``,
					``,
				)
			} else {
				builder.wantSectionNotEmpty(sectionStr)
			}
		})
	}
}

func _uidFetchMailContent(t *testing.T, client *client.Client, first int, last int, expectAfternoon bool) {
	const sectionStr = "BODY[TEXT]"
	nbRes := (last - first) + 1
	seqSet := fmt.Sprint(first)
	if first != last {
		seqSet += ":" + fmt.Sprint(last)
	}
	fetchResult := newFetchCommand(t, client).withItems(sectionStr).fetchUid(seqSet)
	fetchResult.checkAndRequireMessageCount(nbRes)
	for i := 1; i <= nbRes; i++ {
		fetchResult.forSeqNum(uint32(i), func(builder *validatorBuilder) {
			builder.ignoreFlags()
			if expectAfternoon {
				builder.wantSection(sectionStr,
					`Hello Joe, do you think we can meet at 3:30 tomorrow?`,
					``,
					``,
				)
			} else {
				builder.wantSectionNotEmpty(sectionStr)
			}
		})
	}
}

/*
 * 1 user 2 mailboxes
 * ----------------
 * Login
 * list Mailboxes and get their status
 * select Archive
 * Receive a new message on Archive and read it
 * copy the message to INBOX and close Archive
 * check on the INBOX mailbox, that the mail exists
 * check back on Archive that it's still there
 */
func TestSimpleMailCopy(t *testing.T) {
	const (
		mailboxName = "Archive"
		messagePath = "testdata/afternoon-meeting.eml"
	)

	runOneToOneTestClientWithAuth(t, "user", "pass", "/", func(client *client.Client, _ *testSession) {

		require.NoError(t, client.Create("Archive"))
		// list mailbox
		{
			expectedMailboxNames := []string{
				"INBOX",
				mailboxName,
			}
			expectedAttributes := []string{goimap.UnmarkedAttr}
			checkMailboxesMatchNamesAndAttributes(t, client, "", "*", expectedMailboxNames, expectedAttributes)
		}
		// select Archive
		status, err := client.Select(mailboxName, false)
		require.NoError(t, err)
		require.Equal(t, uint32(0), status.Messages, "Expected message count does not match")

		// receive a new mail
		require.NoError(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now()))
		// status
		status, err = client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages, "Expected message count does not match")
		// read the mail
		_uidFetchMailHeader(t, client, 1, 1, true)
		_uidFetchMailContent(t, client, 1, 1, true)

		// copy it to INBOX
		require.NoError(t, client.Copy(createSeqSet("1"), "INBOX"))
		// select INBOX
		status, err = client.Select("INBOX", false)
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages, "Expected message count does not match")
		// read the same mail
		_uidFetchMailHeader(t, client, 1, 1, true)
		_uidFetchMailContent(t, client, 1, 1, true)
	})
}

/*
 * 1 user 1 mailbox
 * ----------------
 * Login
 * list Mailbox and get their status
 * select INBOX
 * IDLE
 * Receive 3 messages
 * Done IDLING (Being notified of the 3 new mails)
 * Noop + Fetch flags (as in thunderbird))
 */

func TestReceptionOnIdle(t *testing.T) {
	const (
		mailboxName = "INBOX"
		messagePath = "testdata/afternoon-meeting.eml"
	)

	runOneToOneTestClientWithAuth(t, "user", "pass", "/", func(c *client.Client, _ *testSession) {
		// list mailbox
		{
			expectedMailboxNames := []string{
				mailboxName,
			}
			expectedAttributes := []string{goimap.UnmarkedAttr}
			checkMailboxesMatchNamesAndAttributes(t, c, "", "*", expectedMailboxNames, expectedAttributes)
		}
		status, err := c.Select(mailboxName, false)
		require.NoError(t, err)
		require.Equal(t, uint32(0), status.Messages, "Expected message count does not match")

		// receive 3 new mail while IDLE
		stop := make(chan struct{})
		go func(t *testing.T, c *client.Client, stop chan struct{}, mailboxName string, messagePath string, nb int) {
			for i := 0; i < nb; i++ {
				time.Sleep(100)
				require.NoError(t, doAppendWithClientFromFile(t, c, mailboxName, messagePath, time.Now()))
			}
			close(stop)
		}(t, c, stop, mailboxName, messagePath, 3)

		require.NoError(t, c.Idle(stop, nil))
		require.NoError(t, c.Noop())
		{
			expectedFlags := []string{
				goimap.RecentFlag,
			}
			_uidFetchFlags(t, c, 1, 3, expectedFlags)
		}

		// status
		status, err = c.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(3), status.Messages, "Expected message count does not match")
	})
}

/*
 * 1 User with a daily routine to filter mails
 * ----------------
 * Login
 * Read Mails
 * Either delete it, Archive it or put it as unseen
 */
func TestMorningFiltering(t *testing.T) {
	runOneToOneTestClientWithData(t, "user", "pass", "/", func(client *client.Client, s *testSession, mbox, mboxID string) {
		require.NoError(t, client.Create("ReadLater"))
		require.NoError(t, client.Create("Archive"))

		// list mailbox
		{
			expectedMailboxNames := []string{
				"ReadLater",
				"Archive",
				"INBOX",
				mbox,
			}
			expectedAttributes := []string{goimap.UnmarkedAttr}
			checkMailboxesMatchNamesAndAttributes(t, client, "", "*", expectedMailboxNames, expectedAttributes)
		}
		{
			// There are 100 messages in the origin and no messages in the destination.
			mailboxStatus, err := client.Status(mbox, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(100), mailboxStatus.Messages)
		}
		{
			expectedFlags := []string{
				goimap.RecentFlag,
			}
			_uidFetchFlags(t, client, 1, 100, expectedFlags)
		}
		nbUnseen := 0
		nbArchived := 0
		for i := 1; i <= 100; i++ {
			strId := fmt.Sprint(i)
			// read the content
			_uidFetchMailHeader(t, client, i, i, false)
			_uidFetchMailContent(t, client, i, i, false)
			switch i % 3 {
			case 0:
				// either Delete
				uidStoreWithRetrievalClient(t, client, createSeqSet(strId), goimap.AddFlags, []interface{}{goimap.DeletedFlag})
				expungedIds := expungeClient(t, client)
				require.Equal(t, 1, len(expungedIds))
				break
			case 1:
				// or unseen
				uidStoreWithRetrievalClient(t, client, createSeqSet(strId), goimap.RemoveFlags, []interface{}{goimap.SeenFlag})
				require.NoError(t, client.UidMove(createSeqSet(strId), "ReadLater"))
				nbUnseen++
				break
			case 2:
				// or Archive
				require.NoError(t, client.UidMove(createSeqSet(strId), "Archive"))
				nbArchived++
				break
			}
		}
		{
			mailboxStatus, err := client.Status(mbox, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(0), mailboxStatus.Messages)
		}
		{
			mailboxStatus, err := client.Status("Archive", []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(nbArchived), mailboxStatus.Messages)
		}
		{
			mailboxStatus, err := client.Status("ReadLater", []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(nbUnseen), mailboxStatus.Messages)

		}
		{
			mailboxStatus, err := client.Select("ReadLater", false)
			require.NoError(t, err)
			require.Equal(t, uint32(nbUnseen), mailboxStatus.Unseen)
		}
	})
}
