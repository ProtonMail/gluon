package store

import (
	"bytes"
	"compress/zlib"
)

type ZLibCompressor struct{}

func (ZLibCompressor) Compress(dec []byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	zw := zlib.NewWriter(buf)

	if _, err := zw.Write(dec); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (ZLibCompressor) Decompress(cmp []byte) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewReader(cmp))
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	if _, err := buf.ReadFrom(zr); err != nil {
		return nil, err
	}

	if err := zr.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
