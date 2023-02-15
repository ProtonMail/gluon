package command

import (
	"bytes"
	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/htmlindex"
	"testing"
	"time"
	"unicode/utf8"
)

func buildSearchTestDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestParser_SearchCommandAll(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyAll{},
		},
	}}

	input := toIMAPLine(`tag SEARCH ALL`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "search", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}

func TestParser_SearchCommandWithCharset(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "UTF-8",
		Keys: []SearchKey{
			&SearchKeyAll{},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH CHARSET UTF-8 ALL`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandWithCharsetInWrongLocation(t *testing.T) {
	_, err := testParseCommand(`tag SEARCH ALL CHARSET UTF-8`)
	require.Error(t, err)
}

func TestParser_SearchCommandAnswered(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyAnswered{},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH ANSWERED`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandBCC(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyBCC{Value: "foobar"},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH BCC foobar`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandBefore(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyBefore{Value: buildSearchTestDate(2009, time.January, 01)},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH BEFORE 01-Jan-2009`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandBody(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyBody{Value: "foobar"},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH BODY foobar`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandCC(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyCC{Value: "foobar"},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH CC foobar`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandDeleted(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyDeleted{},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH DELETED`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandFlagged(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyFlagged{},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH FLAGGED`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandFrom(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyFrom{Value: "foobar"},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH From foobar`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandKeyword(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyKeyword{Value: "foobar"},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH KEYWORD foobar`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandNew(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyNew{},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH NEW`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandOld(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyOld{},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH OLD`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandRecent(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyRecent{},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH RECENT`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandOn(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyOn{Value: buildSearchTestDate(2009, time.January, 01)},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH ON 01-Jan-2009`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandSince(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeySince{Value: buildSearchTestDate(2009, time.January, 01)},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH SINCE 01-Jan-2009`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandSubject(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeySubject{Value: "foobar"},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH SUBJECT foobar`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandText(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyText{Value: "foobar"},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH TEXT foobar`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandTo(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyTo{Value: "foobar"},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH TO foobar`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandUnanswered(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyUnanswered{},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH UNANSWERED`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandUndeleted(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyUndeleted{},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH UNDELETED`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandUnflagged(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyUnflagged{},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH UNFLAGGED`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandUnseen(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyUnseen{},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH UNSEEN`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandUnkeyword(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyUnkeyword{Value: "foobar"},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH UNKEYWORD foobar`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandDraft(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyDraft{},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH DRAFT`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchCommandHeader(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyHeader{Field: "field", Value: "foobar"},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH HEADER field foobar`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchLarger(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyLarger{Value: 1024},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH LARGER 1024`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchNot(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyNot{
				Key: &SearchKeyLarger{Value: 1024},
			},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH NOT LARGER 1024`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchOr(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyOr{
				Key1: &SearchKeyLarger{Value: 1024},
				Key2: &SearchKeySmaller{Value: 4096},
			},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH OR LARGER 1024 SMALLER 4096`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchSentBefore(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeySentBefore{Value: buildSearchTestDate(2009, time.January, 1)},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH SENTBEFORE 01-Jan-2009`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchSentOn(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeySentOn{Value: buildSearchTestDate(2009, time.January, 1)},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH SENTON 01-Jan-2009`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchSentSince(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeySentSince{Value: buildSearchTestDate(2009, time.January, 1)},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH SENTSINCE 01-Jan-2009`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchSmaller(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeySmaller{Value: 512},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH SMALLER 512`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchUID(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyUID{Value: 512},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH UID 512`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchUndraft(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyUndraft{},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH UNDRAFT`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchMultipleKeys(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "",
		Keys: []SearchKey{
			&SearchKeyUndraft{},
			&SearchKeySubject{Value: "foo"},
			&SearchKeySentSince{Value: buildSearchTestDate(2009, time.January, 1)},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH UNDRAFT SUBJECT foo SENTSINCE 01-Jan-2009`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_SearchUtf8String(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "UTF-8",
		Keys: []SearchKey{
			&SearchKeySubject{Value: "割ゃちとた紀別チノホコ隠面ノ"},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH CHARSET UTF-8 SUBJECT "割ゃちとた紀別チノホコ隠面ノ"`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_Search_ISO_8859_1_String(t *testing.T) {
	// Encode "ééé" as ISO-8859-1.
	text := enc("ééé", "ISO-8859-1")
	textWithQuotes := enc(`"ééé"`, "ISO-8859-1")

	// Assert that text is no longer valid UTF-8.
	require.False(t, utf8.Valid(text))
	require.False(t, utf8.Valid(textWithQuotes))

	expected := Command{Tag: "tag", Payload: &Search{
		Charset: "ISO-8859-1",
		Keys: []SearchKey{
			&SearchKeySubject{Value: string(text)},
		},
	}}

	cmd, err := testParseCommand(`tag SEARCH CHARSET ISO-8859-1 SUBJECT ` + string(textWithQuotes))
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func enc(text, encoding string) []byte {
	enc, err := htmlindex.Get(encoding)
	if err != nil {
		panic(err)
	}

	b, err := enc.NewEncoder().Bytes([]byte(text))
	if err != nil {
		panic(err)
	}

	return b
}
