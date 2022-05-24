package remote

import (
	"testing"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/stretchr/testify/require"
)

func TestLoadSaveOps(t *testing.T) {
	wantOps := []operation{
		&OpMailboxCreate{TempID: "tempID", Name: []string{"name"}},
		&OpMailboxDelete{MBoxID: "tempID"},
		&OpMailboxUpdate{MBoxID: "tempID", Name: []string{"other"}},
		&OpMessageCreate{
			TempID:  "tempID",
			MBoxID:  "mboxID",
			Literal: []byte("literal"),
			Flags:   imap.NewFlagSet(imap.FlagAnswered, imap.FlagDeleted),
			Date:    time.Date(2022, time.May, 16, 10, 0o0, 0o0, 0o0, time.UTC),
		},
		&OpMessageAdd{MessageIDs: []string{"messageID1", "messageID2"}, MBoxID: "mboxID"},
		&OpMessageRemove{MessageIDs: []string{"messageID1", "messageID2"}, MBoxID: "mboxID"},
		&OpMessageSeen{MessageIDs: []string{"messageID1", "messageID2"}},
		&OpMessageFlagged{MessageIDs: []string{"messageID1", "messageID2"}},
	}

	// Save the operations.
	b, err := saveOps(wantOps)
	require.NoError(t, err)

	// Load the operations.
	haveOps, err := loadOps(b)
	require.NoError(t, err)

	// They should be the same as what we originally saved.
	require.Equal(t, wantOps, haveOps)
}
