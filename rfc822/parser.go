package rfc822

import (
	"bytes"
	"fmt"
)

type Section struct {
	identifier   []int
	literal      []byte
	parsedHeader *Header
	header       int
	body         int
	end          int
	children     []*Section
}

func Parse(literal []byte) *Section {
	return parse(literal, []int{}, 0, len(literal))
}

func (section *Section) Identifier() []int {
	return section.identifier
}

func (section *Section) ContentType() (MIMEType, map[string]string, error) {
	header, err := section.ParseHeader()
	if err != nil {
		return "", nil, err
	}

	return parseMIMEType(header.Get("Content-Type"))
}

func (section *Section) Header() []byte {
	return section.literal[section.header:section.body]
}

func (section *Section) ParseHeader() (*Header, error) {
	if section.parsedHeader == nil {
		h, err := NewHeader(section.Header())
		if err != nil {
			return nil, err
		}

		section.parsedHeader = h
	}

	return section.parsedHeader, nil
}

func (section *Section) Body() []byte {
	return section.literal[section.body:section.end]
}

func (section *Section) Literal() []byte {
	return section.literal[section.header:section.end]
}

func (section *Section) Children() ([]*Section, error) {
	if len(section.children) == 0 {
		if err := section.load(); err != nil {
			return nil, err
		}
	}

	return section.children, nil
}

func (section *Section) Part(identifier ...int) (*Section, error) {
	if len(identifier) > 0 {
		children, err := section.Children()
		if err != nil {
			return nil, err
		}

		if identifier[0] <= 0 || identifier[0]-1 > len(children) {
			return nil, fmt.Errorf("no such part exists")
		}

		if len(children) != 0 {
			return children[identifier[0]-1].Part(identifier[1:]...)
		}
	}

	return section, nil
}

func (section *Section) load() error {
	contentType, contentParams, err := section.ContentType()
	if err != nil {
		return err
	}

	if MIMEType(contentType) == MessageRFC822 {
		child := parse(
			section.literal[section.body:section.end],
			section.identifier,
			0,
			section.end-section.body,
		)

		if err := child.load(); err != nil {
			return err
		}

		section.children = append(section.children, child.children...)
	} else if contentType.IsMultiPart() {
		scanner, err := NewByteScanner(section.literal[section.body:section.end], []byte(contentParams["boundary"]))
		if err != nil {
			return err
		}

		res := scanner.ScanAll()

		for idx, res := range res {
			child := parse(
				section.literal,
				append(section.identifier, idx+1),
				section.body+res.Offset,
				section.body+res.Offset+len(res.Data),
			)

			section.children = append(section.children, child)
		}
	}

	return nil
}

func Split(b []byte) ([]byte, []byte) {
	remaining := b
	splitIndex := int(0)
	separator := []byte{'\n'}

	for len(remaining) != 0 {
		index := bytes.Index(remaining, separator)
		if index < 0 {
			splitIndex += len(remaining)
			break
		}

		splitIndex += index + 1

		if len(bytes.Trim(remaining[0:index], "\r\n")) == 0 {
			break
		}

		remaining = remaining[index+1:]
	}

	return b[0:splitIndex], b[splitIndex:]
}

func parse(literal []byte, identifier []int, begin, end int) *Section {
	header, _ := Split(literal[begin:end])

	parsedHeader, err := NewHeader(header)
	if err != nil {
		header = nil
		parsedHeader = nil
	}

	return &Section{
		identifier:   identifier,
		literal:      literal,
		parsedHeader: parsedHeader,
		header:       begin,
		body:         begin + len(header),
		end:          end,
	}
}
