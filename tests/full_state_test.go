package tests

import (
	"context"
	"fmt"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/imap"
	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

/*
 * 1 user 2 mailboxes
 * ----------------
 * Login
 * list Mailboxes and get their status
 * select Archive
 * Receive a new message on Archive and read it
 * copy the message to INBOX and close Archive
 * check on the INBOX mailbox, that the mail exists
 * check back on Archive that it's still there.
 */
func TestSimpleMailCopy(t *testing.T) {
	const (
		mailboxName = "Archive"
		messagePath = "testdata/afternoon-meeting.eml"
	)

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
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
		uidFetchAndCheckMailHeader(t, client, 1, 1, true)
		uidFetchAndCheckMailContent(t, client, 1, 1, true)

		// copy it to INBOX
		require.NoError(t, client.Copy(createSeqSet("1"), "INBOX"))
		// select INBOX
		status, err = client.Select("INBOX", false)
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages, "Expected message count does not match")
		// read the same mail
		uidFetchAndCheckMailHeader(t, client, 1, 1, true)
		uidFetchAndCheckMailContent(t, client, 1, 1, true)
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
 * Noop + Fetch flags (as in thunderbird)).
 */
func TestReceptionOnIdle(t *testing.T) {
	const (
		mailboxName = "INBOX"
		messagePath = "testdata/afternoon-meeting.eml"
	)

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(c *client.Client, sess *testSession) {
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

		// prepare to stop idling.
		stopped := false
		stop := make(chan struct{})
		done := make(chan error, 1)
		// Create a channel to receive mailbox updates.
		updates := make(chan client.Update)
		c.Updates = updates

		// idling.
		go func() {
			labels := pprof.Labels("test", "client", "idling", "idle")
			pprof.Do(context.Background(), labels, func(_ context.Context) {
				defer func() {
					stopped = true
				}()
				done <- c.Idle(stop, nil)
			})
		}()

		// receiving messages from another client.
		go func() {
			labels := pprof.Labels("test", "client", "sending", "idle")
			pprof.Do(context.Background(), labels, func(_ context.Context) {
				cli := sess.newClient()
				defer func() {
					require.NoError(t, cli.Logout())
					close(stop)
				}()
				require.NoError(t, cli.Login("user", "pass"))
				for i := 0; i < 3; i++ {
					require.NoError(t, doAppendWithClientFromFile(t, cli, mailboxName, messagePath, time.Now()))
				}
			})
		}()

		// Listen for updates
		var existsUpdate uint32 = 0
		var recentUpdate uint32 = 0

		for !stopped {
			select {
			case update := <-updates:
				boxUpdate, ok := update.(*client.MailboxUpdate)
				if ok {
					recentUpdate = boxUpdate.Mailbox.Recent
					existsUpdate = boxUpdate.Mailbox.Messages
				}
			case err := <-done:
				require.NoError(t, err)
			}
		}

		c.Updates = nil
		close(updates)
		close(done)

		require.Equal(t, existsUpdate, uint32(3), "Not received the good amount of exists update")
		require.Equal(t, recentUpdate, uint32(3), "Not received the good amount of recent update")
		{
			expectedFlags := []string{
				goimap.RecentFlag,
			}
			uidFetchAndCheckFlags(t, c, 1, 3, expectedFlags)
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
 * Either delete it, Archive it or put it as unseen.
 */
func TestMorningFiltering(t *testing.T) {
	runOneToOneTestClientWithData(t, defaultServerOptions(t), func(client *client.Client, s *testSession, mbox string, mboxID imap.LabelID) {
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
			uidFetchAndCheckFlags(t, client, 1, 100, expectedFlags)
		}
		nbUnseen := 0
		nbArchived := 0
		for i := 1; i <= 100; i++ {
			strId := fmt.Sprint(i)
			// read the content
			uidFetchAndCheckMailHeader(t, client, i, i, false)
			uidFetchAndCheckMailContent(t, client, i, i, false)
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
	})
}

func uidFetchAndCheckFlags(t *testing.T, client *client.Client, first int, last int, flags []string) {
	const sectionStr = "FLAGS"

	nbRes := (last - first) + 1
	seqSet := fmt.Sprint(first)

	if first != last {
		seqSet += ":" + fmt.Sprint(last)
	}

	fetchResult := newFetchCommand(t, client).withItems(sectionStr).fetchUid(seqSet)

	for i := first; i <= last; i++ {
		fetchResult.forUid(uint32(i), func(builder *validatorBuilder) {
			for _, flag := range flags {
				builder.wantFlags(flag)
			}
		})
	}
	fetchResult.checkAndRequireMessageCount(nbRes)
}

func uidFetchAndCheckMailHeader(t *testing.T, client *client.Client, first int, last int, expectAfternoon bool) {
	const sectionStr = "BODY.PEEK[HEADER.FIELDS (Date From Subject)]"

	const sectionNotPeekStr = "BODY[HEADER.FIELDS (Date From Subject)]"

	nbRes := (last - first) + 1
	seqSet := fmt.Sprint(first)

	if first != last {
		seqSet += ":" + fmt.Sprint(last)
	}

	fetchResult := newFetchCommand(t, client).withItems(sectionStr).fetchUid(seqSet)
	for i := first; i <= last; i++ {
		fetchResult.forUid(uint32(i), func(builder *validatorBuilder) {
			builder.ignoreFlags()
			if expectAfternoon {
				builder.wantSection(sectionNotPeekStr,
					`Date: Mon, 7 Feb 1994 21:52:25 -0800 (PST)`,
					`From: Fred Foobar <foobar@Blurdybloop.COM>`,
					`Subject: afternoon meeting`,
					``,
					``,
				)
			} else {
				builder.wantSectionNotEmpty(sectionNotPeekStr)
			}
		})
	}
	fetchResult.checkAndRequireMessageCount(nbRes)
}

func uidFetchAndCheckMailContent(t *testing.T, client *client.Client, first int, last int, expectAfternoon bool) {
	const sectionStr = "BODY[TEXT]"

	nbRes := (last - first) + 1
	seqSet := fmt.Sprint(first)

	if first != last {
		seqSet += ":" + fmt.Sprint(last)
	}

	fetchResult := newFetchCommand(t, client).withItems(sectionStr).fetchUid(seqSet)

	for i := first; i <= last; i++ {
		fetchResult.forUid(uint32(i), func(builder *validatorBuilder) {
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
	fetchResult.checkAndRequireMessageCount(nbRes)
}
