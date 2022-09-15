package rfc822

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"strings"
)

type headerEntry struct {
	ParsedHeaderEntry
	mapKey string
	merged string
	prev   *headerEntry
	next   *headerEntry
}

func (he *headerEntry) getMerged(data []byte) string {
	if len(he.merged) == 0 {
		he.merged = mergeMultiline(he.GetValue(data))
	}

	return he.merged
}

type Header struct {
	keys       map[string][]*headerEntry
	firstEntry *headerEntry
	lastEntry  *headerEntry
	data       []byte
}

func NewHeader(data []byte) (*Header, error) {
	h := &Header{
		keys: make(map[string][]*headerEntry),
		data: data,
	}

	parser := newHeaderParser(data)

	for {
		entry, err := parser.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return nil, err
			}
		}

		hentry := &headerEntry{
			ParsedHeaderEntry: entry,
			merged:            "",
			next:              nil,
		}

		if entry.HasKey() {
			hashKey := strings.ToLower(string(entry.GetKey(data)))
			hentry.mapKey = hashKey

			if v, ok := h.keys[hashKey]; !ok {
				h.keys[hashKey] = []*headerEntry{hentry}
			} else {
				h.keys[hashKey] = append(v, hentry)
			}
		}

		if h.firstEntry == nil {
			h.firstEntry = hentry
			h.lastEntry = hentry
		} else {
			h.lastEntry.next = hentry
			hentry.prev = h.lastEntry
			h.lastEntry = hentry
		}
	}

	return h, nil
}

func (h *Header) Raw() []byte {
	return h.data
}

func (h *Header) Has(key string) bool {
	_, ok := h.keys[strings.ToLower(key)]

	return ok
}

func (h *Header) GetChecked(key string) (string, bool) {
	v, ok := h.keys[strings.ToLower(key)]
	if !ok {
		return "", false
	}

	return v[0].getMerged(h.data), true
}

func (h *Header) Get(key string) string {
	v, ok := h.keys[strings.ToLower(key)]
	if !ok {
		return ""
	}

	return v[0].getMerged(h.data)
}

func (h *Header) GetLine(key string) []byte {
	v, ok := h.keys[strings.ToLower(key)]
	if !ok {
		return nil
	}

	return v[0].GetAll(h.data)
}

func (h *Header) getLines() [][]byte {
	var res [][]byte
	for e := h.firstEntry; e != nil; e = e.next {
		res = append(res, h.data[e.keyStart:e.valueEnd])
	}

	return res
}

func (h *Header) GetRaw(key string) []byte {
	v, ok := h.keys[strings.ToLower(key)]
	if !ok {
		return nil
	}

	return v[0].GetValue(h.data)
}

func (h *Header) Set(key, val string) {
	// We can only add entries to the front of the header.
	key = textproto.CanonicalMIMEHeaderKey(key)
	mapKey := strings.ToLower(key)

	keyBytes := []byte(key)

	entryBytes := joinLine([]byte(key), []byte(val))
	newHeaderEntry := &headerEntry{
		ParsedHeaderEntry: ParsedHeaderEntry{
			keyStart:   0,
			keyEnd:     len(keyBytes),
			valueStart: len(keyBytes) + 2,
			valueEnd:   len(entryBytes),
		},
		mapKey: mapKey,
	}

	if v, ok := h.keys[mapKey]; !ok {
		h.keys[mapKey] = []*headerEntry{newHeaderEntry}
	} else {
		h.keys[mapKey] = append([]*headerEntry{newHeaderEntry}, v...)
	}

	if h.firstEntry == nil {
		h.data = entryBytes
		h.firstEntry = newHeaderEntry
	} else {
		insertOffset := h.firstEntry.keyStart
		newHeaderEntry.next = h.firstEntry
		h.firstEntry.prev = newHeaderEntry
		h.firstEntry = newHeaderEntry

		buffer := bytes.Buffer{}
		if insertOffset != 0 {
			if _, err := buffer.Write(h.data[0:insertOffset]); err != nil {
				panic("failed to write to byte buffer")
			}
		}

		if _, err := buffer.Write(entryBytes); err != nil {
			panic("failed to write to byte buffer")
		}

		if _, err := buffer.Write(h.data[insertOffset:]); err != nil {
			panic("failed to write to byte buffer")
		}

		h.data = buffer.Bytes()
		h.applyOffset(newHeaderEntry.next, len(entryBytes))
	}
}

func (h *Header) Del(key string) {
	mapKey := strings.ToLower(key)

	v, ok := h.keys[mapKey]
	if !ok {
		return
	}

	he := v[0]

	if len(v) == 1 {
		delete(h.keys, mapKey)
	} else {
		h.keys[mapKey] = v[1:]
	}

	if he.prev != nil {
		he.prev.next = he.next
	}

	if he.next != nil {
		he.next.prev = he.prev
	}

	dataLen := he.valueEnd - he.keyStart

	h.data = append(h.data[0:he.keyStart], h.data[he.valueEnd:]...)

	h.applyOffset(he.next, -dataLen)
}

func (h *Header) Fields(fields []string) []byte {
	wantFields := make(map[string]struct{})

	for _, field := range fields {
		wantFields[strings.ToLower(field)] = struct{}{}
	}

	var res []byte

	for e := h.firstEntry; e != nil; e = e.next {
		if len(bytes.TrimSpace(e.GetAll(h.data))) == 0 {
			res = append(res, e.GetAll(h.data)...)
			continue
		}

		if !e.HasKey() {
			continue
		}

		_, ok := wantFields[e.mapKey]
		if !ok {
			continue
		}

		res = append(res, e.GetAll(h.data)...)
	}

	return res
}

func (h *Header) FieldsNot(fields []string) []byte {
	wantFieldsNot := make(map[string]struct{})

	for _, field := range fields {
		wantFieldsNot[strings.ToLower(field)] = struct{}{}
	}

	var res []byte

	for e := h.firstEntry; e != nil; e = e.next {
		if len(bytes.TrimSpace(e.GetAll(h.data))) == 0 {
			res = append(res, e.GetAll(h.data)...)
			continue
		}

		if !e.HasKey() {
			continue
		}

		_, ok := wantFieldsNot[e.mapKey]
		if ok {
			continue
		}

		res = append(res, e.GetAll(h.data)...)
	}

	// Since we are only applying the entries that have a key, we need to add a new line at the end.
	return res
}

func (h *Header) Entries(fn func(key, val string)) {
	for e := h.firstEntry; e != nil; e = e.next {
		if !e.HasKey() {
			continue
		}

		fn(string(e.GetKey(h.data)), e.getMerged(h.data))
	}
}

func (h *Header) applyOffset(start *headerEntry, offset int) {
	for e := start; e != nil; e = e.next {
		e.applyOffset(offset)
	}
}

// SetHeaderValue is a helper method that sets a header value in a message literal.
// It does not check whether the existing value already exists.
func SetHeaderValue(literal []byte, key, val string) ([]byte, error) {
	rawHeader, body := Split(literal)

	parser := newHeaderParser(rawHeader)

	var foundFirstEntry bool

	var parsedHeaderEntry ParsedHeaderEntry

	// find first header entry.
	for {
		entry, err := parser.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return nil, err
			}
		}

		if entry.HasKey() {
			foundFirstEntry = true
			parsedHeaderEntry = entry

			break
		}
	}

	key = textproto.CanonicalMIMEHeaderKey(key)
	data := joinLine([]byte(key), []byte(val))

	if !foundFirstEntry {
		return append(rawHeader, append(data, body...)...), nil
	} else {
		return append(literal[0:parsedHeaderEntry.keyStart], append(data, literal[parsedHeaderEntry.keyStart:]...)...), nil
	}
}

// GetHeaderValue is a helper method that queries a header value in a message literal.
func GetHeaderValue(literal []byte, key string) (string, error) {
	rawHeader, _ := Split(literal)

	parser := newHeaderParser(rawHeader)

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

func (p ParsedHeaderEntry) GetAll(header []byte) []byte {
	return header[p.keyStart:p.valueEnd]
}

func (p *ParsedHeaderEntry) applyOffset(offset int) {
	p.keyStart += offset
	p.keyEnd += offset
	p.valueStart += offset
	p.valueEnd += offset
}

type headerParser struct {
	header []byte
	offset int
}

// Next will keep parsing until it collects a new entry. io.EOF is returned when there is nothing left to parse.
func (hp *headerParser) Next() (ParsedHeaderEntry, error) {
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
	searchOffset := result.keyEnd + 2
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

func newHeaderParser(header []byte) headerParser {
	return headerParser{header: header}
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
