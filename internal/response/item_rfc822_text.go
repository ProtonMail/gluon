package response

import "fmt"

type itemRFC822Text struct {
	text []byte
}

func ItemRFC822Text(text []byte) *itemRFC822Text {
	return &itemRFC822Text{
		text: text,
	}
}

func (r *itemRFC822Text) Strings() (raw string, filtered string) {
	raw = fmt.Sprintf("RFC822.TEXT {%v}\r\n%s", len(r.text), r.text)
	filtered = fmt.Sprintf("RFC822.TEXT {%v}", len(r.text))

	return raw, filtered
}
