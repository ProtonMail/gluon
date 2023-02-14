package command

import (
	"bytes"
	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_UIDCommandCopy(t *testing.T) {
	input := toIMAPLine(`tag UID COPY 1:* INBOX`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{
		Tag: "tag",
		Payload: &UIDCommand{
			Command: &CopyCommand{
				Mailbox: "INBOX",
				SeqSet:  []SeqRange{{Begin: 1, End: SeqNumValueAsterisk}},
			},
		},
	}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "uid", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}

func TestParser_UIDCommandMove(t *testing.T) {
	expected := Command{
		Tag: "tag",
		Payload: &UIDCommand{
			Command: &MoveCommand{
				Mailbox: "INBOX",
				SeqSet:  []SeqRange{{Begin: 1, End: SeqNumValueAsterisk}},
			},
		},
	}

	cmd, err := testParseCommand(`tag UID MOVE 1:* INBOX`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_UIDCommandStore(t *testing.T) {
	expected := Command{
		Tag: "tag",
		Payload: &UIDCommand{
			Command: &StoreCommand{
				SeqSet: []SeqRange{{
					Begin: 1,
					End:   1,
				}},
				Action: StoreActionAddFlags,
				Flags:  []string{"Foo", "Bar"},
				Silent: true,
			},
		},
	}

	cmd, err := testParseCommand(`tag UID STORE 1 +FLAGS.SILENT (Foo Bar)`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_UIDCommandExpunge(t *testing.T) {
	expected := Command{
		Tag: "tag",
		Payload: &UIDExpungeCommand{
			SeqSet: []SeqRange{{Begin: 1, End: SeqNumValueAsterisk}},
		},
	}

	cmd, err := testParseCommand(`tag UID EXPUNGE 1:*`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_UIDCommandFetch(t *testing.T) {
	expected := Command{
		Tag: "tag",
		Payload: &UIDCommand{
			Command: &FetchCommand{
				SeqSet: []SeqRange{{Begin: 1, End: 1}},
				Attributes: []FetchAttribute{
					&FetchAttributeFast{},
				},
			},
		},
	}

	cmd, err := testParseCommand(`tag UID FETCH 1 FAST`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_UIDCommandSearch(t *testing.T) {
	expected := Command{
		Tag: "tag",
		Payload: &UIDCommand{
			Command: &SearchCommand{
				Charset: "",
				Keys: []SearchKey{
					&SearchKeyAnswered{},
				},
			},
		},
	}

	cmd, err := testParseCommand(`tag UID SEARCH ANSWERED`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_UIDCommandInvalid(t *testing.T) {
	_, err := testParseCommand(`tag UID LIST`)
	require.Error(t, err)
}
