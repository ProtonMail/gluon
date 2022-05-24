package response

import "fmt"

type itemRFC822Header struct {
	header []byte
}

func ItemRFC822Header(header []byte) *itemRFC822Header {
	return &itemRFC822Header{
		header: header,
	}
}

func (r *itemRFC822Header) String() string {
	return fmt.Sprintf("RFC822.HEADER {%v}\r\n%s", len(r.header), r.header)
}
