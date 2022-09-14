package rfc822

import (
	"bytes"
	"net/textproto"
	"regexp"
	"strings"

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
	rawHeader, body := Split(literal)

	header, err := ParseHeader(rawHeader)
	if err != nil {
		return nil, err
	}

	header.Set(key, val)

	return append(header.Raw(), body...), nil
}

// GetHeaderValue is a helper method that queries a header value in a message literal.
func GetHeaderValue(literal []byte, key string) (string, error) {
	rawHeader, _ := Split(literal)

	header, err := ParseHeader(rawHeader)
	if err != nil {
		return "", err
	}

	return header.Get(key), nil
}

// TODO: This is shitty -- should use a real generated parser like the IMAP parser or the RFC5322 parser.
func ParseHeader(header []byte) (*Header, error) {
	var (
		lines [][]byte
		quote int
	)

	forLines(header, func(line []byte) {
		split := splitLine(line)

		switch {
		case len(bytes.Trim(line, "\r\n")) == 0:
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
	})

	return &Header{lines: lines}, nil
}

func forLines(lines []byte, fn func(line []byte)) {
	remaining := lines
	separator := []byte{'\n'}

	for {
		index := bytes.Index(remaining, separator)
		if index < 0 {
			if len(remaining) > 0 {
				fn(remaining)
			}

			return
		}

		line := remaining[0 : index+1]
		remaining = remaining[index+1:]

		fn(line)
	}
}

func mergeMultiline(line []byte) string {
	remaining := line

	builder := strings.Builder{}
	separator := []byte{'\n'}

	for len(remaining) != 0 {
		index := bytes.Index(remaining, separator)
		if index < 0 {
			builder.Write(bytes.TrimSpace(remaining))
			break
		}

		var section []byte
		if index >= 1 && remaining[index-1] == '\r' {
			section = remaining[0 : index-1]
		} else {
			section = remaining[0:index]
		}

		remaining = remaining[index+1:]

		if len(section) != 0 {
			builder.Write(bytes.TrimSpace(section))

			if len(remaining) != 0 {
				builder.WriteRune(' ')
			}
		}
	}

	return builder.String()
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
