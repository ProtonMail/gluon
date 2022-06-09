package remote

import (
	"testing"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/stretchr/testify/require"
)

func TestLoadSaveOps(t *testing.T) {
	data := userSerializedData{
		ConnMetadataStore: newConnMetadataStore(),
	}

	data.PendingOps = []operation{
		&OpMailboxCreate{OperationBase: OperationBase{MetadataID: ConnMetadataID(204)}, TempID: "tempID", Name: []string{"name"}},
		&OpMailboxDelete{OperationBase: OperationBase{MetadataID: ConnMetadataID(2)}, MBoxID: "tempID"},
		&OpMailboxUpdate{OperationBase: OperationBase{MetadataID: ConnMetadataID(3)}, MBoxID: "tempID", Name: []string{"other"}},
		&OpMessageCreate{
			OperationBase: OperationBase{MetadataID: ConnMetadataID(4)},
			TempID:        "tempID",
			MBoxID:        "mboxID",
			Literal:       []byte("literal"),
			Flags:         imap.NewFlagSet(imap.FlagAnswered, imap.FlagDeleted),
			Date:          time.Date(2022, time.May, 16, 10, 0o0, 0o0, 0o0, time.UTC),
		},
		&OpMessageAdd{OperationBase: OperationBase{MetadataID: ConnMetadataID(6)}, MessageIDs: []string{"messageID1", "messageID2"}, MBoxID: "mboxID"},
		&OpMessageRemove{OperationBase: OperationBase{MetadataID: ConnMetadataID(7)}, MessageIDs: []string{"messageID1", "messageID2"}, MBoxID: "mboxID"},
		&OpMessageSeen{OperationBase: OperationBase{MetadataID: ConnMetadataID(8)}, MessageIDs: []string{"messageID1", "messageID2"}},
		&OpMessageFlagged{OperationBase: OperationBase{MetadataID: ConnMetadataID(9)}, MessageIDs: []string{"messageID1", "messageID2"}},
		&OpConnMetadataStoreDelete{OperationBase: OperationBase{MetadataID: ConnMetadataID(10)}},
		&OpConnMetadataStoreCreate{OperationBase: OperationBase{MetadataID: ConnMetadataID(11)}},
		&OpConnMetadataStoreSetValue{OperationBase: OperationBase{MetadataID: ConnMetadataID(12)}, Key: "10", Value: "foobar"},
	}

	const metadataID = ConnMetadataID(20)

	data.ConnMetadataStore.CreateStore(metadataID)
	data.ConnMetadataStore.SetValue(metadataID, "30", "bar")

	// Save the operations.
	b, err := data.saveToBytes()
	require.NoError(t, err)

	loadedData := userSerializedData{
		ConnMetadataStore: newConnMetadataStore(),
	}
	// Load the operations.
	err = loadedData.loadFromBytes(b)
	require.NoError(t, err)

	// They should be the same as what we originally saved.
	require.Equal(t, data, loadedData)
}
