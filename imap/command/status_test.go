package command

import (
	"bytes"
	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParser_StatusCommandRecent(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Status{
		Mailbox:    "INBOX",
		Attributes: []StatusAttribute{StatusAttributeRecent},
	}}

	input := toIMAPLine(`tag STATUS INBOX (RECENT)`)
	s := rfcparser.NewScanner(bytes.NewReader(input))
	p := NewParser(s)

	cmd, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
	require.Equal(t, "status", p.LastParsedCommand())
	require.Equal(t, "tag", p.LastParsedTag())
}

func TestParser_StatusCommandMessages(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Status{
		Mailbox:    "Foo",
		Attributes: []StatusAttribute{StatusAttributeRecent},
	}}

	cmd, err := testParseCommand(`tag STATUS Foo (RECENT)`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_StatusCommandUIDNext(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Status{
		Mailbox:    "Foo",
		Attributes: []StatusAttribute{StatusAttributeUIDNext},
	}}

	cmd, err := testParseCommand(`tag STATUS Foo (UIDNEXT)`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_StatusCommandUIDValidity(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Status{
		Mailbox:    "Foo",
		Attributes: []StatusAttribute{StatusAttributeUIDValidity},
	}}

	cmd, err := testParseCommand(`tag STATUS Foo (UIDVALIDITY)`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_StatusCommandUnseen(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Status{
		Mailbox:    "Foo",
		Attributes: []StatusAttribute{StatusAttributeUnseen},
	}}

	cmd, err := testParseCommand(`tag STATUS Foo (UNSEEN)`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}

func TestParser_StatusCommandMultiple(t *testing.T) {
	expected := Command{Tag: "tag", Payload: &Status{
		Mailbox:    "Foo",
		Attributes: []StatusAttribute{StatusAttributeUnseen, StatusAttributeRecent},
	}}

	cmd, err := testParseCommand(`tag STATUS Foo (UNSEEN RECENT)`)
	require.NoError(t, err)
	require.Equal(t, expected, cmd)
}
