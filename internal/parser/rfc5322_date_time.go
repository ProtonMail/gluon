package parser

/*
#include "src/rfc5322/rfc5322_parser_capi.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

func ParseRFC5322DateTime(input string) (time.Time, error) {
	cstr := C.CString(input)
	defer C.free(unsafe.Pointer(cstr))

	dateTime := C.RFC5322DateTime{}
	if C.RFC5322DateTime_parse(&dateTime, cstr) != 0 {
		return time.Time{}, fmt.Errorf("failed to parse date time")
	}

	var tz *time.Location

	if dateTime.tzType == C.TZ_TYPE_OFFSET {
		var sign rune

		var multiplier int

		if uint32(dateTime.tz)&(uint32(1)<<31) != 0 {
			sign = '+'
			multiplier = 1
		} else {
			sign = '-'
			multiplier = -1
		}

		hours := int((dateTime.tz >> 8) & 0xFF)
		min := int(dateTime.tz & 0xFF)

		repr := fmt.Sprintf("%c%02d%02d", sign, hours, min)
		duration := multiplier * (hours*60*60 + min*60)
		tz = time.FixedZone(repr, duration)
	} else {
		switch dateTime.tz {
		case C.TZ_CODE_UT:
			tz = time.FixedZone("UT", 0)
		case C.TZ_CODE_UTC:
			tz = time.FixedZone("UTC", 0)
		case C.TZ_CODE_GMT:
			tz = time.FixedZone("GMT", 0)
		case C.TZ_CODE_EST:
			tz = time.FixedZone("EST", -5*60*60)
		case C.TZ_CODE_EDT:
			tz = time.FixedZone("EDT", -4*60*60)
		case C.TZ_CODE_CST:
			tz = time.FixedZone("CST", -6*60*60)
		case C.TZ_CODE_CDT:
			tz = time.FixedZone("CDT", -5*60*60)
		case C.TZ_CODE_MST:
			tz = time.FixedZone("MST", -7*60*60)
		case C.TZ_CODE_MDT:
			tz = time.FixedZone("MDT", -6*60*60)
		case C.TZ_CODE_PST:
			tz = time.FixedZone("PST", -8*60*60)
		case C.TZ_CODE_PDT:
			tz = time.FixedZone("PDT", -7*60*60)
		default:
			return time.Time{}, fmt.Errorf("unknown timezone")
		}
	}

	return time.Date(int(dateTime.year), time.Month(dateTime.month), int(dateTime.day), int(dateTime.hour), int(dateTime.min), int(dateTime.sec), 0, tz), nil
}
