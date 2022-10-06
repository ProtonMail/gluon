package parser

import (
	"github.com/stretchr/testify/require"
	"net/mail"
	"testing"
)

func TestRFC5322AddressListSuccess(T *testing.T) {
	input := `Alice <alice@example.com>, Bob <bob@example.com>, Eve <eve@example.com>`
	expected := []*mail.Address{
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
	}
	addrList, err := ParseRFC5322AddressList(input)
	require.NoError(T, err)
	require.Equal(T, expected, addrList)
}

func TestRFC5322Failure(T *testing.T) {
	input := `"comma, name"  <username@server.com>, another, name <address@server.com>`
	_, err := ParseRFC5322AddressList(input)
	require.Error(T, err)
}

func BenchmarkParseEmailGo(B *testing.B) {
	input := `Alice <alice@example.com>, Bob <bob@example.com>, Eve <eve@example.com>`
	for i := 0; i < B.N; i++ {
		_, err := mail.ParseAddressList(input)
		require.NoError(B, err)
	}
}

func BenchmarkParseEmailCPP(B *testing.B) {
	input := `Alice <alice@example.com>, Bob <bob@example.com>, Eve <eve@example.com>`
	for i := 0; i < B.N; i++ {
		_, err := ParseRFC5322AddressList(input)
		require.NoError(B, err)
	}
}
