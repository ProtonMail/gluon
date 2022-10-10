package state

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListSuperiorNames(t *testing.T) {
	for name, data := range map[string]struct {
		name, delimiter string
		want            []string
	}{
		"no parents": {
			name:      "this",
			delimiter: "/",
			want:      nil,
		},
		"has parents": {
			name:      "this/is/a/test",
			delimiter: "/",
			want:      []string{"this", "this/is", "this/is/a"},
		},
		"wrong delimiter": {
			name:      "this.is.a.test",
			delimiter: "/",
			want:      nil,
		},
		"nil delimiter": {
			name:      "/nil/delimiter.used",
			delimiter: "",
			want:      nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, data.want, listSuperiors(data.name, data.delimiter))
		})
	}
}

func TestListInferiorNames(t *testing.T) {
	for name, data := range map[string]struct {
		parent    string
		delimiter string
		names     []string
		want      []string
	}{
		"no children": {
			parent:    "this",
			delimiter: "/",
			names:     []string{"one", "two", "three", "this", "a/b", "c/d"},
			want:      []string{},
		},
		"has children": {
			parent:    "this",
			delimiter: "/",
			names:     []string{"a/b", "this/one", "this", "this/two", "this/one/two/three", "c/d"},
			want:      []string{"this/two", "this/one/two/three", "this/one"},
		},
		"nil delimiter": {
			parent:    "this",
			delimiter: "",
			names:     []string{"a/b", "this/one", "this", "this/two", "this/one/two/three", "c/d"},
			want:      []string{},
		},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, data.want, listInferiors(data.parent, data.delimiter, data.names))
		})
	}
}
