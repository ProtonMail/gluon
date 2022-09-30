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

func ParseRFC5322AddressList(input string) ([]*mail.Address, error) {
	parser := C.RFC5322AddressList_new()
	defer C.RFC5322AddressList_free(parser)

	cstr := C.CString(input)
	defer C.free(unsafe.Pointer(cstr))

	addressCount := int(C.RFC5322AddressList_parse(parser, cstr))
	if addressCount < 0 {
		return nil, fmt.Errorf(C.GoString(C.RFC5322AddressList_error_str(parser)))
	}

	result := make([]*mail.Address, addressCount)

	for i := 0; i < addressCount; i++ {
		addr := C.RFC5322AddressList_get(parser, C.int(i))

		result[i] = &mail.Address{
			Name:    C.GoString(addr.name),
			Address: C.GoString(addr.address),
		}
	}

	return result, nil
}
