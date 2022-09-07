package imap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeqSet(t *testing.T) {
	tests := []struct {
		have []SeqID
		want string
	}{
		{have: []SeqID{}, want: ""},
		{have: []SeqID{1}, want: "1"},
		{have: []SeqID{1, 3}, want: "1,3"},
		{have: []SeqID{1, 3, 5}, want: "1,3,5"},
		{have: []SeqID{1, 2, 3, 5}, want: "1:3,5"},
		{have: []SeqID{1, 2, 3, 5, 6}, want: "1:3,5:6"},
		{have: []SeqID{1, 2, 3, 4, 5, 6}, want: "1:6"},
		{have: []SeqID{1, 3, 4, 5, 6}, want: "1,3:6"},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.want, func(t *testing.T) {
			assert.Equal(t, tc.want, NewSeqSet(tc.have).String())
		})
	}
}
