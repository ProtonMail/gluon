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

func (r *itemBody) Strings() (raw string, _ string) {
	raw = fmt.Sprintf("BODY %v", r.structure)
	return raw, raw
}
