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

func TestDummyConnector_ValidateCreate(t *testing.T) {
	conn := NewDummy(
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

	{
		_, _, _, err := conn.ValidateCreate([]string{"something"})
		require.NoError(t, err)
	}

	conn.SetFolderPrefix("Folders")

	{
		_, _, _, err := conn.ValidateCreate([]string{"Folders", "something"})
		require.NoError(t, err)
	}

	{
		_, _, _, err := conn.ValidateCreate([]string{"something"})
		require.NoError(t, err)
	}

	conn.SetLabelPrefix("Labels")

	{
		_, _, _, err := conn.ValidateCreate([]string{"Folders", "something"})
		require.NoError(t, err)
	}

	{
		_, _, _, err := conn.ValidateCreate([]string{"Labels", "something"})
		require.NoError(t, err)
	}

	{
		_, _, _, err := conn.ValidateCreate([]string{"something"})
		require.Error(t, err)
	}
}

func TestDummyConnector_ValidateUpdate(t *testing.T) {
	conn := NewDummy(
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

	require.NoError(t, conn.ValidateUpdate([]string{"something"}, []string{"other"}))

	conn.SetFolderPrefix("Folders")
	require.NoError(t, conn.ValidateUpdate([]string{"Folders", "something"}, []string{"Folders", "other"}))
	require.NoError(t, conn.ValidateUpdate([]string{"Folders", "something"}, []string{"Folders", "other", "long"}))
	require.Error(t, conn.ValidateUpdate([]string{"Folders", "something"}, []string{"other"}))
	require.Error(t, conn.ValidateUpdate([]string{"something"}, []string{"Folders", "other"}))

	conn.SetLabelPrefix("Labels")
	require.NoError(t, conn.ValidateUpdate([]string{"Labels", "something"}, []string{"Labels", "other"}))
	require.NoError(t, conn.ValidateUpdate([]string{"Labels", "something"}, []string{"Labels", "other", "long"}))
	require.Error(t, conn.ValidateUpdate([]string{"Labels", "something"}, []string{"other"}))
	require.Error(t, conn.ValidateUpdate([]string{"something"}, []string{"Labels", "other"}))
	require.Error(t, conn.ValidateUpdate([]string{"Folders", "something"}, []string{"Labels", "something"}))
}

func TestDummyConnector_ValidateDelete(t *testing.T) {
	conn := NewDummy(
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

	require.NoError(t, conn.ValidateDelete([]string{"something"}))

	conn.SetFolderPrefix("Folders")
	require.NoError(t, conn.ValidateDelete([]string{"Folders", "something"}))
	require.Error(t, conn.ValidateDelete([]string{"Folders"}))

	conn.SetLabelPrefix("Labels")
	require.NoError(t, conn.ValidateDelete([]string{"Labels", "something"}))
	require.Error(t, conn.ValidateDelete([]string{"Labels"}))
}
