package response

import "fmt"

type itemBody struct {
	structure string
}

func ItemBody(structure string) *itemBody {
	return &itemBody{
		structure: structure,
	}
}

func (r *itemBody) String() string {
	return fmt.Sprintf("BODY %v", r.structure)
}
