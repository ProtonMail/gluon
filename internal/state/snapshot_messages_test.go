package state

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/stretchr/testify/require"
)

func TestMessages(t *testing.T) {
	msg := newMsgList()

	msg.insert(ids.NewMessageIDPairWithoutRemote("msgID1"), 10, imap.NewFlagSet(imap.FlagSeen))
	msg.insert(ids.NewMessageIDPairWithoutRemote("msgID2"), 20, imap.NewFlagSet(imap.FlagSeen))
	msg.insert(ids.NewMessageIDPairWithoutRemote("msgID3"), 30, imap.NewFlagSet(imap.FlagSeen))
	msg.insert(ids.NewMessageIDPairWithoutRemote("msgID4"), 40, imap.NewFlagSet(imap.FlagSeen))
	msg.insert(ids.NewMessageIDPairWithoutRemote("msgID5"), 50, imap.NewFlagSet(imap.FlagSeen))
	msg.insert(ids.NewMessageIDPairWithoutRemote("msgID6"), 60, imap.NewFlagSet(imap.FlagSeen))

	msg.remove("msgID2")
	msg.remove("msgID4")
	msg.remove("msgID6")

	{
		require.Equal(t, 3, msg.len())
	}

	{
		require.True(t, msg.has("msgID1"))
		require.True(t, msg.has("msgID3"))
		require.True(t, msg.has("msgID5"))

		require.False(t, msg.has("msgID2"))
		require.False(t, msg.has("msgID4"))
		require.False(t, msg.has("msgID6"))
	}

	{
		msg1, ok := msg.get("msgID1")
		require.True(t, ok)
		require.Equal(t, imap.SeqID(1), msg1.Seq)
		require.Equal(t, imap.UID(10), msg1.UID)

		msg2, ok := msg.get("msgID2")
		require.False(t, ok)
		require.Nil(t, msg2)

		msg3, ok := msg.get("msgID3")
		require.True(t, ok)
		require.Equal(t, imap.SeqID(2), msg3.Seq)
		require.Equal(t, imap.UID(30), msg3.UID)

		msg4, ok := msg.get("msgID4")
		require.False(t, ok)
		require.Nil(t, msg4)

		msg5, ok := msg.get("msgID5")
		require.True(t, ok)
		require.Equal(t, imap.SeqID(3), msg5.Seq)
		require.Equal(t, imap.UID(50), msg5.UID)
	}

	{
		require.Equal(t, must(msg.get("msgID1")), must(msg.seq(1)))
		require.Equal(t, must(msg.get("msgID3")), must(msg.seq(2)))
		require.Equal(t, must(msg.get("msgID5")), must(msg.seq(3)))
	}
}

func must[T any](val T, ok bool) T {
	if !ok {
		panic("not ok")
	}

	return val
}
