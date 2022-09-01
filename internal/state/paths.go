package state

import (
	"strings"

	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/exp/slices"
)

// listSuperiors returns all names superior to the given name, if hierarchies are indicated with the given delimiter.
func listSuperiors(name, delimiter string) []string {
	split := strings.Split(name, delimiter)
	if len(split) == 0 {
		return nil
	}

	var inferiors []string

	for i := range split {
		if i == 0 {
			continue
		}

		inferiors = append(inferiors, strings.Join(split[0:i], delimiter))
	}

	return inferiors
}

func listInferiors(parent, delimiter string, names []string) []string {
	inferiors := xslices.Filter(names, func(name string) bool {
		return slices.Contains(listSuperiors(name, delimiter), parent)
	})

	slices.Sort(inferiors)

	xslices.Reverse(inferiors)

	return inferiors
}
