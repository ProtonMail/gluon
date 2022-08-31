package state

import (
	"fmt"
	"testing"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		ref, pattern, delimiter, name string

		want string
	}{
		{
			ref:       "",
			pattern:   "",
			delimiter: "/",
			want:      "",
		},
		{
			ref:       "#news.comp.mail.misc",
			pattern:   "",
			delimiter: ".",
			want:      "#news.",
		},
		{
			ref:       "/usr/staff/jones",
			pattern:   "",
			delimiter: "/",
			want:      "/",
		},
		{
			ref:       "some.",
			pattern:   "",
			delimiter: ".",
			want:      "some.",
		},
		{
			ref:       "some",
			pattern:   "",
			delimiter: ".",
			want:      "",
		},
		{
			ref:       "",
			pattern:   "*",
			delimiter: "/",
			name:      "INBOX",
			want:      "INBOX",
		},
		{
			ref:       "",
			pattern:   "%",
			delimiter: "/",
			name:      "INBOX",
			want:      "INBOX",
		},
		{
			ref:       "~/Mail/",
			pattern:   "%",
			delimiter: "/",
			name:      "~/Mail/meetings",
			want:      "~/Mail/meetings",
		},
		{
			ref:       "~/Mail/",
			pattern:   "%",
			delimiter: "/",
			name:      "~/Mail/foo/bar",
			want:      "~/Mail/foo",
		},
		{
			ref:       "some.",
			pattern:   "thing",
			delimiter: ".",
			name:      "some.thing",
			want:      "some.thing",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(fmt.Sprintf("%#v", tt), func(t *testing.T) {
			res, match := match(tt.ref, tt.pattern, tt.delimiter, tt.name)
			if !match {
				t.Errorf("expected match to be true but it wasn't")
			}
			if res != tt.want {
				t.Errorf("expected match of %v but got %v", tt.want, res)
			}
		})
	}
}
