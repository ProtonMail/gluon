// Package liner handles reading lines from clients that may or may not require continuation.
package liner

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"sync"

	"github.com/google/uuid"
)

// rxLiteral matches a line that ends in a literal length indicator.
var rxLiteral = regexp.MustCompile(`\{(\d+)\}\r\n$`)

type Liner struct {
	br *bufio.Reader
	mu sync.RWMutex
}

type Line struct {
	Line     []byte
	Literals map[string][]byte
}

func New(r io.Reader) *Liner {
	return &Liner{br: bufio.NewReader(r)}
}

// Reset resets the liner to read from a new reader.
func (l *Liner) Reset(r io.Reader) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.br.Reset(r)
}

// Lines returns a channel that will receive lines as they are read.
func (l *Liner) Lines(doContinuation func() error) <-chan Line {
	ch := make(chan Line)

	go func() {
		defer close(ch)

		for {
			line, lits, err := l.Read(doContinuation)
			if err != nil {
				return
			}

			ch <- Line{line, lits}
		}
	}()

	return ch
}

// Read reads a full line, automatically reading again if the line was not complete.
// Each time an additional read is performed, doContinuation is called.
// If the callback returns an error, the operation is aborted.
func (l *Liner) Read(doContinuation func() error) ([]byte, map[string][]byte, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	line, err := l.br.ReadBytes('\n')
	if err != nil {
		return nil, nil, err
	}

	lits := make(map[string][]byte)

	for {
		length := shouldReadLiteral(line)
		if length == 0 {
			break
		}

		if err := doContinuation(); err != nil {
			return nil, nil, err
		}

		uuid := uuid.New().String()

		lits[uuid] = make([]byte, length)

		if _, err := io.ReadFull(l.br, lits[uuid]); err != nil {
			return nil, nil, err
		}

		line = append(line, uuid...)

		rest, err := l.br.ReadBytes('\n')
		if err != nil {
			return nil, nil, err
		}

		line = append(line, rest...)
	}

	return line, lits, nil
}

func shouldReadLiteral(line []byte) int {
	match := rxLiteral.FindSubmatch(line)

	if match != nil {
		length, err := strconv.Atoi(string(match[1]))
		if err != nil {
			panic("bad line")
		}

		return length
	}

	return 0
}
