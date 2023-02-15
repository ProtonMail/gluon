package command

import (
	"fmt"
	"github.com/bradenaw/juniper/xslices"
	"strconv"
	"strings"
)

type FetchAttributeAll struct{}

func (f FetchAttributeAll) String() string {
	return "ALL"
}

type FetchAttributeFull struct{}

func (f FetchAttributeFull) String() string {
	return "FULL"
}

type FetchAttributeFast struct{}

func (f FetchAttributeFast) String() string {
	return "FAST"
}

type FetchAttributeEnvelope struct{}

func (f FetchAttributeEnvelope) String() string {
	return "ENVELOPE"
}

type FetchAttributeFlags struct{}

func (f FetchAttributeFlags) String() string {
	return "FLAGS"
}

type FetchAttributeInternalDate struct{}

func (f FetchAttributeInternalDate) String() string {
	return "INTERNALDATE"
}

type FetchAttributeRFC822Header struct{}

func (f FetchAttributeRFC822Header) String() string {
	return "RFC822.Header"
}

type FetchAttributeRFC822Size struct{}

func (f FetchAttributeRFC822Size) String() string {
	return "RFC822.Size"
}

type FetchAttributeRFC822 struct{}

func (f FetchAttributeRFC822) String() string {
	return "RFC822"
}

type FetchAttributeRFC822Text struct{}

func (f FetchAttributeRFC822Text) String() string {
	return "RFC822.Text"
}

type FetchAttributeBodyStructure struct{}

func (f FetchAttributeBodyStructure) String() string {
	return "BODYSTRUCTURE"
}

type FetchAttributeBody struct{}

func (f FetchAttributeBody) String() string {
	return "BODY"
}

type FetchAttributeUID struct{}

func (f FetchAttributeUID) String() string {
	return "UID"
}

type BodySection interface {
	String() string
}

type BodySectionPartial struct {
	Offset int64
	Count  int64
}

type FetchAttributeBodySection struct {
	Section BodySection
	Peek    bool
	Partial *BodySectionPartial
}

func (f FetchAttributeBodySection) String() string {
	var firstPart = "BODY"
	if f.Peek {
		firstPart += ".PEEK"
	}

	if f.Section == nil {
		return fmt.Sprintf("%v[]", firstPart)
	}

	return fmt.Sprintf("%v[%v]", firstPart, f.Section)
}

type BodySectionHeader struct{}

func (b BodySectionHeader) String() string {
	return "HEADER"
}

type BodySectionHeaderFields struct {
	Negate bool
	Fields []string
}

func (b BodySectionHeaderFields) String() string {
	negateStr := " "

	if b.Negate {
		negateStr = " NOT "
	}

	return fmt.Sprintf("HEADER.FIELDS%v%v", negateStr, b.Fields)
}

type BodySectionText struct{}

func (b BodySectionText) String() string {
	return "TEXT"
}

type BodySectionMIME struct{}

func (b BodySectionMIME) String() string {
	return "MIME"
}

type BodySectionPart struct {
	Part    []int
	Section BodySection
}

func (b BodySectionPart) String() string {
	partText := strings.Join(xslices.Map(b.Part, func(v int) string {
		return strconv.FormatInt(int64(v), 10)
	}), `.`)

	if b.Section == nil {
		return partText
	}

	return fmt.Sprintf("%v.%v", partText, b.Section.String())
}
