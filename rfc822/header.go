package rfc822

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

	parser := NewHeaderParser(rawHeader)

	for {
		entry, err := parser.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return "", err
			}
		}

		if !entry.HasKey() {
			continue
		}

		if !strings.EqualFold(key, string(entry.GetKey(rawHeader))) {
			continue
		}

		return mergeMultiline(entry.GetValue(rawHeader)), nil
	}

	return "", nil
}

var (
	errNonASCIIHeaderKey = fmt.Errorf("header key contains invalid characters")
	errKeyNotFound       = fmt.Errorf("invalid header key")
)

type ParsedHeaderEntry struct {
	keyStart   int
	keyEnd     int
	valueStart int
	valueEnd   int
}

func (p ParsedHeaderEntry) HasKey() bool {
	return p.keyStart != p.keyEnd
}

func (p ParsedHeaderEntry) GetKey(header []byte) []byte {
	return header[p.keyStart:p.keyEnd]
}

func (p ParsedHeaderEntry) GetValue(header []byte) []byte {
	return header[p.valueStart:p.valueEnd]
}

type HeaderParser struct {
	header []byte
	offset int
}

func (hp *HeaderParser) Next() (ParsedHeaderEntry, error) {
	headerLen := len(hp.header)

	if hp.offset >= headerLen {
		return ParsedHeaderEntry{}, io.EOF
	}

	result := ParsedHeaderEntry{
		keyStart:   hp.offset,
		keyEnd:     -1,
		valueStart: -1,
		valueEnd:   -1,
	}

	// Detect key, have to handle prelude case where there is no header information or last empty new line.
	{
		for hp.offset < headerLen {
			if hp.header[hp.offset] == ':' {
				prevOffset := hp.offset
				hp.offset++
				if hp.offset < headerLen && (hp.header[hp.offset] == ' ' || hp.header[hp.offset] == '\r' || hp.header[hp.offset] == '\n') {
					result.keyEnd = prevOffset

					// Validate the header key.
					for i := result.keyStart; i < result.keyEnd; i++ {
						v := hp.header[i]
						if v < 33 || v > 126 {
							return ParsedHeaderEntry{}, errNonASCIIHeaderKey
						}
					}

					break
				}
			} else if hp.header[hp.offset] == '\n' {
				hp.offset++
				result.keyEnd = result.keyStart
				result.valueStart = result.keyStart
				result.valueEnd = hp.offset
				return result, nil
			} else {
				hp.offset++
			}
		}

	}

	// collect value.
	searchOffset := result.keyEnd + 1
	result.valueStart = searchOffset

	for searchOffset < headerLen {
		// consume all content in between two quotes.
		if hp.header[searchOffset] == '"' {
			searchOffset++
			for searchOffset < headerLen && hp.header[searchOffset] != '"' {
				searchOffset++
			}
			searchOffset++

			continue
		} else if hp.header[searchOffset] == '\n' {
			searchOffset++
			// if folding the next line has to start with space or tab.
			if searchOffset < headerLen && (hp.header[searchOffset] != ' ' && hp.header[searchOffset] != '\t') {
				result.valueEnd = searchOffset
				break
			}
		} else {
			searchOffset++
		}
	}

	hp.offset = searchOffset

	// handle case where we may have reached EOF without concluding any previous processing.
	if result.valueEnd == -1 && searchOffset >= headerLen {
		result.valueEnd = headerLen
	}

	return result, nil
}

func NewHeaderParser(header []byte) HeaderParser {
	return HeaderParser{header: header}
}

func ParseHeader(header []byte) (*Header, error) {
	parser := NewHeaderParser(header)

	var lines [][]byte

	for {
		entry, err := parser.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return nil, err
			}
		}

		lines = append(lines, header[entry.keyStart:entry.valueEnd])
	}

	return &Header{lines: lines}, nil
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
