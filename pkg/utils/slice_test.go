package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type filterTestCase[T any] struct {
	name     string
	input    []T
	keep     func(T) bool
	expected []T
}

func TestSlice_Filter_Int(t *testing.T) {
	testCases := []filterTestCase[int]{
		{
			name:  "empty",
			input: []int{},
			keep: func(_ int) bool {
				return true
			},
			expected: []int{},
		},
		{
			name:  "all",
			input: []int{1, 2, 3},
			keep: func(_ int) bool {
				return true
			},
			expected: []int{1, 2, 3},
		},
		{
			name:  "none",
			input: []int{1, 2, 3},
			keep: func(_ int) bool {
				return false
			},
			expected: []int{},
		},
		{
			name:  "only one",
			input: []int{1, 2, 3},
			keep: func(i int) bool {
				return i == 2
			},
			expected: []int{2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Filter(tc.input, tc.keep)

			require.Equal(t, tc.expected, result)
			require.Len(t, result, len(tc.expected))
		})
	}
}

func TestSlice_Filter_String(t *testing.T) {
	testCases := []filterTestCase[string]{
		{
			name:  "empty",
			input: []string{},
			keep: func(_ string) bool {
				return true
			},
			expected: []string{},
		},
		{
			name:  "all",
			input: []string{"a", "b", "c"},
			keep: func(_ string) bool {
				return true
			},
			expected: []string{"a", "b", "c"},
		},
		{
			name:  "none",
			input: []string{"a", "b", "c"},
			keep: func(_ string) bool {
				return false
			},
			expected: []string{},
		},
		{
			name:  "only one",
			input: []string{"a", "b", "c"},
			keep: func(s string) bool {
				return s == "b"
			},
			expected: []string{"b"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Filter(tc.input, tc.keep)

			require.Equal(t, tc.expected, result)
			require.Len(t, result, len(tc.expected))
		})
	}
}

func TestSlice_Filter_Struct(t *testing.T) {
	type testStruct struct {
		ID   int
		Name string
	}

	newRandomTestStruct := func(id int) testStruct {
		return testStruct{
			ID:   id,
			Name: fmt.Sprintf("test-%d", id),
		}
	}

	testCases := []filterTestCase[testStruct]{
		{
			name:  "empty",
			input: []testStruct{},
			keep: func(_ testStruct) bool {
				return true
			},
			expected: []testStruct{},
		},
		{
			name:  "all",
			input: []testStruct{newRandomTestStruct(1), newRandomTestStruct(2), newRandomTestStruct(3)},
			keep: func(_ testStruct) bool {
				return true
			},
			expected: []testStruct{newRandomTestStruct(1), newRandomTestStruct(2), newRandomTestStruct(3)},
		},
		{
			name:  "none",
			input: []testStruct{newRandomTestStruct(1), newRandomTestStruct(2), newRandomTestStruct(3)},
			keep: func(_ testStruct) bool {
				return false
			},
			expected: []testStruct{},
		},
		{
			name:  "specific id",
			input: []testStruct{newRandomTestStruct(1), newRandomTestStruct(2), newRandomTestStruct(3)},
			keep: func(ts testStruct) bool {
				return ts.ID == 2
			},
			expected: []testStruct{newRandomTestStruct(2)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Filter(tc.input, tc.keep)

			require.Equal(t, tc.expected, result)
			require.Len(t, result, len(tc.expected))
		})
	}
}
