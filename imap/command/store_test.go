package command

import (
	"bytes"
	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_StoreCommandSetFlags(t *testing.T) {
	input := toIMAPLine(`tag STORE 1 FLAGS Foo`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &StoreCommand{
		SeqSet: []SeqRange{{
			Begin: 1,
			End:   1,
		}},
		Action: StoreActionSetFlags,
		Flags:  []string{"Foo"},
		Silent: false,
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "store", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}

func TestParser_StoreCommandAddFlags(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &StoreCommand{
		SeqSet: []SeqRange{{
			Begin: 1,
			End:   1,
		}},
		Action: StoreActionAddFlags,
		Flags:  []string{"Foo"},
		Silent: false,
	}}

	cmd, err := testParseCommand(`tag STORE 1 +FLAGS Foo`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_StoreCommandRemoveFlags(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &StoreCommand{
		SeqSet: []SeqRange{{
			Begin: 1,
			End:   1,
		}},
		Action: StoreActionRemFlags,
		Flags:  []string{"Foo"},
		Silent: false,
	}}

	cmd, err := testParseCommand(`tag STORE 1 -FLAGS Foo`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_StoreCommandSilent(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &StoreCommand{
		SeqSet: []SeqRange{{
			Begin: 1,
			End:   1,
		}},
		Action: StoreActionAddFlags,
		Flags:  []string{"Foo"},
		Silent: true,
	}}

	cmd, err := testParseCommand(`tag STORE 1 +FLAGS.SILENT Foo`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_StoreCommandMultipleFlags(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &StoreCommand{
		SeqSet: []SeqRange{{
			Begin: 1,
			End:   1,
		}},
		Action: StoreActionAddFlags,
		Flags:  []string{"Foo", "Bar"},
		Silent: true,
	}}

	cmd, err := testParseCommand(`tag STORE 1 +FLAGS.SILENT Foo Bar`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_StoreCommandMultipleFlagsWithParen(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &StoreCommand{
		SeqSet: []SeqRange{{
			Begin: 1,
			End:   1,
		}},
		Action: StoreActionAddFlags,
		Flags:  []string{"Foo", "Bar"},
		Silent: true,
	}}

	cmd, err := testParseCommand(`tag STORE 1 +FLAGS.SILENT (Foo Bar)`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}
