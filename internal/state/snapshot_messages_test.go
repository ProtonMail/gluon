package state

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/stretchr/testify/require"
)

// nolint:govet
func TestMessages(t *testing.T) {
	msg := newMsgList(8)

	msg.insert(messageIDPair(1, "1"), 10, imap.NewFlagSet(imap.FlagSeen))
	msg.insert(messageIDPair(2, "2"), 20, imap.NewFlagSet(imap.FlagSeen))
	msg.insert(messageIDPair(3, "3"), 30, imap.NewFlagSet(imap.FlagSeen))
	msg.insert(messageIDPair(4, "4"), 40, imap.NewFlagSet(imap.FlagSeen))
	msg.insert(messageIDPair(5, "5"), 50, imap.NewFlagSet(imap.FlagSeen))
	msg.insert(messageIDPair(6, "6"), 60, imap.NewFlagSet(imap.FlagSeen))

	msg.remove(2)
	msg.remove(4)
	msg.remove(6)

	{
		require.Equal(t, 3, msg.len())
	}

	{
		require.True(t, msg.has(1))
		require.True(t, msg.has(3))
		require.True(t, msg.has(5))

		require.False(t, msg.has(2))
		require.False(t, msg.has(4))
		require.False(t, msg.has(6))
	}

	{
		msg1, ok := msg.get(1)
		require.True(t, ok)
		require.Equal(t, imap.SeqID(1), msg1.Seq)
		require.Equal(t, imap.UID(10), msg1.UID)

		_, ok = msg.get(2)
		require.False(t, ok)

		msg3, ok := msg.get(3)
		require.True(t, ok)
		require.Equal(t, imap.SeqID(2), msg3.Seq)
		require.Equal(t, imap.UID(30), msg3.UID)

		_, ok = msg.get(4)
		require.False(t, ok)

		msg5, ok := msg.get(5)
		require.True(t, ok)
		require.Equal(t, imap.SeqID(3), msg5.Seq)
		require.Equal(t, imap.UID(50), msg5.UID)
	}

	{
		require.Equal(t, must(msg.get(1)), must(msg.seq(1)))
		require.Equal(t, must(msg.get(3)), must(msg.seq(2)))
		require.Equal(t, must(msg.get(5)), must(msg.seq(3)))
	}
}

// nolint:govet
func TestMessageUIDRange(t *testing.T) {
	msg := newMsgList(8)

	msg.insert(messageIDPair(1, "1"), 10, imap.NewFlagSet(imap.FlagSeen))
	msg.insert(messageIDPair(2, "2"), 20, imap.NewFlagSet(imap.FlagSeen))
	msg.insert(messageIDPair(3, "3"), 30, imap.NewFlagSet(imap.FlagSeen))
	msg.insert(messageIDPair(4, "4"), 40, imap.NewFlagSet(imap.FlagSeen))
	msg.insert(messageIDPair(5, "5"), 50, imap.NewFlagSet(imap.FlagSeen))
	msg.insert(messageIDPair(6, "6"), 60, imap.NewFlagSet(imap.FlagSeen))

	// UIDRange Higher than maximum
	{
		result := msg.uidRange(imap.UID(40), imap.UID(80))
		require.Equal(t, 3, len(result))
		require.Equal(t, result[0].UID, imap.UID(40))
		require.Equal(t, result[1].UID, imap.UID(50))
		require.Equal(t, result[2].UID, imap.UID(60))
	}

	// UIDRange Lower than minimum
	{
		result := msg.uidRange(imap.UID(1), imap.UID(10))
		require.Equal(t, 1, len(result))
		require.Equal(t, result[0].UID, imap.UID(10))
	}

	// UIDRange lower than all values
	{
		result := msg.uidRange(imap.UID(1), imap.UID(5))
		require.Empty(t, result)
	}

	// UIDRange higher than all values
	{
		result := msg.uidRange(imap.UID(100), imap.UID(120))
		require.Empty(t, result)
	}

	// UIDRange higher that doesn't exist in between
	{
		result := msg.uidRange(imap.UID(21), imap.UID(29))
		require.Empty(t, result)
	}

	// UIDRange for interval that is valid, but not all values exist
	{
		result := msg.uidRange(imap.UID(25), imap.UID(42))
		require.Equal(t, 2, len(result))
		require.Equal(t, result[0].UID, imap.UID(30))
		require.Equal(t, result[1].UID, imap.UID(40))
	}
}

func messageIDPair(internalID imap.InternalMessageID, remoteID imap.MessageID) ids.MessageIDPair {
	return ids.MessageIDPair{InternalID: internalID, RemoteID: remoteID}
}

func must[T any](val T, ok bool) T {
	if !ok {
		panic("not ok")
	}

	return val
}
