package tests

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/connector"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/ids"
	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestRecoveryMBoxNotVisibleWhenEmpty(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, s *testSession) {
		c.C(`A103 LIST "" *`)
		c.S(`* LIST (\Unmarked) "/" "INBOX"`)
		c.OK(`A103`)
	})
}

func TestRecoveryMBoxVisibleWhenNotEmpty(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t, withConnectorBuilder(&failAppendLabelConnectorBuilder{})), func(c *testConnection, s *testSession) {
		c.doAppend("INBOX", "INBOX", "To: Test@test.com").expect("NO")
		c.C(`A103 LIST "" *`)
		c.S(`* LIST (\Unmarked) "/" "INBOX"`,
			fmt.Sprintf(`* LIST (\Marked \Noinferiors) "/" "%v"`, ids.GluonRecoveryMailboxName),
		)
		c.OK(`A103`)
	})
}

func TestRecoveryMBoxCanNotBeCreated(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.Error(t, client.Create(ids.GluonRecoveryMailboxNameLowerCase))
		require.Error(t, client.Create(ids.GluonRecoveryMailboxName))
		require.Error(t, client.Create(fmt.Sprintf("%v/sub", ids.GluonRecoveryMailboxName)))
	})
}

func TestRecoveryMBoxCanNotBeRenamed(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.Error(t, client.Rename(ids.GluonRecoveryMailboxName, "SomethingElse"))
		require.Error(t, client.Rename("INBOX", ids.GluonRecoveryMailboxName))
	})
}

func TestRecoveryMBoxCanNotBeAppended(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.Error(t, client.Append(ids.GluonRecoveryMailboxName, nil, time.Now(), bytes.NewReader([]byte("RandomGibberish"))))
	})
}

func TestRecoveryMBoxCanNotBeMovedOrCopiedInto(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, s *testSession) {
		const mboxName = "Foo"
		mboxID := s.mailboxCreated("user", []string{mboxName})
		s.messageCreated("user", mboxID, []byte("To: Test@test.com"), time.Now())
		s.flush("user")
		status, err := client.Select(mboxName, false)
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages)

		require.Error(t, client.Move(createSeqSet("1"), ids.GluonRecoveryMailboxName))
		require.Error(t, client.Copy(createSeqSet("1"), ids.GluonRecoveryMailboxName))
	})
}

func TestRecoveryMBoxCanBeMovedOutOf(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withConnectorBuilder(&disableRemoveFromMailboxBuilder{})), func(client *client.Client, s *testSession) {
		// Insert first message, fails.
		require.Error(t, doAppendWithClient(client, "INBOX", "To: Test@test.com", time.Now()))
		status, err := client.Status(ids.GluonRecoveryMailboxName, []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages)
		{
			_, err := client.Select(ids.GluonRecoveryMailboxName, false)
			require.NoError(t, err)
			newFetchCommand(t, client).withItems("BODY[]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
				builder.ignoreFlags()
				builder.wantSection("BODY[]", "To: Test@test.com")
			}).checkAndRequireMessageCount(1)
		}

		// Move message.
		require.NoError(t, client.Move(createSeqSet("1"), "INBOX"))

		// Check state.
		status, err = client.Status("INBOX", []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages)
		status, err = client.Status(ids.GluonRecoveryMailboxName, []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(0), status.Messages)

		{
			_, err := client.Select("INBOX", false)
			require.NoError(t, err)
			// Check that message has the new internal ID header.
			newFetchCommand(t, client).withItems("BODY[]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
				builder.ignoreFlags()
				builder.wantSectionAndSkipGLUONHeaderOrPanic("BODY[]", "To: Test@test.com")
			}).checkAndRequireMessageCount(1)
		}
	})
}

func TestRecoveryMBoxCanBeCopiedOutOf(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withConnectorBuilder(&disableRemoveFromMailboxBuilder{})), func(client *client.Client, s *testSession) {
		// Insert first message, fails.
		require.Error(t, doAppendWithClient(client, "INBOX", "To: Test@test.com", time.Now()))
		status, err := client.Status(ids.GluonRecoveryMailboxName, []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages)
		{
			_, err := client.Select(ids.GluonRecoveryMailboxName, false)
			require.NoError(t, err)
			newFetchCommand(t, client).withItems("BODY[]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
				builder.ignoreFlags()
				builder.wantSection("BODY[]", "To: Test@test.com")
			}).checkAndRequireMessageCount(1)
		}

		// Copy message.
		require.NoError(t, client.Copy(createSeqSet("1"), "INBOX"))

		// Validate state.
		status, err = client.Status("INBOX", []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages)
		status, err = client.Status(ids.GluonRecoveryMailboxName, []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages)

		{
			_, err := client.Select("INBOX", false)
			require.NoError(t, err)
			// Check that message has the new internal ID header.
			newFetchCommand(t, client).withItems("BODY[]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
				builder.ignoreFlags()
				builder.wantSectionAndSkipGLUONHeaderOrPanic("BODY[]", "To: Test@test.com")
			}).checkAndRequireMessageCount(1)
		}
	})
}

func TestRecoveryMBoxCanBeExpunged(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withConnectorBuilder(&disableRemoveFromMailboxBuilder{})), func(client *client.Client, s *testSession) {
		// Insert first message, fails.
		require.Error(t, doAppendWithClient(client, "INBOX", "To: Test@test.com", time.Now()))
		// Execute expunge
		status, err := client.Select(ids.GluonRecoveryMailboxName, false)
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages)
		require.NoError(t, client.Store(createSeqSet("1"), goimap.AddFlags, []interface{}{goimap.DeletedFlag}, nil))
		require.NoError(t, client.Expunge(nil))
		status, err = client.Status(ids.GluonRecoveryMailboxName, []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(0), status.Messages)
	})
}

func TestFailedAppendEndsInRecovery(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withConnectorBuilder(&failAppendLabelConnectorBuilder{})), func(client *client.Client, s *testSession) {
		{
			status, err := client.Status(ids.GluonRecoveryMailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(0), status.Messages)
		}

		status, err := client.Select("INBOX", false)
		require.NoError(t, err)
		require.Equal(t, uint32(0), status.Messages)
		require.Error(t, doAppendWithClient(client, "INBOX", "To: Foo@bar.com", time.Now()))

		{
			status, err := client.Status(ids.GluonRecoveryMailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(1), status.Messages)
		}
		{
			status, err := client.Status("INBOX", []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(0), status.Messages)
		}

		{
			_, err := client.Select(ids.GluonRecoveryMailboxName, false)
			require.NoError(t, err)
			// Check that no custom headers are appended to the message.
			newFetchCommand(t, client).withItems("BODY[]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
				builder.ignoreFlags()
				builder.wantSection("BODY[]", "To: Foo@bar.com")
			}).checkAndRequireMessageCount(1)
		}
	})
}

func TestFailedAppendAreDedupedInRecoveryMailbox(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withConnectorBuilder(&failAppendLabelConnectorBuilder{})), func(client *client.Client, s *testSession) {
		{
			status, err := client.Status(ids.GluonRecoveryMailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(0), status.Messages)
		}

		status, err := client.Select("INBOX", false)
		require.NoError(t, err)
		require.Equal(t, uint32(0), status.Messages)
		require.Error(t, doAppendWithClient(client, "INBOX", "To: Foo@bar.com", time.Now()))
		require.Error(t, doAppendWithClient(client, "INBOX", "To: Foo@bar.com", time.Now()))
		require.Error(t, doAppendWithClient(client, "INBOX", "To: Bar@bar.com", time.Now()))

		{
			status, err := client.Status(ids.GluonRecoveryMailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(2), status.Messages)
		}
		{
			status, err := client.Status("INBOX", []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(0), status.Messages)
		}

		{
			_, err := client.Select(ids.GluonRecoveryMailboxName, false)
			require.NoError(t, err)
			// Check that no custom headers are appended to the message.
			newFetchCommand(t, client).withItems("BODY[]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
				builder.ignoreFlags()
				builder.wantSection("BODY[]", "To: Foo@bar.com")
			}).checkAndRequireMessageCount(1)
		}
	})
}

func TestRecoveryMailboxOnlyReportsOnFirstDedupedMessage(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withConnectorBuilder(&failAppendLabelConnectorBuilder{})), func(client *client.Client, s *testSession) {
		{
			status, err := client.Status(ids.GluonRecoveryMailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(0), status.Messages)
		}

		status, err := client.Select("INBOX", false)
		require.NoError(t, err)
		require.Equal(t, uint32(0), status.Messages)
		require.Error(t, doAppendWithClientFromFile(t, client, "INBOX", "testdata/multipart-mixed.eml", time.Now()))
		require.Error(t, doAppendWithClientFromFile(t, client, "INBOX", "testdata/multipart-mixed2.eml", time.Now()))

		{
			status, err := client.Status(ids.GluonRecoveryMailboxName, []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(1), status.Messages)
		}
		{
			status, err := client.Status("INBOX", []goimap.StatusItem{goimap.StatusMessages})
			require.NoError(t, err)
			require.Equal(t, uint32(0), status.Messages)
		}

		{
			reports := s.reporter.getReports()
			require.Equal(t, 1, len(reports))
		}
	})
}

func TestRecoveryMBoxCanBeCopiedOutOfDedup(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withConnectorBuilder(&recoveryDedupConnectorConnectorBuilder{})), func(client *client.Client, s *testSession) {
		// Insert first message, fails.
		require.Error(t, doAppendWithClient(client, "INBOX", "To: Test@test.com", time.Now()))
		{
			_, err := client.Select(ids.GluonRecoveryMailboxName, false)
			require.NoError(t, err)
			newFetchCommand(t, client).withItems("BODY[]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
				builder.ignoreFlags()
				builder.wantSection("BODY[]", "To: Test@test.com")
			}).checkAndRequireMessageCount(1)
		}

		// Insert same message, succeeds.
		require.NoError(t, doAppendWithClient(client, "INBOX", "To: Test@test.com", time.Now()))

		{
			_, err := client.Select("INBOX", false)
			require.NoError(t, err)
			newFetchCommand(t, client).withItems("BODY[]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
				builder.ignoreFlags()
				builder.wantSectionAndSkipGLUONHeaderOrPanic("BODY[]", "To: Test@test.com")
			}).checkAndRequireMessageCount(1)
		}

		msgInInbox := fetchMessageBody(t, client, 1)

		// Copy message out of recovery, triggers insert will return the same ID.
		status, err := client.Select(ids.GluonRecoveryMailboxName, false)
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages)
		require.NoError(t, client.Copy(createSeqSet("1"), "INBOX"))
		status, err = client.Status("INBOX", []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages)
		status, err = client.Status(ids.GluonRecoveryMailboxName, []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages)

		// Check that no new message was created in INBOX as we already have this message available there.
		{
			_, err := client.Select("INBOX", false)
			require.NoError(t, err)
			// Check that message has the new internal ID header.
			newFetchCommand(t, client).withItems("BODY[]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
				builder.ignoreFlags()
				builder.wantSection("BODY[]", msgInInbox)
			}).checkAndRequireMessageCount(1)
		}
	})
}

func TestRecoveryMBoxCanBeMovedOutOfDedup(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t, withConnectorBuilder(&recoveryDedupConnectorConnectorBuilder{})), func(client *client.Client, s *testSession) {
		// Insert first message, fails.
		require.Error(t, doAppendWithClient(client, "INBOX", "To: Test@test.com", time.Now()))
		{
			_, err := client.Select(ids.GluonRecoveryMailboxName, false)
			require.NoError(t, err)
			newFetchCommand(t, client).withItems("BODY[]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
				builder.ignoreFlags()
				builder.wantSection("BODY[]", "To: Test@test.com")
			}).checkAndRequireMessageCount(1)
		}

		// Insert same message, succeeds.
		require.NoError(t, doAppendWithClient(client, "INBOX", "To: Test@test.com", time.Now()))

		{
			_, err := client.Select("INBOX", false)
			require.NoError(t, err)
			newFetchCommand(t, client).withItems("BODY[]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
				builder.ignoreFlags()
				builder.wantSectionAndSkipGLUONHeaderOrPanic("BODY[]", "To: Test@test.com")
			}).checkAndRequireMessageCount(1)
		}

		msgInInbox := fetchMessageBody(t, client, 1)

		// Copy message out of recovery, triggers insert will return the same ID.
		status, err := client.Select(ids.GluonRecoveryMailboxName, false)
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages)
		require.NoError(t, client.Move(createSeqSet("1"), "INBOX"))
		status, err = client.Status("INBOX", []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(1), status.Messages)
		status, err = client.Status(ids.GluonRecoveryMailboxName, []goimap.StatusItem{goimap.StatusMessages})
		require.NoError(t, err)
		require.Equal(t, uint32(0), status.Messages)

		// Check that no new message was created in INBOX as we already have this message available there.
		{
			_, err := client.Select("INBOX", false)
			require.NoError(t, err)
			// Check that message has the new internal ID header.
			newFetchCommand(t, client).withItems("BODY[]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
				builder.ignoreFlags()
				builder.wantSection("BODY[]", msgInInbox)
			}).checkAndRequireMessageCount(1)
		}
	})
}

// disableRemoveFromMailboxConnector fails the first append and panics if move or remove takes place on the
// connector.
type disableRemoveFromMailboxConnector struct {
	*connector.Dummy
	createFailed bool
}

func (r *disableRemoveFromMailboxConnector) CreateMessage(
	ctx context.Context,
	mboxID imap.MailboxID,
	literal []byte,
	flags imap.FlagSet,
	date time.Time) (imap.Message, []byte, error) {
	if !r.createFailed {
		r.createFailed = true
		return imap.Message{}, nil, fmt.Errorf("failed")
	}

	return r.Dummy.CreateMessage(ctx, mboxID, literal, flags, date)
}

func (r *disableRemoveFromMailboxConnector) RemoveMessagesFromMailbox(
	_ context.Context,
	_ []imap.MessageID,
	_ imap.MailboxID,
) error {
	panic("Should not be called")
}

func (r *disableRemoveFromMailboxConnector) MoveMessages(
	_ context.Context,
	_ []imap.MessageID,
	_ imap.MailboxID,
	_ imap.MailboxID,
) (bool, error) {
	panic("Should not be called")
}

type disableRemoveFromMailboxBuilder struct{}

func (disableRemoveFromMailboxBuilder) New(usernames []string, password []byte, period time.Duration, flags, permFlags, attrs imap.FlagSet) Connector {
	return &disableRemoveFromMailboxConnector{
		Dummy: connector.NewDummy(usernames, password, period, flags, permFlags, attrs),
	}
}

// recoveryDedupConnector fails the first CreateMessage call and then returns the same remote id for all
// the next messages which are created to simulated de-duplication.
type recoveryDedupConnector struct {
	*connector.Dummy

	firstMessage   bool
	messageCreated bool
	createdMessage imap.Message
	messageLiteral []byte
}

func (r *recoveryDedupConnector) CreateMessage(
	ctx context.Context,
	mboxID imap.MailboxID,
	literal []byte,
	flags imap.FlagSet,
	date time.Time) (imap.Message, []byte, error) {
	if r.firstMessage {
		r.firstMessage = false
		return imap.Message{}, nil, fmt.Errorf("failed")
	}

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

func (r *recoveryDedupConnector) AddMessagesToMailbox(
	ctx context.Context,
	ids []imap.MessageID,
	mboxID imap.MailboxID,
) error {
	if r.firstMessage {
		return fmt.Errorf("failed")
	}

	return r.Dummy.AddMessagesToMailbox(ctx, ids, mboxID)
}

func (r *recoveryDedupConnector) MoveMessagesFromMailbox(
	_ context.Context,
	_ []imap.MessageID,
	_ imap.MailboxID,
	_ imap.MailboxID,
) error {
	return fmt.Errorf("failed")
}

type recoveryDedupConnectorConnectorBuilder struct{}

func (recoveryDedupConnectorConnectorBuilder) New(usernames []string, password []byte, period time.Duration, flags, permFlags, attrs imap.FlagSet) Connector {
	return &recoveryDedupConnector{
		Dummy:        connector.NewDummy(usernames, password, period, flags, permFlags, attrs),
		firstMessage: true,
	}
}

// failAppendLabelConnector simulate Create Message failures and also ensures that no calls to Add or Move can take place.
type failAppendLabelConnector struct {
	*connector.Dummy
}

func (r *failAppendLabelConnector) CreateMessage(
	_ context.Context,
	_ imap.MailboxID,
	_ []byte,
	_ imap.FlagSet,
	_ time.Time) (imap.Message, []byte, error) {
	return imap.Message{}, nil, fmt.Errorf("failed")
}

func (r *failAppendLabelConnector) AddMessagesToMailbox(
	_ context.Context,
	_ []imap.MessageID,
	_ imap.MailboxID,
) error {
	return fmt.Errorf("failed")
}

func (r *failAppendLabelConnector) MoveMessagesFromMailbox(
	_ context.Context,
	_ []imap.MessageID,
	_ imap.MailboxID,
	_ imap.MailboxID,
) error {
	return fmt.Errorf("failed")
}

type failAppendLabelConnectorBuilder struct{}

func (failAppendLabelConnectorBuilder) New(usernames []string, password []byte, period time.Duration, flags, permFlags, attrs imap.FlagSet) Connector {
	return &failAppendLabelConnector{
		Dummy: connector.NewDummy(usernames, password, period, flags, permFlags, attrs),
	}
}
