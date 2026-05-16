package command

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// These tests cover parsing of the X-GM-EXT-1 Gmail label extension across the
// STORE, FETCH and SEARCH commands. The extension is what Paperless-NGX (and
// other Gmail-style clients) use to tag mail over IMAP.

// --- STORE +/-/= X-GM-LABELS ----------------------------------------------

func TestParser_StoreCommandAddGmailLabels(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Store{
		SeqSet:   []SeqRange{{Begin: 1, End: 1}},
		Action:   StoreActionAddFlags,
		Flags:    []string{"Label1", "Label2"},
		Silent:   false,
		DataItem: StoreDataItemGmailLabels,
	}}

	cmd, err := testParseCommand(`tag STORE 1 +X-GM-LABELS ("Label1" "Label2")`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_StoreCommandRemoveGmailLabels(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Store{
		SeqSet:   []SeqRange{{Begin: 1, End: 1}},
		Action:   StoreActionRemFlags,
		Flags:    []string{"Label1"},
		Silent:   false,
		DataItem: StoreDataItemGmailLabels,
	}}

	cmd, err := testParseCommand(`tag STORE 1 -X-GM-LABELS ("Label1")`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

// STORE X-GM-LABELS without a +/- prefix parses as StoreActionSetFlags. Note
// that the session handler currently treats anything other than AddFlags as a
// removal (see handleStoreGmailLabels) — this test pins the parse only.
func TestParser_StoreCommandSetGmailLabels(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Store{
		SeqSet:   []SeqRange{{Begin: 1, End: 1}},
		Action:   StoreActionSetFlags,
		Flags:    []string{"Label1"},
		Silent:   false,
		DataItem: StoreDataItemGmailLabels,
	}}

	cmd, err := testParseCommand(`tag STORE 1 X-GM-LABELS ("Label1")`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

// Python's imaplib sends the label without wrapping it in parentheses. This is
// the exact form Paperless-NGX emits, so it must parse (regression for the
// "accept bare labels" fix).
func TestParser_StoreCommandAddGmailLabelsBare(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Store{
		SeqSet:   []SeqRange{{Begin: 1, End: 1}},
		Action:   StoreActionAddFlags,
		Flags:    []string{"Paperless"},
		Silent:   false,
		DataItem: StoreDataItemGmailLabels,
	}}

	cmd, err := testParseCommand(`tag STORE 1 +X-GM-LABELS Paperless`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_StoreCommandGmailLabelsWithSpaces(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Store{
		SeqSet:   []SeqRange{{Begin: 1, End: 1}},
		Action:   StoreActionAddFlags,
		Flags:    []string{"Label With Spaces", "Other"},
		Silent:   false,
		DataItem: StoreDataItemGmailLabels,
	}}

	cmd, err := testParseCommand(`tag STORE 1 +X-GM-LABELS ("Label With Spaces" "Other")`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_StoreCommandGmailLabelsSystemLabel(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Store{
		SeqSet:   []SeqRange{{Begin: 1, End: 1}},
		Action:   StoreActionAddFlags,
		Flags:    []string{`\Inbox`},
		Silent:   false,
		DataItem: StoreDataItemGmailLabels,
	}}

	cmd, err := testParseCommand(`tag STORE 1 +X-GM-LABELS (\Inbox)`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_StoreCommandGmailLabelsSilent(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Store{
		SeqSet:   []SeqRange{{Begin: 1, End: 1}},
		Action:   StoreActionAddFlags,
		Flags:    []string{"Label1"},
		Silent:   true,
		DataItem: StoreDataItemGmailLabels,
	}}

	cmd, err := testParseCommand(`tag STORE 1 +X-GM-LABELS.SILENT ("Label1")`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

// --- FETCH X-GM-LABELS ------------------------------------------------------

func TestParser_FetchCommandGmailLabels(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Fetch{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeGmailLabels{},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 X-GM-LABELS`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandGmailLabelsWithFlags(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Fetch{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeFlags{},
			&FetchAttributeGmailLabels{},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 (FLAGS X-GM-LABELS)`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandGmailLabelsLowercase(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Fetch{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeGmailLabels{},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 (x-gm-labels)`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

// --- SEARCH X-GM-LABELS -----------------------------------------------------

func TestParser_SearchCommandGmailLabels(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyGmailLabels{Value: "Paperless"},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH X-GM-LABELS "Paperless"`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandGmailLabelsAtom(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyGmailLabels{Value: "Paperless"},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH X-GM-LABELS Paperless`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

// NOT X-GM-LABELS "<tag>" is exactly how Paperless-NGX excludes already-tagged
// mail from its search criteria, so the negated form must parse.
func TestParser_SearchCommandNotGmailLabels(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyNot{Key: &SearchKeyGmailLabels{Value: "Paperless"}},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH NOT X-GM-LABELS "Paperless"`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}
