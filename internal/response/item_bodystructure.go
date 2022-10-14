package response

import "fmt"

type itemBodyStructure struct {
	structure string
}

func ItemBodyStructure(structure string) *itemBodyStructure {
	return &itemBodyStructure{
		structure: structure,
	}
}

func (r *itemBodyStructure) Strings() (raw string, _ string) {
	raw = fmt.Sprintf("BODYSTRUCTURE %v", r.structure)
	return raw, raw
}
