package command

import (
	"bytes"
	"github.com/ProtonMail/gluon/imap/parser"
	cppParser "github.com/ProtonMail/gluon/internal/parser"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_FetchCommandAll(t *testing.T) {
	input := toIMAPLine(`tag FETCH 1 ALL`)
	s := parser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeAll{},
		},
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "fetch", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}

func TestParser_FetchCommandFull(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeFull{},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 Full`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandFast(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeFast{},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 Fast`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandEnvelope(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeEnvelope{},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 ENVELOPE`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandFlags(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeFlags{},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 FLAGS`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandInternalDate(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeInternalDate{},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 INTERNALDATE`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandRFC822Header(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeRFC822Header{},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 RFC822.HEADER`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandRFC822Size(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeRFC822Size{},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 RFC822.SIZE`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandRFC822Text(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeRFC822Text{},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 RFC822.TEXT`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandBodyStructure(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeBodyStructure{},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 BODYSTRUCTURE`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandBody(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeBody{},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 BODY`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandUID(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeUID{},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 UID`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandBodySection_Empty(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeBodySection{
				Section: nil,
				Peek:    false,
				Partial: nil,
			},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 BODY[]`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandBodySection_Header(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeBodySection{
				Section: &BodySectionHeader{},
				Peek:    false,
				Partial: nil,
			},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 BODY[HEADER]`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandBodySection_Text(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeBodySection{
				Section: &BodySectionText{},
				Peek:    false,
				Partial: nil,
			},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 BODY[TEXT]`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandBodySection_HeaderFieldsSingular(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeBodySection{
				Section: &BodySectionHeaderFields{
					Negate: false,
					Fields: []string{"FROM"},
				},
				Peek:    false,
				Partial: nil,
			},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 BODY[HEADER.FIELDS (FROM)]`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandBodySection_HeaderFieldsMultiple(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeBodySection{
				Section: &BodySectionHeaderFields{
					Negate: false,
					Fields: []string{"FROM", "TO", "SUBJECT"},
				},
				Peek:    false,
				Partial: nil,
			},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 BODY[HEADER.FIELDS (FROM TO SUBJECT)]`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandBodySection_HeaderFieldsNot(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeBodySection{
				Section: &BodySectionHeaderFields{
					Negate: true,
					Fields: []string{"FROM", "TO", "SUBJECT"},
				},
				Peek:    false,
				Partial: nil,
			},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 BODY[HEADER.FIELDS.NOT (FROM TO SUBJECT)]`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandBodySection_MIMEIsErrorWithoutPart(t *testing.T) {
	_, err := testParseCommand(`tag FETCH 1 BODY[MIME]`)
	require.Error(t, err)
}

func TestParser_FetchCommandBodySection_MIME(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeBodySection{
				Section: &BodySectionPart{
					Part:    []int{4, 2, 1},
					Section: &BodySectionMIME{},
				},
				Peek:    false,
				Partial: nil,
			},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 BODY[4.2.1.MIME]`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandBodySection_PartWithSectionMsgText(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeBodySection{
				Section: &BodySectionPart{
					Part:    []int{4, 2, 1},
					Section: &BodySectionHeader{},
				},
				Peek:    false,
				Partial: nil,
			},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 BODY[4.2.1.HEADER]`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandBodySection_Partial(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeBodySection{
				Section: &BodySectionText{},
				Peek:    false,
				Partial: &BodySectionPartial{
					Offset: 100,
					Count:  50,
				},
			},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 BODY[TEXT]<100.50>`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandBodySection_Peek(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 1, End: 1}},
		Attributes: []FetchAttribute{
			&FetchAttributeBodySection{
				Section: &BodySectionText{},
				Peek:    true,
				Partial: nil,
			},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 1 BODY.PEEK[TEXT]`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_FetchCommandMultiple(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &FetchCommand{
		SeqSet: []SeqRange{{Begin: 2, End: 4}},
		Attributes: []FetchAttribute{
			&FetchAttributeFlags{},
			&FetchAttributeInternalDate{},
			&FetchAttributeRFC822Size{},
			&FetchAttributeEnvelope{},
			&FetchAttributeBodySection{
				Section: &BodySectionPart{
					Part:    []int{1, 3},
					Section: &BodySectionText{},
				},
				Peek: true,
				Partial: &BodySectionPartial{
					Offset: 50,
					Count:  100,
				},
			},
		},
	}}

	cmd, err := testParseCommand(`tag FETCH 2:4 (FLAGS INTERNALDATE RFC822.SIZE ENVELOPE BODY.PEEK[1.3.TEXT]<50.100>)`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func BenchmarkParser_Fetch(b *testing.B) {
	input := toIMAPLine(`tag FETCH 2:4 (FLAGS INTERNALDATE RFC822.SIZE ENVELOPE BODY.PEEK[1.3.TEXT]<50.100>)`)

	for i := 0; i < b.N; i++ {
		s := parser.NewScanner(bytes.NewReader(input))
		p := NewParser(s)

		_, err := p.Parse()
		require.NoError(b, err)
	}
}

func BenchmarkParser_FetchCPP(b *testing.B) {
	input := string(toIMAPLine(`tag FETCH 2:4 (FLAGS INTERNALDATE RFC822.SIZE ENVELOPE BODY.PEEK[1.3.TEXT]<50.100>)`))

	literalMap := cppParser.NewStringMap()

	for i := 0; i < b.N; i++ {
		cppParser.Parse(input, literalMap, "/")
	}
}
