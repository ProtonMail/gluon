package response

import "fmt"

type itemBodyLiteral struct {
	section string
	literal []byte
	partial int
}

func ItemBodyLiteral(section string, literal []byte) *itemBodyLiteral {
	return &itemBodyLiteral{
		section: section,
		literal: literal,
		partial: -1,
	}
}

func (r *itemBodyLiteral) WithPartial(begin, count int) *itemBodyLiteral {
	r.partial = begin

	if literalLen := len(r.literal); begin >= literalLen {
		r.literal = nil
	} else if begin+count > literalLen {
		r.literal = r.literal[begin:]
	} else {
		r.literal = r.literal[begin : begin+count]
	}

	return r
}

func (r *itemBodyLiteral) String(isPrivateByDefault bool) string {
	if isPrivateByDefault {
		return ""
	}

	var partial string

	if r.partial >= 0 {
		partial = fmt.Sprintf("<%v>", r.partial)
	}

	return fmt.Sprintf("BODY[%v]%v {%v}\r\n%s", r.section, partial, len(r.literal), r.literal)
}
