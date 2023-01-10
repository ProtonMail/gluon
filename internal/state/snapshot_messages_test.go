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

	id1 := imap.NewInternalMessageID()
	id2 := imap.NewInternalMessageID()
	id3 := imap.NewInternalMessageID()
	id4 := imap.NewInternalMessageID()
	id5 := imap.NewInternalMessageID()
	id6 := imap.NewInternalMessageID()

	require.NoError(t, msg.insert(messageIDPair(id1, "1"), 10, imap.NewFlagSet(imap.FlagSeen)))
	require.NoError(t, msg.insert(messageIDPair(id2, "2"), 20, imap.NewFlagSet(imap.FlagSeen)))
	require.NoError(t, msg.insert(messageIDPair(id3, "3"), 30, imap.NewFlagSet(imap.FlagSeen)))
	require.NoError(t, msg.insert(messageIDPair(id4, "4"), 40, imap.NewFlagSet(imap.FlagSeen)))
	require.NoError(t, msg.insert(messageIDPair(id5, "5"), 50, imap.NewFlagSet(imap.FlagSeen)))
	require.NoError(t, msg.insert(messageIDPair(id6, "6"), 60, imap.NewFlagSet(imap.FlagSeen)))

	msg.remove(id2)
	msg.remove(id4)
	msg.remove(id6)

	{
		require.Equal(t, 3, msg.len())
	}

	{
		require.True(t, msg.has(id1))
		require.True(t, msg.has(id3))
		require.True(t, msg.has(id5))

		require.False(t, msg.has(id2))
		require.False(t, msg.has(id4))
		require.False(t, msg.has(id6))
	}

	{
		msg1, ok := msg.get(id1)
		require.True(t, ok)
		require.Equal(t, imap.SeqID(1), msg1.Seq)
		require.Equal(t, imap.UID(10), msg1.UID)

		_, ok = msg.get(id2)
		require.False(t, ok)

		msg3, ok := msg.get(id3)
		require.True(t, ok)
		require.Equal(t, imap.SeqID(2), msg3.Seq)
		require.Equal(t, imap.UID(30), msg3.UID)

		_, ok = msg.get(id4)
		require.False(t, ok)

		msg5, ok := msg.get(id5)
		require.True(t, ok)
		require.Equal(t, imap.SeqID(3), msg5.Seq)
		require.Equal(t, imap.UID(50), msg5.UID)
	}

	{
		require.Equal(t, must(msg.get(id1)), must(msg.seq(1)))
		require.Equal(t, must(msg.get(id3)), must(msg.seq(2)))
		require.Equal(t, must(msg.get(id5)), must(msg.seq(3)))
	}
}

// nolint:govet
func TestMessageUIDRange(t *testing.T) {
	msg := newMsgList(8)

	id1 := imap.NewInternalMessageID()
	id2 := imap.NewInternalMessageID()
	id3 := imap.NewInternalMessageID()
	id4 := imap.NewInternalMessageID()
	id5 := imap.NewInternalMessageID()
	id6 := imap.NewInternalMessageID()

	require.NoError(t, msg.insert(messageIDPair(id1, "1"), 10, imap.NewFlagSet(imap.FlagSeen)))
	require.NoError(t, msg.insert(messageIDPair(id2, "2"), 20, imap.NewFlagSet(imap.FlagSeen)))
	require.NoError(t, msg.insert(messageIDPair(id3, "3"), 30, imap.NewFlagSet(imap.FlagSeen)))
	require.NoError(t, msg.insert(messageIDPair(id4, "4"), 40, imap.NewFlagSet(imap.FlagSeen)))
	require.NoError(t, msg.insert(messageIDPair(id5, "5"), 50, imap.NewFlagSet(imap.FlagSeen)))
	require.NoError(t, msg.insert(messageIDPair(id6, "6"), 60, imap.NewFlagSet(imap.FlagSeen)))

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

func TestMessageRange1HigherThanMax(t *testing.T) {
	msg := newMsgList(8)

	id1 := imap.NewInternalMessageID()
	id2 := imap.NewInternalMessageID()

	require.NoError(t, msg.insert(messageIDPair(id1, "1"), 1, imap.NewFlagSet(imap.FlagSeen)))
	require.NoError(t, msg.insert(messageIDPair(id2, "2"), 2, imap.NewFlagSet(imap.FlagSeen)))

	seqSetInterval := [][]string{{"3", "*"}}

	{
		uidInterval, err := msg.resolveUIDInterval(seqSetInterval)
		require.NoError(t, err)
		seqInterval, err := msg.resolveSeqInterval(seqSetInterval)
		require.NoError(t, err)

		require.Equal(t, uidInterval, []UIDInterval{{begin: imap.UID(3), end: imap.UID(3)}})
		require.Equal(t, seqInterval, []SeqInterval{{begin: imap.SeqID(3), end: imap.SeqID(3)}})
	}

	{
		_, err := msg.getMessagesInSeqRange(seqSetInterval)
		require.Error(t, err)
	}
	{
		messages, err := msg.getMessagesInUIDRange(seqSetInterval)
		require.NoError(t, err)
		require.Empty(t, messages)
	}
}

func TestSnapListGetMessages(t *testing.T) {
	msg := newMsgList(8)

	id1 := imap.NewInternalMessageID()
	id2 := imap.NewInternalMessageID()

	require.NoError(t, msg.insert(messageIDPair(id1, "1"), 1, imap.NewFlagSet(imap.FlagSeen)))
	require.NoError(t, msg.insert(messageIDPair(id2, "2"), 2, imap.NewFlagSet(imap.FlagSeen)))

	{
		seqSetInterval := [][]string{{"3", "*"}}
		_, err := msg.getMessagesInSeqRange(seqSetInterval)
		require.Error(t, err)

		uidInterval, err := msg.getMessagesInUIDRange(seqSetInterval)
		require.NoError(t, err)
		require.Empty(t, uidInterval)
	}
	{
		seqSetInterval := [][]string{{"*", "3"}}
		_, err := msg.getMessagesInSeqRange(seqSetInterval)
		require.Error(t, err)

		uidInterval, err := msg.getMessagesInUIDRange(seqSetInterval)
		require.NoError(t, err)
		require.Empty(t, uidInterval)
	}
	{
		seqSetInterval := [][]string{{"1", "*"}}
		seqList, err := msg.getMessagesInSeqRange(seqSetInterval)
		require.NoError(t, err)
		require.Equal(t, 2, len(seqList))
		require.Equal(t, seqList[0].Seq, imap.SeqID(1))
		require.Equal(t, seqList[1].Seq, imap.SeqID(2))

		uidList, err := msg.getMessagesInUIDRange(seqSetInterval)
		require.NoError(t, err)
		require.Equal(t, 2, len(uidList))
		require.Equal(t, uidList[0].UID, imap.UID(1))
		require.Equal(t, uidList[1].UID, imap.UID(2))
	}
	{
		seqSetInterval := [][]string{{"2", "3"}}
		_, err := msg.getMessagesInSeqRange(seqSetInterval)
		require.Error(t, err)

		uidList, err := msg.getMessagesInUIDRange(seqSetInterval)
		require.NoError(t, err)
		require.Equal(t, 1, len(uidList))
		require.Equal(t, uidList[0].UID, imap.UID(2))
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
