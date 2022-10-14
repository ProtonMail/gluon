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

func (r *itemBody) String(_ bool) string {
	return fmt.Sprintf("BODY %v", r.structure)
}
