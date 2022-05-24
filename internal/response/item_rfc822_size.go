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

func (s *itemRFC822Size) String() string {
	return fmt.Sprintf("RFC822.SIZE %v", s.size)
}
