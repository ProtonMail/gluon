package imap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeqSet(t *testing.T) {
	tests := []struct {
		have []int
		want string
	}{
		{have: []int{}, want: ""},
		{have: []int{1}, want: "1"},
		{have: []int{1, 3}, want: "1,3"},
		{have: []int{1, 3, 5}, want: "1,3,5"},
		{have: []int{1, 2, 3, 5}, want: "1:3,5"},
		{have: []int{1, 2, 3, 5, 6}, want: "1:3,5:6"},
		{have: []int{1, 2, 3, 4, 5, 6}, want: "1:6"},
		{have: []int{1, 3, 4, 5, 6}, want: "1,3:6"},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.want, func(t *testing.T) {
			assert.Equal(t, tc.want, NewSeqSet(tc.have).String())
		})
	}
}
