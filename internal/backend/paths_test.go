package backend

import (
	"fmt"
	"strings"
	"testing"
)

func TestListSuperiorNames(t *testing.T) {
	tests := []struct {
		name, delimiter string

		want []string
	}{
		{
			name:      "this",
			delimiter: "/",
			want:      []string{},
		},
		{
			name:      "this/is/a/test",
			delimiter: "/",
			want:      []string{"this", "this/is", "this/is/a"},
		},
		{
			name:      "/this/is/a/test",
			delimiter: "/",
			want:      []string{"/this", "/this/is", "/this/is/a"},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(fmt.Sprintf("%#v", tt), func(t *testing.T) {
			if res := listSuperiors(tt.name, tt.delimiter); strings.Join(res, "") != strings.Join(tt.want, "") {
				t.Errorf("expected result of %v but got %v", tt.want, res)
			}
		})
	}
}
