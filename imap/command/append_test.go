package command

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
)

func buildAppendDateTime(year int, month time.Month, day int, hour int, min int, sec int, zoneHour int, zoneMinutes int, negativeZone bool) time.Time {
	zoneMultiplier := 1
	if negativeZone {
		zoneMultiplier = -1
	}

	zone := (zoneHour*3600 + zoneMinutes*60) * zoneMultiplier

	location := time.FixedZone("zone", zone)

	return time.Date(year, month, day, hour, min, sec, 0, location)
}

func TestParser_AppendCommandWithAllFields(t *testing.T) {
	input := toIMAPLine(`A003 APPEND saved-messages (\Seen) "15-Nov-1984 13:37:01 +0730" {23}`, `My message body is here`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "A003", Payload: &Append{
		Mailbox:  "saved-messages",
		Flags:    []string{`\Seen`},
		Literal:  []byte("My message body is here"),
		DateTime: buildAppendDateTime(1984, time.November, 15, 13, 37, 1, 07, 30, false),
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "append", p.LastParsedCommand())
	require.Equal(t, "A003", p.LastParsedTag())
}

func TestParser_AppendCommandWithLiteralOnly(t *testing.T) {
	input := toIMAPLine(`A003 APPEND saved-messages {23}`, `My message body is here`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "A003", Payload: &Append{
		Mailbox: "saved-messages",
		Literal: []byte("My message body is here"),
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "append", p.LastParsedCommand())
	require.Equal(t, "A003", p.LastParsedTag())
}

func TestParser_AppendCommandWithFlagAndLiteral(t *testing.T) {
	input := toIMAPLine(`A003 APPEND saved-messages (\Seen) {23}`, `My message body is here`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "A003", Payload: &Append{
		Mailbox: "saved-messages",
		Flags:   []string{`\Seen`},
		Literal: []byte("My message body is here"),
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "append", p.LastParsedCommand())
	require.Equal(t, "A003", p.LastParsedTag())
}

func TestParser_AppendCommandWithDateTimeAndLiteral(t *testing.T) {
	input := toIMAPLine(`A003 APPEND saved-messages "15-Nov-1984 13:37:01 +0730" {23}`, `My message body is here`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "A003", Payload: &Append{
		Mailbox:  "saved-messages",
		Literal:  []byte("My message body is here"),
		DateTime: buildAppendDateTime(1984, time.November, 15, 13, 37, 1, 07, 30, false),
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "append", p.LastParsedCommand())
	require.Equal(t, "A003", p.LastParsedTag())
}

func TestParser_AppendWithUTF8Literal(t *testing.T) {
	const literal = `割ゃちとた紀別チノホコ隠面ノネシ披畑つ筋型菊ラ済百チユネ報能げほべえ一王ユサ禁未シムカ学康ほル退今ずはぞゃ宿理古えべにあ。民さぱをだ意宇りう医6通海ヘクヲ点71丈2社つぴげわ中知多ッ場親られ見要クラ著喜栄潟ぼゆラけ。著災ンう三育府能に汁裁ラヤユ哉能ルサレ開30被ちゃ英死オイ教禁能ふてっせ社化選市へす。`
	input := toIMAPLine(fmt.Sprintf("A003 APPEND saved-messages (\\Seen) {%v}", len(literal)), literal)

	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	expected := Command{Tag: "A003", Payload: &Append{
		Mailbox: "saved-messages",
		Flags:   []string{`\Seen`},
		Literal: []byte(literal),
	}}

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}
