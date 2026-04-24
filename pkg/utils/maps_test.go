package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type keysTestCase[K comparable, V any] struct {
	name     string
	expected []K
	builder  func() map[K]V
}

type valuesTestCase[K comparable, V any] struct {
	name     string
	expected []V
	builder  func() map[K]V
}

func TestMaps_Keys_String(t *testing.T) {
	testCases := []keysTestCase[string, int]{
		{
			name: "empty",
			builder: func() map[string]int {
				return make(map[string]int, 0)
			},
			expected: []string{},
		},
		{
			name: "valid",
			builder: func() map[string]int {
				return map[string]int{
					"a": 1,
					"b": 2,
					"c": 3,
				}
			},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provided := tc.builder()

			result := Keys(provided)
			require.Len(t, result, len(tc.expected))

			for _, v := range result {
				require.Contains(t, tc.expected, v)
			}
		})
	}
}

func TestMaps_Keys_Int(t *testing.T) {
	testCases := []keysTestCase[int, int]{
		{
			name: "empty",
			builder: func() map[int]int {
				return make(map[int]int, 0)
			},
			expected: []int{},
		},
		{
			name: "valid",
			builder: func() map[int]int {
				return map[int]int{
					1: 2,
					3: 4,
					5: 6,
				}
			},
			expected: []int{1, 3, 5},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provided := tc.builder()

			result := Keys(provided)
			require.Len(t, result, len(tc.expected))

			for _, v := range result {
				require.Contains(t, tc.expected, v)
			}
		})
	}
}

func TestMaps_Keys_Struct(t *testing.T) {
	type customStruct struct {
		ID   string
		Name string
	}

	testCases := []keysTestCase[customStruct, int]{
		{
			name: "empty",
			builder: func() map[customStruct]int {
				return make(map[customStruct]int, 0)
			},
			expected: []customStruct{},
		},
		{
			name: "valid",
			builder: func() map[customStruct]int {
				return map[customStruct]int{
					{ID: "1", Name: "test"}: 1,
				}
			},
			expected: []customStruct{{ID: "1", Name: "test"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provided := tc.builder()

			result := Keys(provided)
			require.Len(t, result, len(tc.expected))

			for _, v := range result {
				require.Contains(t, tc.expected, v)
			}
		})
	}
}

func TestMaps_Values_String(t *testing.T) {
	testCases := []valuesTestCase[int, string]{
		{
			name: "empty",
			builder: func() map[int]string {
				return make(map[int]string, 0)
			},
			expected: []string{},
		},
		{
			name: "valid",
			builder: func() map[int]string {
				return map[int]string{
					1: "a",
					3: "b",
					5: "c",
				}
			},
			expected: []string{"a", "b", "c"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provided := tc.builder()

			result := Values(provided)
			require.Len(t, result, len(tc.expected))

			for _, v := range result {
				require.Contains(t, tc.expected, v)
			}
		})
	}
}

func TestMaps_Values_Int(t *testing.T) {
	testCases := []valuesTestCase[int, int]{
		{
			name: "empty",
			builder: func() map[int]int {
				return make(map[int]int, 0)
			},
			expected: []int{},
		},
		{
			name: "valid",
			builder: func() map[int]int {
				return map[int]int{
					1: 2,
					3: 4,
					5: 6,
				}
			},
			expected: []int{2, 4, 6},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provided := tc.builder()

			result := Values(provided)
			require.Len(t, result, len(tc.expected))

			for _, v := range result {
				require.Contains(t, tc.expected, v)
			}
		})
	}
}

func TestMaps_Values_Struct(t *testing.T) {
	type customStruct struct {
		ID   string
		Name string
	}

	testCases := []valuesTestCase[int, customStruct]{
		{
			name: "empty",
			builder: func() map[int]customStruct {
				return make(map[int]customStruct, 0)
			},
			expected: []customStruct{},
		},
		{
			name: "valid",
			builder: func() map[int]customStruct {
				return map[int]customStruct{
					1: {ID: "1", Name: "test"},
				}
			},
			expected: []customStruct{{ID: "1", Name: "test"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provided := tc.builder()

			result := Values(provided)
			require.Len(t, result, len(tc.expected))

			for _, v := range result {
				require.Contains(t, tc.expected, v)
			}
		})
	}
}
