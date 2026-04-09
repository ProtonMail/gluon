package command

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/require"
)

func continuationChecker(continued *bool) func(string) error {
	return func(string) error { *continued = true; return nil }
}

func TestParser_Authenticate(t *testing.T) {
	testData := []*Authenticate{
		{UserID: "user1@example.com", Password: "pass"},
		{UserID: "user1@example.com", Password: ""},
		{UserID: "", Password: "pass"},
		{UserID: "", Password: ""},
	}

	for i, data := range testData {
		var continued bool

		tag := fmt.Sprintf("A%04d", i)
		authString := base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "\x00%s\x00%s", data.UserID, data.Password))
		input := toIMAPLine(tag+` AUTHENTICATE PLAIN`, authString)
		s := rfcparser.NewScanner(bytes.NewReader(input))
		p := NewParser(s, WithLiteralContinuationCallback(continuationChecker(&continued)))
		cmd, err := p.Parse()
		message := fmt.Sprintf(" test failed for input %#v", data)

		require.NoError(t, err, "error"+message)
		require.True(t, continued, "continuation"+message)
		require.Equal(t, data, cmd.Payload, "payload"+message)
		require.Equal(t, "authenticate", p.LastParsedCommand(), "command"+message)
		require.Equal(t, tag, p.LastParsedTag(), "tag"+message)
	}
}

func TestParser_AuthenticationWithIdentity(t *testing.T) {
	var continued bool

	authString := base64.StdEncoding.EncodeToString([]byte("identity\x00user\x00pass"))
	s := rfcparser.NewScanner(bytes.NewReader(toIMAPLine(`A0001 authenticate plain`, authString)))
	p := NewParser(s, WithLiteralContinuationCallback(continuationChecker(&continued)))
	cmd, err := p.Parse()

	require.NoError(t, err, "error test failed")
	require.True(t, continued, "continuation test failed")
	require.Equal(t, &Authenticate{UserID: "user", Password: "pass"}, cmd.Payload, "payload test failed")
	require.Equal(t, "authenticate", p.LastParsedCommand(), "command test failed")
	require.Equal(t, "A0001", p.LastParsedTag(), "tag test failed")
}

func TestParser_AuthenticateFailures(t *testing.T) {
	testData := []struct {
		input                []string
		expectedMessage      string
		continuationExpected bool
		description          string
	}{
		{
			input:                []string{`A003 AUTHENTICATE PLAIN`, `*`},
			expectedMessage:      messageClientAbortedAuthentication,
			continuationExpected: true,
			description:          "AUTHENTICATE abortion should return an error",
		},
		{
			input:                []string{`A003 AUTHENTICATE NONE`, `*`},
			expectedMessage:      messageUnsupportedAuthenticationMechanism,
			continuationExpected: false,
			description:          "AUTHENTICATE with unknown mechanism should fail",
		},
		{
			input:                []string{`A003 AUTHENTICATE PLAIN GARBAGE`, `*`},
			expectedMessage:      "expected CR",
			continuationExpected: false,
			description:          "AUTHENTICATE with garbage before CRLF should fail",
		},
		{
			input:                []string{`A003 AUTHENTICATE PLAIN `, `*`},
			expectedMessage:      "expected CR",
			continuationExpected: false,
			description:          "AUTHENTICATE with extra space before CRLF should fail",
		},
		{
			input:                []string{`A003 AUTHENTICATE PLAIN`, `* `},
			expectedMessage:      messageInvalidBase64Content,
			continuationExpected: true,
			description:          "AUTHENTICATE with extra space after the abort `*` should fail",
		},
		{
			input:                []string{`A003 AUTHENTICATE PLAIN`, `* `},
			expectedMessage:      messageInvalidBase64Content,
			continuationExpected: true,
			description:          "AUTHENTICATE with extra space after the abort `*` should fail",
		},
		{
			input:                []string{`A003 AUTHENTICATE PLAIN`, `not-base64`},
			expectedMessage:      messageInvalidBase64Content,
			continuationExpected: true,
			description:          "AUTHENTICATE with invalid base 64 message after continuation should fail",
		},
		{
			input:                []string{`A003 AUTHENTICATE PLAIN`, base64.StdEncoding.EncodeToString([]byte("username+password"))},
			expectedMessage:      messageInvalidAuthenticationData,
			continuationExpected: true,
			description:          "AUTHENTICATE with invalid decoded base64 content should fail",
		},
		{
			input:                []string{`A003 AUTHENTICATE PLAIN`, base64.StdEncoding.EncodeToString([]byte("\x00username\x00password")) + " "},
			expectedMessage:      "expected CR",
			continuationExpected: true,
			description:          "AUTHENTICATE with trailing spaces after a valid base64 message should fail",
		},
	}

	for _, test := range testData {
		var continued bool

		s := rfcparser.NewScanner(bytes.NewReader(toIMAPLine(test.input...)))
		p := NewParser(s, WithLiteralContinuationCallback(continuationChecker(&continued)))
		_, err := p.Parse()
		failureDescription := fmt.Sprintf(" test failed for input %#v", test)

		var parserError *rfcparser.Error

		require.ErrorAs(t, err, &parserError, "error"+failureDescription)
		require.Equal(t, test.expectedMessage, parserError.Message, "error message"+failureDescription)
		require.Equal(t, test.continuationExpected, continued, "continuation"+failureDescription)
	}
}

func TestParser_AuthenticateDisabled(t *testing.T) {
	s := rfcparser.NewScanner(bytes.NewReader(toIMAPLine(`A0001 authenticate plain`)))
	p := NewParser(s, WithDisableIMAPAuthenticate())
	_, err := p.Parse()

	var parserError *rfcparser.Error

	require.ErrorAs(t, err, &parserError)
	require.Equal(t, "unknown command 'authenticate'", parserError.Message)
}
