package tests

import (
	"bytes"
	"context"
	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"os"
	"sync"
	"testing"
	"time"

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
		const (
			validityUid = uint32(1)
		)
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

func (r *returnSameRemoteIDConnector) CreateMessage(ctx context.Context, mboxID imap.LabelID, literal []byte, flags imap.FlagSet, date time.Time) (imap.Message, []byte, error) {
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
