package parser

/*
#include "src/imap/parser_capi.h"
#include <stdlib.h>
*/
import "C"
import (
	"encoding/hex"
	"fmt"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	protobuf "google.golang.org/protobuf/proto"
	"unsafe"
)

type ParserError struct {
	error string
}

func (pe *ParserError) Error() string {
	return pe.error
}

type UnmarshalError struct {
	cmdString string
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("Failed to unmarshal command data: '%v'", e.cmdString)
}

type IMAPParser struct {
	p *C.IMAPParser
}

func NewIMAPParser() *IMAPParser {
	return &IMAPParser{p: C.IMAPParser_new()}
}

func (i *IMAPParser) Close() {
	C.IMAPParser_free(i.p)
}

func (i *IMAPParser) Parse(input string, delimiter rune) (string, *proto.Command, error) {
	cstr := C.CString(input)
	defer C.free(unsafe.Pointer(cstr))

	if r := C.IMAPParser_parse(i.p, cstr, C.char(delimiter)); r != 0 {
		return C.GoString(C.IMAPParser_getTag(i.p)), nil, &ParserError{error: C.GoString(C.IMAPParser_getError(i.p))}
	}

	tag := C.GoString(C.IMAPParser_getTag(i.p))
	cmdPtr := C.IMAPParser_getCommandData(i.p)
	cmdSize := C.IMAPParser_getCommandSize(i.p)

	cmdBytes := C.GoBytes(cmdPtr, cmdSize)

	cmd := &proto.Command{}
	if err := protobuf.Unmarshal(cmdBytes, cmd); err != nil {
		return "", nil, &UnmarshalError{cmdString: hex.EncodeToString(cmdBytes)}
	}

	return tag, cmd, nil
}
