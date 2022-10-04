package response

import (
	"testing"

	"github.com/ProtonMail/gluon/imap"
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
		"interrupted exists": {
			given: []Response{
				Exists().WithCount(1),
				Exists().WithCount(2),
				Expunge(1),
				Exists().WithCount(1),
				Exists().WithCount(2),
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
		"interrupted recent": {
			given: []Response{
				Recent().WithCount(1),
				Recent().WithCount(2),
				Expunge(1),
				Recent().WithCount(1),
				Recent().WithCount(2),
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
		"interrupting exists and recent": {
			given: []Response{
				Exists().WithCount(1),
				Recent().WithCount(1),
				Exists().WithCount(2),
				Recent().WithCount(2),
				Expunge(1),
				Exists().WithCount(2),
				Recent().WithCount(2),
			},
			want: []Response{
				Exists().WithCount(2),
				Recent().WithCount(2),
				Expunge(1),
				Exists().WithCount(2),
				Recent().WithCount(2),
			},
		},
	} {
		t.Run(testCase, func(t *testing.T) {
			require.Equal(t, testData.want, Merge(testData.given))
		})
	}
}

func TestMergeFetchResponses(t *testing.T) {
	noFlags := ItemFlags(imap.NewFlagSet())
	seen := ItemFlags(imap.NewFlagSet(imap.FlagSeen))
	seenDeleted := ItemFlags(imap.NewFlagSet(imap.FlagSeen, imap.FlagDeleted))

	for testCase, testData := range map[string]struct {
		given, want []Response
	}{
		"different fetch ids": {
			given: []Response{
				Fetch(1).WithItems(seen),
				Fetch(2).WithItems(seen),
				Fetch(3).WithItems(seen),
				Fetch(4).WithItems(seen),
			},
			want: []Response{
				Fetch(1).WithItems(seen),
				Fetch(2).WithItems(seen),
				Fetch(3).WithItems(seen),
				Fetch(4).WithItems(seen),
			},
		},
		"same fetch ids": {
			given: []Response{
				Fetch(1).WithItems(seen),
				Fetch(1).WithItems(noFlags),
				Fetch(1).WithItems(seenDeleted),
			},
			want: []Response{
				Fetch(1).WithItems(seenDeleted),
			},
		},
		"alter fetch ids": {
			given: []Response{
				Fetch(1).WithItems(seen),
				Fetch(2).WithItems(seen),
				Fetch(2).WithItems(seenDeleted),
				Fetch(1).WithItems(noFlags),
				Fetch(3).WithItems(seen),
				Fetch(3).WithItems(noFlags),
				Fetch(3).WithItems(seen),
			},
			want: []Response{
				Fetch(1).WithItems(noFlags),
				Fetch(2).WithItems(seenDeleted),
				Fetch(3).WithItems(seen),
			},
		},
		"interrupt fetch merge": {
			given: []Response{
				Fetch(1).WithItems(seen),
				Expunge(2),
				Fetch(1).WithItems(noFlags),
				Fetch(1).WithItems(seenDeleted),
			},
			want: []Response{
				Fetch(1).WithItems(seen),
				Expunge(2),
				Fetch(1).WithItems(seenDeleted),
			},
		},
		"don't interrupt fetch merge": {
			given: []Response{
				Fetch(1).WithItems(seen),
				Exists().WithCount(2),
				Fetch(1).WithItems(noFlags),
				Fetch(1).WithItems(seenDeleted),
			},
			want: []Response{
				Fetch(1).WithItems(seenDeleted),
				Exists().WithCount(2),
			},
		},
		"combination of all": {
			given: []Response{
				Fetch(1).WithItems(seen),
				Exists().WithCount(10),
				Fetch(1).WithItems(noFlags),
				Expunge(2),
				Exists().WithCount(9),
				Recent().WithCount(2),
				Fetch(1).WithItems(seenDeleted),
				Fetch(3).WithItems(seenDeleted),
				Exists().WithCount(10),
				Recent().WithCount(3),
				Fetch(3).WithItems(noFlags),
				Fetch(3).WithItems(seen),
			},
			want: []Response{
				Fetch(1).WithItems(noFlags),
				Exists().WithCount(10),
				Expunge(2),
				Exists().WithCount(10),
				Recent().WithCount(3),
				Fetch(1).WithItems(seenDeleted),
				Fetch(3).WithItems(seen),
			},
		},
	} {
		t.Run(testCase, func(t *testing.T) {
			require.Equal(t, testData.want, Merge(testData.given))
		})
	}
}

func TestMergeResponsesPanics(t *testing.T) {
	for testCase, testData := range map[string]struct {
		given     []Response
		wantPanic string
	}{
		"exists decreased": {
			given: []Response{
				Exists().WithCount(1),
				Exists().WithCount(2),
				Exists().WithCount(1),
				Exists().WithCount(2),
				Exists().WithCount(3),
				Exists().WithCount(4),
			},
			wantPanic: "consecutive exists must be non-decreasing, but had 2 and new 1",
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
			wantPanic: "consecutive recents must be non-decreasing, but had 2 and new 1",
		},
		"decreasing exists while having recent": {
			given: []Response{
				Exists().WithCount(1),
				Recent().WithCount(1),
				Exists().WithCount(2),
				Recent().WithCount(2),
				Exists().WithCount(1),
			},
			wantPanic: "consecutive exists must be non-decreasing, but had 2 and new 1",
		},
		"decreasing recent while having exists": {
			given: []Response{
				Recent().WithCount(1),
				Exists().WithCount(1),
				Recent().WithCount(2),
				Exists().WithCount(2),
				Recent().WithCount(1),
			},
			wantPanic: "consecutive recents must be non-decreasing, but had 2 and new 1",
		},
	} {
		t.Run(testCase, func(t *testing.T) {
			require.PanicsWithValue(t, testData.wantPanic, func() { Merge(testData.given) })
		})
	}
}
