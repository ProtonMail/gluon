package utils

import (
	"maps"
	"slices"
)

// Keys returns a slice of keys from the map.
// Alternative to using maps.Keys which returns an iterator instead of a slice.
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	keys := make([]K, 0, len(m))

	return slices.AppendSeq(
		keys,
		maps.Keys(m),
	)
}

// Values returns a slice of values from the map.
// Alternative to using maps.Values which returns an iterator instead of a slice.
func Values[M ~map[K]V, K comparable, V any](m M) []V {
	values := make([]V, 0, len(m))

	return slices.AppendSeq(
		values,
		maps.Values(m),
	)
}
