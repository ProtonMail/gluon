// Package utils provide common slice/maps utilities
package utils

import "slices"

// Filter returns a new slice which is filtered by the provided keep function.
func Filter[S ~[]E, E any](s S, keep func(E) bool) S {
	return slices.DeleteFunc(slices.Clone(s), func(e E) bool {
		return !keep(e)
	})
}
