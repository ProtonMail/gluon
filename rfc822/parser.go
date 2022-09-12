package rfc822

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
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

func Parse(literal []byte) (*Section, error) {
	return parse(literal, []int{}, 0, len(literal))
}

func (section *Section) Identifier() []int {
	return section.identifier
}

func (section *Section) ContentType() (string, map[string]string, error) {
	header, err := section.ParseHeader()
	if err != nil {
		return "", nil, err
	}

	return ParseContentType(header.Get("Content-Type"))
}

func (section *Section) Header() []byte {
	return section.literal[section.header:section.body]
}

func (section *Section) ParseHeader() (*Header, error) {
	if section.parsedHeader == nil {
		h, err := ParseHeader(section.Header())
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

func (section *Section) Children() []*Section {
	if len(section.children) == 0 {
		if err := section.load(); err != nil {
			panic(err)
		}
	}

	return section.children
}

func (section *Section) Part(identifier ...int) *Section {
	if len(identifier) > 0 {
		children := section.Children()

		if identifier[0] <= 0 || identifier[0]-1 > len(children) {
			return nil
		}

		if len(children) != 0 {
			return children[identifier[0]-1].Part(identifier[1:]...)
		}
	}

	return section
}

func (section *Section) load() error {
	contentType, contentParams, err := section.ContentType()
	if err != nil {
		return err
	}

	if MIMEType(contentType) == MessageRFC822 {
		child, err := parse(
			section.literal[section.body:section.end],
			section.identifier,
			0,
			section.end-section.body,
		)
		if err != nil {
			return err
		}

		if err := child.load(); err != nil {
			return err
		}

		section.children = append(section.children, child.children...)
	} else if strings.HasPrefix(contentType, "multipart/") {
		scanner, err := NewScanner(bytes.NewReader(section.literal[section.body:section.end]), contentParams["boundary"])
		if err != nil {
			return err
		}

		res, err := scanner.ScanAll()
		if err != nil {
			return err
		}

		for idx, res := range res {
			child, err := parse(
				section.literal,
				append(section.identifier, idx+1),
				section.body+res.Offset,
				section.body+res.Offset+len(res.Data),
			)
			if err != nil {
				return err
			}

			section.children = append(section.children, child)
		}
	}

	return nil
}

func Split(b []byte) ([]byte, []byte, error) {
	br := bufio.NewReader(bytes.NewReader(b))

	var header []byte

	for {
		b, err := br.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				panic(err)
			}

			if len(b) > 0 {
				header = append(header, b...)
			}

			break
		}

		header = append(header, b...)

		if len(bytes.Trim(b, "\r\n")) == 0 {
			break
		}
	}

	body, err := io.ReadAll(br)
	if err != nil {
		return nil, nil, err
	}

	return header, body, nil
}

func parse(literal []byte, identifier []int, begin, end int) (*Section, error) {
	header, _, err := Split(literal[begin:end])
	if err != nil {
		return nil, err
	}

	return &Section{
		identifier: identifier,
		literal:    literal,
		header:     begin,
		body:       begin + len(header),
		end:        end,
	}, nil
}
