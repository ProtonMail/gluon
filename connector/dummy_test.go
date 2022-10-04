package connector

import (
	"testing"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/stretchr/testify/require"
)

const defaultPeriod = time.Second

var (
	defaultFlags          = imap.NewFlagSet(imap.FlagSeen, imap.FlagFlagged, imap.FlagDeleted)
	defaultPermanentFlags = imap.NewFlagSet(imap.FlagSeen, imap.FlagFlagged, imap.FlagDeleted)
	defaultAttributes     = imap.NewFlagSet()
)

func TestDummyConnector_validateUpdate(t *testing.T) {
	conn := NewDummy(
		[]string{"username"},
		[]byte("password"),
		defaultPeriod,
		defaultFlags,
		defaultPermanentFlags,
		defaultAttributes,
	)

	go func() {
		for update := range conn.GetUpdates() {
			update.Done()
		}
	}()

	require.NoError(t, conn.validateUpdate([]string{"something"}, []string{"other"}))

	conn.SetFolderPrefix("Folders")
	require.NoError(t, conn.validateUpdate([]string{"Folders", "something"}, []string{"Folders", "other"}))
	require.NoError(t, conn.validateUpdate([]string{"Folders", "something"}, []string{"Folders", "other", "long"}))
	require.Error(t, conn.validateUpdate([]string{"Folders", "something"}, []string{"other"}))
	require.Error(t, conn.validateUpdate([]string{"something"}, []string{"Folders", "other"}))

	conn.SetLabelPrefix("Labels")
	require.NoError(t, conn.validateUpdate([]string{"Labels", "something"}, []string{"Labels", "other"}))
	require.NoError(t, conn.validateUpdate([]string{"Labels", "something"}, []string{"Labels", "other", "long"}))
	require.Error(t, conn.validateUpdate([]string{"Labels", "something"}, []string{"other"}))
	require.Error(t, conn.validateUpdate([]string{"something"}, []string{"Labels", "other"}))
	require.Error(t, conn.validateUpdate([]string{"Folders", "something"}, []string{"Labels", "something"}))
}
