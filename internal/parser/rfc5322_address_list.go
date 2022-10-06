package parser

/*
#include "src/rfc5322/rfc5322_parser_capi.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"net/mail"
	"unsafe"
)

type RFC5322AddressListParser struct {
	parser *C.RFC5322AddressList
}

func NewRFC5322AddressListParser() *RFC5322AddressListParser {
	return &RFC5322AddressListParser{parser: C.RFC5322AddressList_new()}
}

func (p *RFC5322AddressListParser) Close() {
	C.RFC5322AddressList_free(p.parser)
}

func (p *RFC5322AddressListParser) Parse(input string) ([]*mail.Address, error) {
	cstr := C.CString(input)

	defer C.free(unsafe.Pointer(cstr))

	addressCount := int(C.RFC5322AddressList_parse(p.parser, cstr))
	if addressCount < 0 {
		return nil, fmt.Errorf(C.GoString(C.RFC5322AddressList_error_str(p.parser)))
	}

	result := make([]*mail.Address, addressCount)

	for i := 0; i < addressCount; i++ {
		addr := C.RFC5322AddressList_get(p.parser, C.int(i))

		result[i] = &mail.Address{
			Name:    C.GoString(addr.name),
			Address: C.GoString(addr.address),
		}
	}

	return result, nil
}

func ParseRFC5322AddressList(input string) ([]*mail.Address, error) {
	parser := NewRFC5322AddressListParser()
	defer parser.Close()

	return parser.Parse(input)
}
