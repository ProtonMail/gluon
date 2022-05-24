package response

import "fmt"

type itemRFC822Literal struct {
	literal []byte
}

func ItemRFC822Literal(literal []byte) *itemRFC822Literal {
	return &itemRFC822Literal{
		literal: literal,
	}
}

func (r *itemRFC822Literal) String() string {
	return fmt.Sprintf("RFC822 {%v}\r\n%s", len(r.literal), r.literal)
}
