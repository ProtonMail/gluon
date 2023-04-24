package rfc5322

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"net/mail"
	"testing"

	"github.com/ProtonMail/gluon/rfcparser"
	"github.com/stretchr/testify/assert"
)

func newTestRFCParser(s string) *rfcparser.Parser {
	p := rfcparser.NewParser(rfcparser.NewScanner(bytes.NewReader([]byte(s))))
	if p.Advance() != nil {
		panic("failed to advance parser")
	}

	return p
}

func TestParseAddress(t *testing.T) {
	tests := []struct {
		input string
		addrs []*mail.Address
	}{
		{
			input: `user@example.com`,
			addrs: []*mail.Address{{
				Address: `user@example.com`,
			}},
		},
		{
			input: `John Doe <jdoe@machine.example>`,
			addrs: []*mail.Address{{
				Name:    `John Doe`,
				Address: `jdoe@machine.example`,
			}},
		},
		{
			input: `Mary Smith <mary@example.net>`,
			addrs: []*mail.Address{{
				Name:    `Mary Smith`,
				Address: `mary@example.net`,
			}},
		},
		{
			input: `"Joe Q. Public" <john.q.public@example.com>`,
			addrs: []*mail.Address{{
				Name:    `Joe Q. Public`,
				Address: `john.q.public@example.com`,
			}},
		},
		{
			input: `Mary Smith <mary@x.test>`,
			addrs: []*mail.Address{{
				Name:    `Mary Smith`,
				Address: `mary@x.test`,
			}},
		},
		{
			input: `jdoe@example.org`,
			addrs: []*mail.Address{{
				Address: `jdoe@example.org`,
			}},
		},
		{
			input: `Who? <one@y.test>`,
			addrs: []*mail.Address{{
				Name:    `Who?`,
				Address: `one@y.test`,
			}},
		},
		{
			input: `<boss@nil.test>`,
			addrs: []*mail.Address{{
				Address: `boss@nil.test`,
			}},
		},
		{
			input: `"Giant; \"Big\" Box" <sysservices@example.net>`,
			addrs: []*mail.Address{{
				Name:    `Giant; "Big" Box`,
				Address: `sysservices@example.net`,
			}},
		},
		{
			input: `Pete <pete@silly.example>`,
			addrs: []*mail.Address{{
				Name:    `Pete`,
				Address: `pete@silly.example`,
			}},
		},
		{
			input: `"Mary Smith: Personal Account" <smith@home.example>`,
			addrs: []*mail.Address{{
				Name:    `Mary Smith: Personal Account`,
				Address: `smith@home.example`,
			}},
		},
		{
			input: `Pete(A nice \) chap) <pete(his account)@silly.test(his host)>`,
			addrs: []*mail.Address{{
				Name:    `Pete`,
				Address: `pete@silly.test`,
			}},
		},
		{
			input: `Gogh Fir <gf@example.com>`,
			addrs: []*mail.Address{{
				Name:    `Gogh Fir`,
				Address: `gf@example.com`,
			}},
		},
		{
			input: `normal name  <username@server.com>`,
			addrs: []*mail.Address{{
				Name:    `normal name`,
				Address: `username@server.com`,
			}},
		},
		{
			input: `"comma, name"  <username@server.com>`,
			addrs: []*mail.Address{{
				Name:    `comma, name`,
				Address: `username@server.com`,
			}},
		},
		{
			input: `name  <username@server.com> (ignore comment)`,
			addrs: []*mail.Address{{
				Name:    `name`,
				Address: `username@server.com`,
			}},
		},
		{
			input: `"Mail Robot" <>`,
			addrs: []*mail.Address{{
				Name: `Mail Robot`,
			}},
		},
		{
			input: `Michal Hořejšek <hořejšek@mail.com>`,
			addrs: []*mail.Address{{
				Name:    `Michal Hořejšek`,
				Address: `hořejšek@mail.com`, // Not his real address.
			}},
		},
		{
			input: `First Last <user@domain.com >`,
			addrs: []*mail.Address{{
				Name:    `First Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First Last <user@domain.com. >`,
			addrs: []*mail.Address{{
				Name:    `First Last`,
				Address: `user@domain.com.`,
			}},
		},
		{
			input: `First Last <user@domain.com.>`,
			addrs: []*mail.Address{{
				Name:    `First Last`,
				Address: `user@domain.com.`,
			}},
		},
		{
			input: `First Last <user@domain.com:25>`,
			addrs: []*mail.Address{{
				Name:    `First Last`,
				Address: `user@domain.com:25`,
			}},
		},
		{
			input: `First Last <user@[10.0.0.1]>`,
			addrs: []*mail.Address{{
				Name:    `First Last`,
				Address: `user@[10.0.0.1]`,
			}},
		},
		{
			input: `<postmaster@[10.10.10.10]>`,
			addrs: []*mail.Address{{
				Address: `postmaster@[10.10.10.10]`,
			}},
		},
		{
			input: `First Last < user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `user@domain.com,`,
			addrs: []*mail.Address{{
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First Middle "Last" <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First Middle Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First Middle Last <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First Middle Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First Middle"Last" <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First Middle Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First Middle "Last"<user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First Middle Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First "Middle" "Last" <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First Middle Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First "Middle""Last" <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First Middle Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `first.last <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `first.last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `first . last <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `first.last`,
				Address: `user@domain.com`,
			}},
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.input, func(t *testing.T) {
			addrs, err := ParseAddress(test.input)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.addrs, addrs)
		})
	}
}

func TestParseAddressList(t *testing.T) {
	tests := []struct {
		input string
		addrs []*mail.Address
	}{
		{
			input: `Alice <alice@example.com>, Bob <bob@example.com>, Eve <eve@example.com>`,
			addrs: []*mail.Address{
				{
					Name:    `Alice`,
					Address: `alice@example.com`,
				},
				{
					Name:    `Bob`,
					Address: `bob@example.com`,
				},
				{
					Name:    `Eve`,
					Address: `eve@example.com`,
				},
			},
		},
		{
			input: `Alice <alice@example.com>; Bob <bob@example.com>; Eve <eve@example.com>`,
			addrs: []*mail.Address{
				{
					Name:    `Alice`,
					Address: `alice@example.com`,
				},
				{
					Name:    `Bob`,
					Address: `bob@example.com`,
				},
				{
					Name:    `Eve`,
					Address: `eve@example.com`,
				},
			},
		},
		{
			input: `Ed Jones <c@a.test>,joe@where.test,John <jdoe@one.test>`,
			addrs: []*mail.Address{
				{
					Name:    `Ed Jones`,
					Address: `c@a.test`,
				},
				{
					Address: `joe@where.test`,
				},
				{
					Name:    `John`,
					Address: `jdoe@one.test`,
				},
			},
		},
		{
			input: `name (ignore comment)  <username@server.com>,  (Comment as name) username2@server.com`,
			addrs: []*mail.Address{
				{
					Name:    `name`,
					Address: `username@server.com`,
				},
				{
					Address: `username2@server.com`,
				},
			},
		},
		{
			input: `"normal name"  <username@server.com>, "comma, name" <address@server.com>`,
			addrs: []*mail.Address{
				{
					Name:    `normal name`,
					Address: `username@server.com`,
				},
				{
					Name:    `comma, name`,
					Address: `address@server.com`,
				},
			},
		},
		{
			input: `"comma, one"  <username@server.com>, "comma, two" <address@server.com>`,
			addrs: []*mail.Address{
				{
					Name:    `comma, one`,
					Address: `username@server.com`,
				},
				{
					Name:    `comma, two`,
					Address: `address@server.com`,
				},
			},
		},
		{
			input: `normal name  <username@server.com>, (comment)All.(around)address@(the)server.com`,
			addrs: []*mail.Address{
				{
					Name:    `normal name`,
					Address: `username@server.com`,
				},
				{
					Address: `All.address@server.com`,
				},
			},
		},
		{
			input: `normal name  <username@server.com>, All.("comma, in comment")address@(the)server.com`,
			addrs: []*mail.Address{
				{
					Name:    `normal name`,
					Address: `username@server.com`,
				},
				{
					Address: `All.address@server.com`,
				},
			},
		},
		{
			input: `Alice <alice@example.com>, Group:foo@bar;, bar@bar`,
			addrs: []*mail.Address{
				{
					Name:    `Alice`,
					Address: `alice@example.com`,
				},
				{
					Name:    ``,
					Address: `foo@bar`,
				},
				{
					Name:    ``,
					Address: `bar@bar`,
				},
			},
		},
		{
			input: `user@domain <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `user@domain`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `user @ domain <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `user@domain`,
				Address: `user@domain.com`,
			}},
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.input, func(t *testing.T) {
			addrs, err := ParseAddressList(test.input)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.addrs, addrs)
		})
	}
}

func TestParseGroup(t *testing.T) {
	tests := []struct {
		input string
		addrs []*mail.Address
	}{
		{
			input: `A Group:Ed Jones <c@a.test>,joe@where.test,John <jdoe@one.test>;`,
			addrs: []*mail.Address{
				{
					Name:    `Ed Jones`,
					Address: `c@a.test`,
				},
				{
					Address: `joe@where.test`,
				},
				{
					Name:    `John`,
					Address: `jdoe@one.test`,
				},
			},
		},
		{
			input: `undisclosed recipients:;`,
			addrs: []*mail.Address{},
		},
		{
			// We permit the group to not end in a semicolon, although as per RFC5322 it really should.
			input: `undisclosed recipients:`,
			addrs: []*mail.Address{},
		},
		{
			// We permit the group to be surrounded with quotes, although as per RFC5322 it really shouldn't be.
			input: `"undisclosed recipients:"`,
			addrs: []*mail.Address{},
		},
		{
			// We permit the group to be surrounded with quotes, although as per RFC5322 it really shouldn't be.
			input: `"undisclosed recipients:;"`,
			addrs: []*mail.Address{},
		},
		{
			input: `undisclosed recipients:, foo@bar`,
			addrs: []*mail.Address{
				{
					Address: `foo@bar`,
				},
			},
		},
		{
			input: `undisclosed recipients:;, foo@bar`,
			addrs: []*mail.Address{
				{
					Address: `foo@bar`,
				},
			},
		},
		{
			input: `undisclosed recipients:bar@bar;, foo@bar`,
			addrs: []*mail.Address{
				{
					Address: `bar@bar`,
				},
				{
					Address: `foo@bar`,
				},
			},
		},
		{
			input: `"undisclosed recipients:", foo@bar`,
			addrs: []*mail.Address{
				{
					Address: `foo@bar`,
				},
			},
		},
		{
			input: `(Empty list)(start)Hidden recipients  :(nobody(that I know))  ;`,
			addrs: []*mail.Address{},
		},
		{
			input: `foo@bar, g:bar@bar; z@z`,
			addrs: []*mail.Address{
				{
					Address: `foo@bar`,
				},
				{
					Address: `bar@bar`,
				},
				{
					Address: `z@z`,
				},
			},
		},
		{
			input: `foo@bar, g:bar@bar;; z@z`,
			addrs: []*mail.Address{
				{
					Address: `foo@bar`,
				},
				{
					Address: `bar@bar`,
				},
				{
					Address: `z@z`,
				},
			},
		},
		{
			input: `foo@bar, g:bar@bar;, z@z`,
			addrs: []*mail.Address{
				{
					Address: `foo@bar`,
				},
				{
					Address: `bar@bar`,
				},
				{
					Address: `z@z`,
				},
			},
		},
		{
			input: `foo@bar, g:; z@z`,
			addrs: []*mail.Address{
				{
					Address: `foo@bar`,
				},
				{
					Address: `z@z`,
				},
			},
		},
		{
			input: `foo@bar, g:;; z@z`,
			addrs: []*mail.Address{
				{
					Address: `foo@bar`,
				},
				{
					Address: `z@z`,
				},
			},
		},
		{
			input: `foo@bar, g:;, z@z`,
			addrs: []*mail.Address{
				{
					Address: `foo@bar`,
				},
				{
					Address: `z@z`,
				},
			},
		},
		{
			input: `foo@bar, "g:;", z@z`,
			addrs: []*mail.Address{
				{
					Address: `foo@bar`,
				},
				{
					Address: `z@z`,
				},
			},
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.input, func(t *testing.T) {
			addrs, err := ParseAddressList(test.input)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.addrs, addrs)
		})
	}
}

func TestParseSingleAddressEncodedWord(t *testing.T) {
	tests := []struct {
		input string
		addrs []*mail.Address
	}{
		{
			input: `=?US-ASCII?Q?Keith_Moore?= <moore@cs.utk.edu>`,
			addrs: []*mail.Address{{
				Name:    `Keith Moore`,
				Address: `moore@cs.utk.edu`,
			}},
		},
		{
			input: `=?ISO-8859-1?Q?Keld_J=F8rn_Simonsen?= <keld@dkuug.dk>`,
			addrs: []*mail.Address{{
				Name:    `Keld Jørn Simonsen`,
				Address: `keld@dkuug.dk`,
			}},
		},
		{
			input: `=?ISO-8859-1?Q?Andr=E9?= Pirard <PIRARD@vm1.ulg.ac.be>`,
			addrs: []*mail.Address{{
				Name:    `André Pirard`,
				Address: `PIRARD@vm1.ulg.ac.be`,
			}},
		},
		{
			input: `=?ISO-8859-1?Q?Olle_J=E4rnefors?= <ojarnef@admin.kth.se>`,
			addrs: []*mail.Address{{
				Name:    `Olle Järnefors`,
				Address: `ojarnef@admin.kth.se`,
			}},
		},
		{
			input: `=?ISO-8859-1?Q?Patrik_F=E4ltstr=F6m?= <paf@nada.kth.se>`,
			addrs: []*mail.Address{{
				Name:    `Patrik Fältström`,
				Address: `paf@nada.kth.se`,
			}},
		},
		{
			input: `Nathaniel Borenstein <nsb@thumper.bellcore.com> (=?iso-8859-8?b?7eXs+SDv4SDp7Oj08A==?=)`,
			addrs: []*mail.Address{{
				Name:    `Nathaniel Borenstein`,
				Address: `nsb@thumper.bellcore.com`,
			}},
		},
		{
			input: `=?UTF-8?B?PEJlemUgam3DqW5hPg==?= <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `<Beze jména>`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First Middle =?utf-8?Q?Last?= <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First Middle Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First Middle=?utf-8?Q?Last?= <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First Middle=?utf-8?Q?Last?=`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First Middle =?utf-8?Q?Last?=<user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First Middle Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First =?utf-8?Q?Middle?= =?utf-8?Q?Last?= <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First MiddleLast`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First =?utf-8?Q?Middle?==?utf-8?Q?Last?= <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First MiddleLast`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First "Middle"=?utf-8?Q?Last?= <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First Middle Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First "Middle" =?utf-8?Q?Last?= <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First Middle Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `First "Middle" =?utf-8?Q?Last?=<user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `First Middle Last`,
				Address: `user@domain.com`,
			}},
		},
		{
			input: `=?UTF-8?B?PEJlemUgam3DqW5hPg==?= <user@domain.com>`,
			addrs: []*mail.Address{{
				Name:    `<Beze jména>`,
				Address: `user@domain.com`,
			}},
		},
	}
	for _, test := range tests {
		test := test

		t.Run(test.input, func(t *testing.T) {
			addrs, err := ParseAddressList(test.input)
			assert.NoError(t, err)
			assert.ElementsMatch(t, test.addrs, addrs)
		})
	}
}

func TestParseAddressInvalid(t *testing.T) {
	inputs := []string{
		`user@domain...com`,
		`"comma, name"  <username@server.com>, another, name <address@server.com>`,
		`username`,
		`=?ISO-8859-2?Q?First_Last?= <user@domain.com>, <user@domain.com,First/AAA/BBB/CCC,>`,
		`=?windows-1250?Q?Spr=E1vce_syst=E9mu?=`,
		`"'user@domain.com.'"`,
		`<this is not an email address>`,
		`"Mail Delivery System <>" <@>`,
	}

	for _, test := range inputs {
		test := test

		t.Run(test, func(t *testing.T) {
			_, err := ParseAddressList(test)
			assert.Error(t, err)
			assert.True(t, rfcparser.IsError(err))
		})
	}
}

func TestParseDisplayNameOnlyShouldBeError(t *testing.T) {
	const input = "FooBar"
	_, err := ParseAddressList(input)
	require.Error(t, err)
}

func TestParseInvalidHeaderValueShouldBeError(t *testing.T) {
	// E.g: Incorrect header format causes headers fields to be combined into one, this should be invalid.
	const input = "FooBar From:Too Subjects:x"
	_, err := ParseAddressList(input)
	require.Error(t, err)
}

func TestParseEmptyStringIsNotError(t *testing.T) {
	_, err := ParseAddressList("")
	require.NoError(t, err)

	_, err = ParseAddress("")
	require.NoError(t, err)
}

func TestParseAddressListEmoji(t *testing.T) {
	input := `=?utf-8?q?Goce_Test_=F0=9F=A4=A6=F0=9F=8F=BB=E2=99=82=F0=9F=99=88?= =?utf-8?q?=F0=9F=8C=B2=E2=98=98=F0=9F=8C=B4?= <foo@bar.com>, "Proton GMX Edit" <z@bar.com>, "beta@bar.com" <beta@bar.com>, "testios12" <random@bar.com>, "random@bar.com" <random@bar.com>, =?utf-8?q?=C3=9C=C3=A4=C3=B6_Jakdij?= <another@bar.com>, =?utf-8?q?Q=C3=A4_T=C3=B6=C3=BCst_12_Edit?= <random2@bar.com>, =?utf-8?q?=E2=98=98=EF=B8=8F=F0=9F=8C=B2=F0=9F=8C=B4=F0=9F=99=82=E2=98=BA?= =?utf-8?q?=EF=B8=8F=F0=9F=98=83?= <dust@bar.com>, "Somebody Outlook" <hotmal@bar.com>`
	expected := []*mail.Address{
		{
			Name:    "Goce Test 🤦🏻♂🙈🌲☘🌴",
			Address: "foo@bar.com",
		},
		{
			Name:    "Proton GMX Edit",
			Address: "z@bar.com",
		},
		{
			Name:    "beta@bar.com",
			Address: "beta@bar.com",
		},
		{
			Name:    "testios12",
			Address: "random@bar.com",
		},
		{
			Name:    "random@bar.com",
			Address: "random@bar.com",
		},
		{
			Name:    "Üäö Jakdij",
			Address: "another@bar.com",
		},
		{
			Name:    "Qä Töüst 12 Edit",
			Address: "random2@bar.com",
		},
		{
			Name:    "☘️🌲🌴🙂☺️😃",
			Address: "dust@bar.com",
		},
		{
			Name:    "Somebody Outlook",
			Address: "hotmal@bar.com",
		},
	}

	addrs, err := ParseAddressList(input)
	assert.NoError(t, err)
	assert.ElementsMatch(t, expected, addrs)
}

func TestParserAddressEmailValidation(t *testing.T) {
	inputs := []string{
		"test@io",
		"test@iana.org",
		"test@nominet.org.uk",
		"test@about.museum",
		"a@iana.org",
		"test.test@iana.org",
		"!#$%&`*+/=?^`{|}~@iana.org",
		"123@iana.org",
		"test@123.com",
		"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghiklm@iana.org",
		"test@mason-dixon.com",
		"test@c--n.com",
		"test@xn--hxajbheg2az3al.xn--jxalpdlp",
		"xn--test@iana.org",
		"1@pm.me",
	}

	for _, test := range inputs {
		test := test

		t.Run(test, func(t *testing.T) {
			_, err := ParseAddressList(test)
			assert.NoError(t, err)
		})
	}
}

func TestParse_GODT_2587_infinite_loop(t *testing.T) {
	_, err := ParseAddressList("00@[000000000000000")
	assert.Error(t, err)
}
