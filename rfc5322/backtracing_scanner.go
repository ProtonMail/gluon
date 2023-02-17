package rfc5322

import (
	"bytes"
	"io"
)

type BacktrackingByteScanner struct {
	data   []byte
	offset int
}

func NewBacktrackingByteScanner(data []byte) *BacktrackingByteScanner {
	return &BacktrackingByteScanner{
		data: data,
	}
}

type BacktrackingByteScannerScope struct {
	offset int
}

func (bs *BacktrackingByteScanner) Read(dst []byte) (int, error) {
	thisLen := len(bs.data)

	if bs.offset >= thisLen {
		return 0, io.EOF
	}

	dstLen := len(dst)

	if bs.offset+dstLen >= thisLen {
		bytesRead := thisLen - bs.offset

		copy(dst, bs.data[bs.offset:])

		return bytesRead, nil
	}

	nextOffset := bs.offset + dstLen

	copy(dst, bs.data[bs.offset:nextOffset])

	bs.offset = nextOffset

	return dstLen, nil
}

func (bs *BacktrackingByteScanner) ReadByte() (byte, error) {
	if bs.offset >= len(bs.data) {
		return 0, io.EOF
	}

	b := bs.data[bs.offset]

	bs.offset++

	return b, nil
}

func (bs *BacktrackingByteScanner) ReadBytes(delim byte) ([]byte, error) {
	if bs.offset >= len(bs.data) {
		return nil, io.EOF
	}

	var result []byte

	index := bytes.IndexByte(bs.data[bs.offset:], delim)
	if index < 0 {
		copy(result, bs.data[bs.offset:])
		bs.offset = len(bs.data)

		return result, nil
	}

	nextOffset := bs.offset + index + 1
	if nextOffset >= len(bs.data) {
		copy(result, bs.data[bs.offset:])
		bs.offset = len(bs.data)
	} else {
		copy(result, bs.data[bs.offset:nextOffset])
		bs.offset = nextOffset
	}

	return result, nil
}

func (bs *BacktrackingByteScanner) SaveState() BacktrackingByteScannerScope {
	return BacktrackingByteScannerScope{offset: bs.offset}
}

func (bs *BacktrackingByteScanner) RestoreState(scope BacktrackingByteScannerScope) {
	bs.offset = scope.offset
}
