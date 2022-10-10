// Package liner handles reading lines from clients that may or may not require continuation.
package liner

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// rxLiteral matches a line that ends in a literal length indicator.
var rxLiteral = regexp.MustCompile(`\{(\d+)\}\r\n$`)

type Liner struct {
	br *bufio.Reader
}

func New(r io.Reader) *Liner {
	return &Liner{br: bufio.NewReader(r)}
}

// Read reads a full line, automatically reading again if the line was not complete.
// Each time an additional read is performed, doContinuation is called.
// If the callback returns an error, the operation is aborted.
func (l *Liner) Read(doContinuation func() error) (string, error) {
	builder := strings.Builder{}

	line, err := l.br.ReadBytes('\n')
	if err != nil {
		return "", err
	}

	builder.Write(line)

	for {
		length := shouldReadLiteral(line)
		if length == 0 {
			break
		}

		if err := doContinuation(); err != nil {
			return "", err
		}

		buffer := make([]byte, length)

		if _, err := io.ReadFull(l.br, buffer); err != nil {
			return "", err
		}

		builder.Write(buffer)

		rest, err := l.br.ReadBytes('\n')
		if err != nil {
			return "", err
		}

		builder.Write(rest)
		line = rest
	}

	return builder.String(), nil
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
