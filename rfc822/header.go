package rfc822

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net/textproto"
	"regexp"
	"strings"

	"github.com/bradenaw/juniper/iterator"
	"golang.org/x/exp/slices"
)

var rxWhitespace = regexp.MustCompile(`^\s+`)

type Header struct {
	lines [][]byte
}

func (h *Header) Raw() []byte {
	return bytes.Join(h.lines, nil)
}

func (h *Header) Has(key string) bool {
	for _, line := range h.lines {
		split := splitLine(line)

		if len(split) != 2 {
			continue
		}

		if !strings.EqualFold(string(split[0]), key) {
			continue
		}

		return true
	}

	return false
}

func (h *Header) Get(key string) string {
	return mergeMultiline(h.GetRaw(key))
}

func (h *Header) GetRaw(key string) []byte {
	split := splitLine(h.GetLine(key))

	if len(split) != 2 {
		return nil
	}

	return split[1]
}

func (h *Header) GetLine(key string) []byte {
	for _, line := range h.lines {
		split := splitLine(line)

		if len(split) != 2 {
			continue
		}

		if !strings.EqualFold(string(split[0]), key) {
			continue
		}

		return line
	}

	return nil
}

// TODO: Is it okay to add new entries to the front? Probably not, would break sigs or something.
func (h *Header) Set(key, val string) {
	for index, line := range h.lines {
		if split := splitLine(line); len(split) == 2 {
			if strings.EqualFold(string(split[0]), key) {
				// Override the existing key.
				h.lines[index] = joinLine([]byte(textproto.CanonicalMIMEHeaderKey(key)), []byte(val))

				// TODO: How to handle duplicate header values?
				return
			}
		}
	}

	h.lines = slices.Insert(h.lines, 0, joinLine([]byte(key), []byte(val)))
}

func (h *Header) Del(key string) {
	for index, line := range h.lines {
		split := splitLine(line)

		if len(split) != 2 {
			continue
		}

		if !strings.EqualFold(string(split[0]), key) {
			continue
		}

		h.lines = append(h.lines[:index], h.lines[index+1:]...)

		return
	}
}

func (h *Header) Fields(fields []string) []byte {
	wantFields := make(map[string]struct{})

	for _, field := range fields {
		wantFields[strings.ToLower(field)] = struct{}{}
	}

	var res []byte

	for _, line := range h.lines {
		if len(bytes.TrimSpace(line)) == 0 {
			res = append(res, line...)
		} else {
			split := splitLine(line)

			if len(split) != 2 {
				continue
			}

			if _, ok := wantFields[string(bytes.ToLower(split[0]))]; ok {
				res = append(res, line...)
			}
		}
	}

	return res
}

func (h *Header) FieldsNot(fields []string) []byte {
	wantFieldsNot := make(map[string]struct{})

	for _, field := range fields {
		wantFieldsNot[strings.ToLower(field)] = struct{}{}
	}

	var res []byte

	for _, line := range h.lines {
		if len(bytes.TrimSpace(line)) == 0 {
			res = append(res, line...)
		} else {
			split := splitLine(line)

			if len(split) != 2 {
				continue
			}

			if _, ok := wantFieldsNot[string(bytes.ToLower(split[0]))]; !ok {
				res = append(res, line...)
			}
		}
	}

	return res
}

func (h *Header) Entries(fn func(key, val string)) {
	for _, line := range h.lines {
		split := splitLine(line)

		if len(split) != 2 {
			continue
		}

		fn(string(split[0]), mergeMultiline(split[1]))
	}
}

// SetHeaderValue is a helper method that sets a header value in a message literal.
func SetHeaderValue(literal []byte, key, val string) ([]byte, error) {
	rawHeader, body, err := Split(literal)
	if err != nil {
		return nil, err
	}

	header := ParseHeader(rawHeader)

	header.Set(key, val)

	return append(header.Raw(), body...), nil
}

// GetHeaderValue is a helper method that queries a header value in a message literal.
func GetHeaderValue(literal []byte, key string) (string, error) {
	rawHeader, _, err := Split(literal)
	if err != nil {
		return "", err
	}

	return ParseHeader(rawHeader).Get(key), nil
}

// TODO: This is shitty -- should use a real generated parser like the IMAP parser or the RFC5322 parser.
func ParseHeader(header []byte) *Header {
	var (
		lines [][]byte
		quote int
	)

	lineIt := iterator.Chan(forLines(bufio.NewReader(bytes.NewReader(header))))

	for {
		line, ok := lineIt.Next()
		if !ok {
			break
		}

		split := splitLine(line)

		switch {
		case len(bytes.TrimSpace(line)) == 0:
			lines = append(lines, line)

		case quote%2 != 0, rxWhitespace.Match(split[0]), len(split) != 2:
			if len(lines) > 0 {
				lines[len(lines)-1] = append(lines[len(lines)-1], line...)
			} else {
				lines = append(lines, line)
			}

		default:
			lines = append(lines, line)
		}

		quote += bytes.Count(line, []byte(`"`))
	}

	return &Header{lines: lines}
}

func forLines(br *bufio.Reader) chan []byte {
	ch := make(chan []byte)

	go func() {
		defer close(ch)

		for {
			b, err := br.ReadBytes('\n')
			if err != nil {
				if !errors.Is(err, io.EOF) {
					panic(err)
				}

				if len(b) > 0 {
					ch <- b
				}

				return
			}

			ch <- b
		}
	}()

	return ch
}

// TODO: This is terrible! Need to properly handle multiline fields.
func mergeMultiline(line []byte) string {
	var res [][]byte

	for line := range forLines(bufio.NewReader(bytes.NewReader(line))) {
		trimmed := bytes.TrimSpace(line)
		if trimmed != nil {
			res = append(res, trimmed)
		}
	}

	return string(bytes.Join(res, []byte(" ")))
}

func splitLine(line []byte) [][]byte {
	result := bytes.SplitN(line, []byte(`:`), 2)

	if len(result) > 1 && len(result[1]) > 0 && result[1][0] == ' ' {
		result[1] = result[1][1:]
	}

	return result
}

// TODO: Don't assume line ending is \r\n. Bad.
func joinLine(key, val []byte) []byte {
	return []byte(string(key) + ": " + string(val) + "\r\n")
}
