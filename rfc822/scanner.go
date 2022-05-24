package rfc822

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

type Scanner struct {
	r *bufio.Reader

	boundary string
	progress int
}

type Part struct {
	Data   []byte
	Offset int
}

func NewScanner(r io.Reader, boundary string) (*Scanner, error) {
	scanner := &Scanner{r: bufio.NewReader(r), boundary: boundary}

	if _, _, err := scanner.readToBoundary(); err != nil {
		return nil, err
	}

	return scanner, nil
}

func (s *Scanner) ScanAll() ([]Part, error) {
	var parts []Part

	for {
		offset := s.progress

		data, more, err := s.readToBoundary()
		if err != nil {
			return nil, err
		}

		if !more {
			return parts, nil
		}

		parts = append(parts, Part{Data: data, Offset: offset})
	}
}

func (s *Scanner) readToBoundary() ([]byte, bool, error) {
	var res []byte

	for {
		line, err := s.r.ReadBytes('\n')
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return nil, false, err
			}

			if len(line) == 0 {
				return nil, false, nil
			}
		}

		s.progress += len(line)

		switch {
		case bytes.HasPrefix(bytes.TrimSpace(line), []byte("--"+s.boundary)):
			return bytes.TrimSuffix(bytes.TrimSuffix(res, []byte("\n")), []byte("\r")), true, nil

		case bytes.HasSuffix(bytes.TrimSpace(line), []byte(s.boundary+"--")):
			return bytes.TrimSuffix(bytes.TrimSuffix(res, []byte("\n")), []byte("\r")), false, nil

		default:
			res = append(res, line...)
		}
	}
}
