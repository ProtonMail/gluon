package response

import "fmt"

type itemRFC822Size struct {
	size int
}

func ItemRFC822Size(size int) *itemRFC822Size {
	return &itemRFC822Size{
		size: size,
	}
}

func (s *itemRFC822Size) Strings() (raw string, _ string) {
	raw = fmt.Sprintf("RFC822.SIZE %v", s.size)
	return raw, raw
}
