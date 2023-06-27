package sqlite3

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3/utils"
	v0 "github.com/ProtonMail/gluon/internal/db_impl/sqlite3/v0"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestMigration_VersionTooHigh(t *testing.T) {
	testDir := t.TempDir()

	setup := func() {
		client, _, err := NewClient(testDir, "foo", false, false)
		require.NoError(t, err)

		ctx := context.Background()
		require.NoError(t, client.Init(ctx, &imap.IncrementalUIDValidityGenerator{}))

		defer func() {
			require.NoError(t, client.Close())
		}()

		// For version to very high value
		require.NoError(t, client.wrapTx(ctx, func(ctx context.Context, tx *sql.Tx, entry *logrus.Entry) error {
			qw := utils.TXWrapper{TX: tx}
			return updateDBVersion(ctx, qw, 999999)
		}))
	}

	setup()

	client, _, err := NewClient(testDir, "foo", false, false)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, client.Close())
	}()

	err = client.Init(context.Background(), imap.DefaultEpochUIDValidityGenerator())
	require.Error(t, err)
	require.True(t, errors.Is(err, db.ErrInvalidDatabaseVersion))
}

func TestRunMigrations(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	testDir := t.TempDir()

	uidGenerator := &imap.IncrementalUIDValidityGenerator{}

	testData := newTestData(uidGenerator)

	// Fill v0 database.
	prepareV0Database(t, testDir, "foo", testData, uidGenerator)

	// First run, incurs migration.
	runAndValidateDB(t, testDir, "foo", testData, uidGenerator)

	// Second run, no migration.
	runAndValidateDB(t, testDir, "foo", testData, uidGenerator)
}

func runAndValidateDB(t *testing.T, testDir, user string, testData *testData, uidGenerator imap.UIDValidityGenerator) {
	// create client and run all migrations.
	client, _, err := NewClient(testDir, "foo", false, false)
	require.NoError(t, err)

	ctx := context.Background()

	defer func() {
		require.NoError(t, client.Close())
	}()

	require.NoError(t, client.Init(ctx, uidGenerator))

	require.NoError(t, client.Read(ctx, func(ctx context.Context, rd db.ReadOnly) error {
		// Check if mailboxes contain all data.
		for _, m := range testData.mailboxes {
			dbMBox, err := rd.GetMailboxByRemoteID(ctx, m.RemoteID)
			require.NoError(t, err)
			require.Equal(t, m.RemoteID, dbMBox.RemoteID)
			require.Equal(t, m.Name, dbMBox.Name)
			require.Equal(t, m.Subscribed, dbMBox.Subscribed)
			require.NotEqual(t, m.UIDValidity, dbMBox.UIDValidity)

			// Check Flags.
			{
				flags, err := rd.GetMailboxFlags(ctx, dbMBox.ID)
				require.NoError(t, err)
				require.True(t, flags.Equals(m.Flags))
			}

			// Check Perm Flags.
			{
				flags, err := rd.GetMailboxPermanentFlags(ctx, dbMBox.ID)
				require.NoError(t, err)
				require.True(t, flags.Equals(m.PermanentFlags))
			}

			// Check Attributes
			{
				attr, err := rd.GetMailboxAttributes(ctx, dbMBox.ID)
				require.NoError(t, err)
				require.True(t, attr.Equals(m.Attributes))
			}
		}

		// Check if messages contain all data.
		for _, m := range testData.messages {
			dbMsg, err := rd.GetMessageNoEdges(ctx, m.ID)
			require.NoError(t, err)

			require.Equal(t, m.ID, dbMsg.ID)
			require.Equal(t, m.RemoteID, dbMsg.RemoteID)
			require.Equal(t, m.Deleted, dbMsg.Deleted)
			require.Equal(t, m.Body, dbMsg.Body)
			require.Equal(t, m.BodyStructure, dbMsg.BodyStructure)
			require.Equal(t, m.Envelope, dbMsg.Envelope)
			require.Equal(t, m.Size, dbMsg.Size)
			require.Equal(t, m.Date, dbMsg.Date)

			// Check flags
			{
				flags, err := rd.GetMessagesFlags(ctx, []imap.InternalMessageID{m.ID})
				require.Len(t, flags, 1)
				require.NoError(t, err)
				require.True(t, flags[0].FlagSet.Equals(m.Flags))

			}
		}

		// Check messages in mailboxes.
		for _, m := range testData.messagesToMBox {
			msg, err := rd.GetMailboxMessageForNewSnapshot(ctx, m.mboxID)
			require.NoError(t, err)

			idx := slices.IndexFunc(msg, func(result db.SnapshotMessageResult) bool {
				return result.InternalID == m.messageID
			})

			require.True(t, idx >= 0)

			require.Equal(t, m.recent, msg[idx].Recent)
			require.Equal(t, m.deleted, msg[idx].Deleted)
			require.Equal(t, m.uid, msg[idx].UID)
		}

		return nil
	}))
}

type messageToMBox struct {
	messageID imap.InternalMessageID
	mboxID    imap.InternalMailboxID
	uid       imap.UID
	deleted   bool
	recent    bool
}

type mailbox struct {
	ID             imap.InternalMailboxID
	RemoteID       imap.MailboxID
	Name           string
	UIDValidity    imap.UID
	Subscribed     bool
	Flags          imap.FlagSet
	PermanentFlags imap.FlagSet
	Attributes     imap.FlagSet
}

type message struct {
	ID            imap.InternalMessageID
	RemoteID      imap.MessageID
	Date          time.Time
	Size          int
	Body          string
	BodyStructure string
	Envelope      string
	Deleted       bool
	Flags         imap.FlagSet
}

type testData struct {
	mailboxes      []mailbox
	messages       []message
	messagesToMBox []messageToMBox
}

func newTestData(generator imap.UIDValidityGenerator) *testData {
	newUID := func() imap.UID {
		uid, err := generator.Generate()
		if err != nil {
			panic(err)
		}

		return uid
	}

	mailboxes := []mailbox{
		{
			ID:             1,
			RemoteID:       "RemoteID1",
			Name:           "Foobar",
			UIDValidity:    newUID(),
			Subscribed:     true,
			Flags:          imap.NewFlagSet("Flag1", "Flag2"),
			PermanentFlags: imap.NewFlagSet("PermFlag1", "PermFlag2"),
			Attributes:     imap.NewFlagSet("Attr1"),
		},
		{
			ID:             2,
			RemoteID:       "RemoteID2",
			Name:           "Abracadabra",
			UIDValidity:    newUID(),
			Subscribed:     true,
			Flags:          imap.NewFlagSet("Flag3", "Flag4"),
			PermanentFlags: imap.NewFlagSet("PermFlag3"),
			Attributes:     imap.NewFlagSet("Attr2", "Attr3"),
		},
		{
			ID:             3,
			RemoteID:       "RemoteID3",
			Name:           "Mips",
			UIDValidity:    newUID(),
			Subscribed:     false,
			Flags:          imap.NewFlagSet(),
			PermanentFlags: imap.NewFlagSet(),
			Attributes:     imap.NewFlagSet(),
		},
	}

	messages := []message{
		{
			ID:            imap.NewInternalMessageID(),
			RemoteID:      "MessageID2",
			Date:          time.Now().UTC(),
			Size:          512,
			Body:          "Message Body 2",
			BodyStructure: "Message Structure 2",
			Envelope:      "Message Envelope 2",
			Deleted:       false,
			Flags:         imap.NewFlagSet("\\Seen"),
		},
		{
			ID:            imap.NewInternalMessageID(),
			RemoteID:      "MessageID1",
			Date:          time.Now().UTC(),
			Size:          1024,
			Body:          "Message Body 1",
			BodyStructure: "Message Structure 1",
			Envelope:      "Message Envelope 1",
			Deleted:       false,
			Flags:         imap.NewFlagSet("\\Seen", "\\Flagged"),
		},
		{
			ID:            imap.NewInternalMessageID(),
			RemoteID:      "MessageID3",
			Date:          time.Now().UTC(),
			Size:          64,
			Body:          "Message Body 3",
			BodyStructure: "Message Structure 3",
			Envelope:      "Message Envelope 3",
			Deleted:       true,
			Flags:         imap.NewFlagSet(),
		},
	}

	messagesToMbox := []messageToMBox{
		{
			messageID: messages[0].ID,
			mboxID:    mailboxes[0].ID,
			uid:       1,
			deleted:   false,
			recent:    true,
		},
		{
			messageID: messages[1].ID,
			mboxID:    mailboxes[0].ID,
			uid:       2,
			deleted:   true,
			recent:    false,
		},
		{
			messageID: messages[2].ID,
			mboxID:    mailboxes[1].ID,
			uid:       1,
			deleted:   false,
			recent:    false,
		},
	}

	return &testData{mailboxes: mailboxes, messages: messages, messagesToMBox: messagesToMbox}
}

func prepareV0Database(t *testing.T, dir, user string, testData *testData, uidGenerator imap.UIDValidityGenerator) {
	client, _, err := NewClient(dir, "foo", false, false)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, client.Close())
	}()

	ctx := context.Background()

	// Run Migration 0.
	require.NoError(t, client.wrapTx(ctx, func(ctx context.Context, tx *sql.Tx, entry *logrus.Entry) error {
		qw := utils.TXWrapper{TX: tx}
		v0Migration := v0.Migration{}

		require.NoError(t, v0Migration.Run(ctx, qw, uidGenerator))

		// Fill in base data.
		// Mailboxes & Mbox Flags
		{
			// Mailboxes.
			{
				query := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`, `%v`, `%v`) VALUES %v",
					v0.MailboxesTableName,
					v0.MailboxesFieldRemoteID,
					v0.MailboxesFieldName,
					v0.MailboxesFieldUIDValidity,
					v0.MailboxesFieldSubscribed,
					strings.Join(xslices.Repeat("(?, ?, ?, ?)", len(testData.mailboxes)), ","),
				)
				args := make([]any, 0, len(testData.mailboxes)*4)
				for _, mbox := range testData.mailboxes {
					args = append(args, mbox.RemoteID, mbox.Name, mbox.UIDValidity, mbox.Subscribed)
				}

				require.NoError(t, utils.ExecQueryAndCheckUpdatedNotZero(ctx, qw, query, args...))
			}
			// Flags.
			for _, m := range testData.mailboxes {
				if len(m.Flags) == 0 {
					continue
				}

				query := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES %v",
					v0.MailboxFlagsTableName,
					v0.MailboxFlagsFieldMailboxID,
					v0.MailboxFlagsFieldValue,
					strings.Join(xslices.Repeat("(?, ?)", len(m.Flags)), ","),
				)

				args := make([]any, 0, len(m.Flags)*2)

				for _, f := range m.Flags.ToSliceUnsorted() {
					args = append(args, m.ID, f)
				}

				require.NoError(t, utils.ExecQueryAndCheckUpdatedNotZero(ctx, qw, query, args...))
			}
			// Perm flags.
			for _, m := range testData.mailboxes {
				if len(m.PermanentFlags) == 0 {
					continue
				}

				query := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES %v",
					v0.MailboxPermFlagsTableName,
					v0.MailboxPermFlagsFieldMailboxID,
					v0.MailboxPermFlagsFieldValue,
					strings.Join(xslices.Repeat("(?, ?)", len(m.PermanentFlags)), ","),
				)

				args := make([]any, 0, len(m.PermanentFlags)*2)

				for _, f := range m.PermanentFlags.ToSliceUnsorted() {
					args = append(args, m.ID, f)
				}

				require.NoError(t, utils.ExecQueryAndCheckUpdatedNotZero(ctx, qw, query, args...))
			}
			// Attributes
			for _, m := range testData.mailboxes {
				if len(m.Attributes) == 0 {
					continue
				}

				query := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES %v",
					v0.MailboxAttrsTableName,
					v0.MailboxAttrsFieldMailboxID,
					v0.MailboxAttrsFieldValue,
					strings.Join(xslices.Repeat("(?, ?)", len(m.Attributes)), ","),
				)

				args := make([]any, 0, len(m.Attributes)*2)

				for _, f := range m.Attributes.ToSliceUnsorted() {
					args = append(args, m.ID, f)
				}

				require.NoError(t, utils.ExecQueryAndCheckUpdatedNotZero(ctx, qw, query, args...))
			}
		}

		// Messages & Message Flags.
		{
			// Messages
			{
				query := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`, `%v`, `%v`, `%v`, `%v`, `%v`, `%v`) VALUES %v",
					v0.MessagesTableName,
					v0.MessagesFieldID,
					v0.MessagesFieldRemoteID,
					v0.MessagesFieldDate,
					v0.MessagesFieldSize,
					v0.MessagesFieldBody,
					v0.MessagesFieldBodyStructure,
					v0.MessagesFieldEnvelope,
					v0.MessagesFieldDeleted,
					strings.Join(xslices.Repeat("(?, ?, ?, ?, ?, ?, ?, ?)", len(testData.mailboxes)), ","),
				)

				args := make([]any, 0, len(testData.messages)*8)

				for _, m := range testData.messages {
					args = append(args, m.ID, m.RemoteID, m.Date, m.Size, m.Body, m.BodyStructure, m.Envelope, m.Deleted)
				}

				require.NoError(t, utils.ExecQueryAndCheckUpdatedNotZero(ctx, qw, query, args...))
			}

			// Message flags
			{
				query := fmt.Sprintf("INSERT INTO %v (`%v`, `%v`) VALUES %v",
					v0.MessageFlagsTableName,
					v0.MessageFlagsFieldMessageID,
					v0.MessageFlagsFieldValue,
					strings.Join(xslices.Repeat("(?, ?)", len(testData.mailboxes)), ","),
				)

				args := make([]any, 0, len(testData.messages)*2)

				for _, m := range testData.messages {
					for _, f := range m.Flags.ToSliceUnsorted() {
						args = append(args, m.ID, f)
					}
				}

				require.NoError(t, utils.ExecQueryAndCheckUpdatedNotZero(ctx, qw, query, args...))
			}

		}

		// UIDs
		{
			query := fmt.Sprintf("INSERT INTO %v (`%v`,`%v`,`%v`, `%v`, `%v`) VALUES %v",
				v0.UIDsTableName,
				v0.UIDsFieldMessageID,
				v0.UIDsFieldMailboxID,
				v0.UIDsFieldUID,
				v0.UIDsFieldDeleted,
				v0.UIDsFieldRecent,
				strings.Join(xslices.Repeat("(?, ?, ?, ?, ?)", len(testData.mailboxes)), ","),
			)

			args := make([]any, 0, len(testData.messagesToMBox))

			for _, m := range testData.messagesToMBox {
				args = append(args, m.messageID, m.mboxID, m.uid, m.deleted, m.recent)
			}

			require.NoError(t, utils.ExecQueryAndCheckUpdatedNotZero(ctx, qw, query, args...))
		}

		return nil
	}))
}
