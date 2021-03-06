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

func (r *itemBodyStructure) String() string {
	return fmt.Sprintf("BODYSTRUCTURE %v", r.structure)
}
