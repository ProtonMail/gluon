package response

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMergeResponses(t *testing.T) {
	for testCase, testData := range map[string]struct {
		given, want []Response
	}{
		"nil": {},
		"zero length": {
			given: []Response{},
			want:  []Response{},
		},
		"consecutive exists": {
			given: []Response{
				Exists().WithCount(1),
				Exists().WithCount(2),
				Exists().WithCount(3),
				Exists().WithCount(4),
			},
			want: []Response{
				Exists().WithCount(4),
			},
		},
		"exists decreased": {
			given: []Response{
				Exists().WithCount(1),
				Exists().WithCount(2),
				Exists().WithCount(1),
				Exists().WithCount(2),
				Exists().WithCount(3),
				Exists().WithCount(4),
			},
			want: []Response{
				Exists().WithCount(2),
				Exists().WithCount(1),
				Exists().WithCount(4),
			},
		},
		"interupted exists": {
			given: []Response{
				Exists().WithCount(1),
				Exists().WithCount(2),
				Expunge(1),
				Exists().WithCount(3),
			},
			want: []Response{
				Exists().WithCount(2),
				Expunge(1),
				Exists().WithCount(3),
			},
		},
		"consecutive recent": {
			given: []Response{
				Recent().WithCount(1),
				Recent().WithCount(2),
				Recent().WithCount(3),
				Recent().WithCount(4),
			},
			want: []Response{
				Recent().WithCount(4),
			},
		},
		"recent decreased": {
			given: []Response{
				Recent().WithCount(1),
				Recent().WithCount(2),
				Recent().WithCount(1),
				Recent().WithCount(2),
				Recent().WithCount(3),
				Recent().WithCount(4),
			},
			want: []Response{
				Recent().WithCount(2),
				Recent().WithCount(1),
				Recent().WithCount(4),
			},
		},
		"interupted recent": {
			given: []Response{
				Recent().WithCount(1),
				Recent().WithCount(2),
				Expunge(1),
				Recent().WithCount(3),
			},
			want: []Response{
				Recent().WithCount(2),
				Expunge(1),
				Recent().WithCount(3),
			},
		},
		"combining exists and recent": {
			given: []Response{
				Exists().WithCount(1),
				Recent().WithCount(1),
				Exists().WithCount(2),
				Recent().WithCount(2),
				Exists().WithCount(3),
				Exists().WithCount(4),
				Recent().WithCount(3),
				Recent().WithCount(4),
				Recent().WithCount(5),
				Recent().WithCount(6),
				Recent().WithCount(7),
				Recent().WithCount(8),
				Exists().WithCount(5),
				Exists().WithCount(6),
			},
			want: []Response{
				Exists().WithCount(6),
				Recent().WithCount(8),
			},
		},
		"decreasing exists while having recent": {
			given: []Response{
				Exists().WithCount(1),
				Recent().WithCount(1),
				Exists().WithCount(2),
				Recent().WithCount(2),
				Exists().WithCount(1),
			},
			want: []Response{
				Exists().WithCount(2),
				Exists().WithCount(1),
				Recent().WithCount(2),
			},
		},
		"decreasing recent while having exists": {
			given: []Response{
				Recent().WithCount(1),
				Exists().WithCount(1),
				Recent().WithCount(2),
				Exists().WithCount(2),
				Recent().WithCount(1),
			},
			want: []Response{
				Recent().WithCount(2),
				Recent().WithCount(1),
				Exists().WithCount(2),
			},
		},
		"interupting exists and recent": {
			given: []Response{
				Exists().WithCount(1),
				Recent().WithCount(1),
				Exists().WithCount(2),
				Recent().WithCount(2),
				Expunge(1),
				Exists().WithCount(3),
				Recent().WithCount(3),
			},
			want: []Response{
				Exists().WithCount(2),
				Recent().WithCount(2),
				Expunge(1),
				Exists().WithCount(3),
				Recent().WithCount(3),
			},
		},
	} {
		t.Run(testCase, func(t *testing.T) {
			require.Equal(t, testData.want, Merge(testData.given))
		})
	}
}
