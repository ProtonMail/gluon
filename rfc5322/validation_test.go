package rfc5322

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateMessageHeaderFields_RequiredFieldsPass(t *testing.T) {
	const literal = `From: Foo@bar.com
Date: Mon, 7 Feb 1994 21:52:25 -0800 (PST)
`

	require.NoError(t, ValidateMessageHeaderFields([]byte(literal)))
}

func TestValidateMessageHeaderFields_ErrOnMissingFrom(t *testing.T) {
	const literal = `Date: Mon, 7 Feb 1994 21:52:25 -0800 (PST)
`

	require.Error(t, ValidateMessageHeaderFields([]byte(literal)))
}

func TestValidateMessageHeaderFields_ErrOnMissingDate(t *testing.T) {
	const literal = `From: Foo@bar.com
`

	require.Error(t, ValidateMessageHeaderFields([]byte(literal)))
}

func TestValidateMessageHeaderFields_ErrOnSingleFromAndSenderEqual(t *testing.T) {
	const literal = `From: Foo@bar.com
Date: Mon, 7 Feb 1994 21:52:25 -0800 (PST)
Sender: Foo@bar.com
`

	require.Error(t, ValidateMessageHeaderFields([]byte(literal)))
}

func TestValidateMessageHeaderFields_AllowSingleFromWithDifferentSender(t *testing.T) {
	const literal = `From: Foo@bar.com
Date: Mon, 7 Feb 1994 21:52:25 -0800 (PST)
Sender: Bar@bar.com
`

	require.NoError(t, ValidateMessageHeaderFields([]byte(literal)))
}

func TestValidateMessageHeaderFields_ErrOnMultipleFromAndNoSender(t *testing.T) {
	const literal = `From: Foo@bar.com, Bar@bar.com
Date: Mon, 7 Feb 1994 21:52:25 -0800 (PST)
`

	require.Error(t, ValidateMessageHeaderFields([]byte(literal)))
}
