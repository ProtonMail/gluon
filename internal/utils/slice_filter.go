package utils

import "slices"

func Filter[S ~[]E, E any](s S, keep func(E) bool) S {
	return slices.DeleteFunc(slices.Clone(s), func(e E) bool {
		return !keep(e)
	})
}
