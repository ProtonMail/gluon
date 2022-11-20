package tests

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/ids"
	goimap "github.com/emersion/go-imap"
	uidplus "github.com/emersion/go-imap-uidplus"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestAppend(t *testing.T) {
	const (
		mailboxName = "saved-messages"
		messagePath = "testdata/afternoon-meeting.eml"
	)

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create(mailboxName))

		{
			// first check, empty mailbox
			status, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(0), status.Messages, "Expected message count does not match")
		}

		{
			require.NoError(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.SeenFlag))
			require.NoError(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.SeenFlag))
			require.NoError(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.SeenFlag))
			require.Error(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.RecentFlag))
		}

		{
			// second check, there should be 3 messages
			status, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(3), status.Messages, "Expected message count does not match")
		}
	})
}

func TestAppendWithUidPlus(t *testing.T) {
	const (
		mailboxName = "saved-messages"
		messagePath = "testdata/afternoon-meeting.eml"
	)

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		const validityUid = uint32(1)

		require.NoError(t, client.Create(mailboxName))

		{
			// first check, empty mailbox
			status, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(0), status.Messages, "Expected message count does not match")
		}

		{
			// Insert some messages
			clientUidPlus := uidplus.NewClient(client)
			{
				validity, uid, err := doAppendWithClientPlusFromFile(t, clientUidPlus, mailboxName, messagePath, goimap.SeenFlag)
				require.NoError(t, err)
				require.Equal(t, validityUid, validity)
				require.Equal(t, uint32(1), uid)
			}
			{
				validity, uid, err := doAppendWithClientPlusFromFile(t, clientUidPlus, mailboxName, messagePath, goimap.SeenFlag)
				require.NoError(t, err)
				require.Equal(t, validityUid, validity)
				require.Equal(t, uint32(2), uid)
			}
			{
				validity, uid, err := doAppendWithClientPlusFromFile(t, clientUidPlus, mailboxName, messagePath, goimap.SeenFlag)
				require.NoError(t, err)
				require.Equal(t, validityUid, validity)
				require.Equal(t, uint32(3), uid)
			}
			require.Error(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.RecentFlag))
		}

		{
			// second check, there should be 3 messages
			status, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(3), status.Messages, "Expected message count does not match")
		}
	})
}

func TestAppendNoSuchMailbox(t *testing.T) {
	const (
		mailboxName = "saved-messages"
		messagePath = "testdata/afternoon-meeting.eml"
	)

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		err := doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.SeenFlag)
		require.Error(t, err)
	})

	// Run the old test as well as there is no way to check for the `[TRYCREATE]` flag of the response
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.doAppendFromFile(`saved-messages`, `testdata/afternoon-meeting.eml`, `\Seen`).expect(`NO \[TRYCREATE\]`)
	})
}

func TestAppendWhileSelected(t *testing.T) {
	const (
		mailboxName = "saved-messages"
		messagePath = "testdata/afternoon-meeting.eml"
	)

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create(mailboxName))
		// Mailbox should have read-write modifier
		mailboxStatus, err := client.Select(mailboxName, false)
		require.NoError(t, err)
		require.Equal(t, false, mailboxStatus.ReadOnly)
		// Add new message
		require.NoError(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.SeenFlag))
		// EXISTS response is assigned to Messages member
		require.Equal(t, uint32(1), client.Mailbox().Messages)
		// Check if added
		status, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages, "Expected message count does not match")
	})
}

func TestAppendHeaderWithSpaceLine(t *testing.T) {
	r := require.New(t)

	const (
		mailboxName = "saved-messages"
		messagePath = "testdata/space_line_header.eml"
	)

	// Get full header from file
	fullMessage, err := os.ReadFile(messagePath)
	r.NoError(err)

	endOfHeader := []byte{13, 10, 13, 10}
	endOfHeaderKey := bytes.Index(fullMessage, endOfHeader)

	r.NotEqual(-1, endOfHeaderKey)
	wantHeader := string(fullMessage[0 : endOfHeaderKey+len(endOfHeader)])

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		r.NoError(client.Create(mailboxName))
		// Mailbox should have read-write modifier
		mailboxStatus, err := client.Select(mailboxName, false)
		r.NoError(err)
		r.False(mailboxStatus.ReadOnly)

		// Add new message
		require.NoError(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.SeenFlag))
		// EXISTS response is assigned to Messages member
		require.Equal(t, uint32(1), client.Mailbox().Messages)
		// Check if added
		status, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages, "Expected message count does not match")
		// Check has full header
		newFetchCommand(t, client).
			withItems(goimap.FetchRFC822Header).fetch("1").
			forSeqNum(1, func(v *validatorBuilder) {
				v.ignoreFlags()
				v.wantSectionString(goimap.FetchRFC822Header, func(t testing.TB, literal string) {
					haveHeader := skipGLUONHeader(literal)
					r.Equal(wantHeader, haveHeader)
				})
			}).check()
	})
}

type returnSameRemoteIDConnector struct {
	*connector.Dummy
	lock           sync.Mutex
	messageCreated bool
	createdMessage imap.Message
	messageLiteral []byte
}

func (r *returnSameRemoteIDConnector) CreateMessage(ctx context.Context, mboxID imap.MailboxID, literal []byte, flags imap.FlagSet, date time.Time) (imap.Message, []byte, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if !r.messageCreated {
		msg, l, err := r.Dummy.CreateMessage(ctx, mboxID, literal, flags, date)
		if err != nil {
			return imap.Message{}, nil, err
		}

		r.createdMessage = msg
		r.messageLiteral = l
		r.messageCreated = true
	}

	return r.createdMessage, r.messageLiteral, nil
}

type returnSameRemoteIDConnectorBuilder struct{}

func (returnSameRemoteIDConnectorBuilder) New(usernames []string, password []byte, period time.Duration, flags, permFlags, attrs imap.FlagSet) Connector {
	return &returnSameRemoteIDConnector{
		Dummy: connector.NewDummy(usernames, password, period, flags, permFlags, attrs),
	}
}

func TestAppendConnectorReturnsSameRemoteIDSameMBox(t *testing.T) {
	const (
		mailboxName = "INBOX"
		messagePath = "testdata/afternoon-meeting.eml"
	)

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withConnectorBuilder(&returnSameRemoteIDConnectorBuilder{})), func(client *client.Client, _ *testSession) {
		{
			require.NoError(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.SeenFlag))
			require.NoError(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.SeenFlag))
			require.NoError(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.SeenFlag))
		}
		{
			// second check, there should be  1 message
			status, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(1), status.Messages, "Expected message count does not match")
		}
	})
}

func TestAppendConnectorReturnsSameRemoteIDDifferentMBox(t *testing.T) {
	const (
		mailboxName      = "INBOX"
		mailboxNameOther = "saved-messages"
		messagePath      = "testdata/afternoon-meeting.eml"
	)

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withConnectorBuilder(&returnSameRemoteIDConnectorBuilder{})), func(client *client.Client, _ *testSession) {
		require.NoError(t, client.Create(mailboxNameOther))
		{
			require.NoError(t, doAppendWithClientFromFile(t, client, mailboxName, messagePath, time.Now(), goimap.SeenFlag))
			require.NoError(t, doAppendWithClientFromFile(t, client, mailboxNameOther, messagePath, time.Now(), goimap.SeenFlag))
		}
		{
			// there should be  1 message in mailboxName
			status, err := client.Status(mailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(1), status.Messages, "Expected message count does not match")
		}
		{
			// there should be 1 message in mailboxNameOther
			status, err := client.Status(mailboxNameOther, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(1), status.Messages, "Expected message count does not match")
		}
	})
}

func TestAppendCanHandleOutOfOrderUIDUpdates(t *testing.T) {
	// Make sure we are correctly handling the case where we have to clients doing append at the same time.
	// Both clients append a message and get assigned UID according to whichever got there first:
	//
	// * Client A -> Append -> UID 1
	// * Client B -> Append -> UID 2
	//
	// All Clients apply their changes to their local state immediately and will receive a deferred updates for the
	// same mailbox if other clients make updates.
	// In the case of client B, it appends UID2 as the first message and then later receives an update from A with
	// an UID lower than the last UID which caused unnecessary panics in the past.
	runManyToOneTestWithAuth(t, defaultServerOptions(t, withDisableParallelism()), []int{1, 2}, func(c map[int]*testConnection, session *testSession) {
		const MessageCount = 20

		// Select mailbox so that both clients get updates.
		c[1].C("A001 SELECT INBOX").OK("A001")
		c[2].C("A002 SELECT INBOX").OK("A002")

		appendFN := func(clientIndex int) {
			for i := 0; i < MessageCount; i++ {
				c[clientIndex+1].doAppend("INBOX", "To: f3@pm.me\r\n", "\\Seen").expect("OK")
			}
		}

		wg := sync.WaitGroup{}
		wg.Add(2)

		for i := 0; i < 2; i++ {
			go func(index int) {
				defer wg.Done()
				appendFN(index)
			}(i)
		}

		wg.Wait()

		validateUIDListFn := func(index int) {
			c[index].C("F001 FETCH 1:* (UID)")
			for i := 1; i <= MessageCount; i++ {
				c[index].S(fmt.Sprintf("* %v FETCH (UID %v)", i, i))
			}
			c[index].OK("F001")
		}

		validateUIDListFn(1)
		validateUIDListFn(2)
	})
}

func TestGODT2007AppendInternalIDPresentOnDeletedMessage(t *testing.T) {
	const (
		mailboxName = "saved-messages"
	)

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, s *testSession) {
		// Create message and mark deleted.
		mboxID := s.mailboxCreated("user", []string{mailboxName})
		messageID := s.messageCreated("user", mboxID, []byte("To: foo@bar.com\r\n"), time.Now())
		s.flush("user")

		_, err := client.Select(mailboxName, false)
		require.NoError(t, err)

		{
			// Check if the header is correctly set.
			result := newFetchCommand(t, client).withItems("UID", "BODY[HEADER]").fetch("1")
			result.forSeqNum(1, func(builder *validatorBuilder) {
				builder.ignoreFlags().wantSection("BODY[HEADER]", fmt.Sprintf("%v: 1", ids.InternalIDKey), "To: foo@bar.com\r\n")
				builder.wantUID(1)
			})
			result.checkAndRequireMessageCount(1)
		}

		s.messageDeleted("user", messageID)
		s.flush("user")

		// Add the same message back with the same id
		require.NoError(t, doAppendWithClient(client, mailboxName, fmt.Sprintf("%v: 1\r\nTo: foo@bar.com\r\n", ids.InternalIDKey), time.Now()))

		{
			// The message should have been created with a new internal id
			result := newFetchCommand(t, client).withItems("UID", "BODY[HEADER]").fetch("1")
			result.forSeqNum(1, func(builder *validatorBuilder) {
				// The header value appears twice because we don't delete existing headers, we only add new ones.
				builder.ignoreFlags().wantSection("BODY[HEADER]", fmt.Sprintf("%v: 2", ids.InternalIDKey), fmt.Sprintf("%v: 1", ids.InternalIDKey), "To: foo@bar.com\r\n")
				builder.wantUID(2)
			})
			result.checkAndRequireMessageCount(1)
		}
	})
}
