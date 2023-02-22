package fallback_v0

import (
	"bytes"
	"compress/gzip"
)

type GZipCompressor struct{}

func (GZipCompressor) Compress(dec []byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	zw := gzip.NewWriter(buf)

	if _, err := zw.Write(dec); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (GZipCompressor) Decompress(cmp []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(cmp))
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
