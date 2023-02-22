package store

import (
	"crypto/cipher"
	"io"
)

// Fallback provides an interface to supply an alternative way to read a store file should the main route fail.
// This is mainly intended to allow users of the library to read old store formats they may have kept on disk.
// This is a stop-gap until a complete data migration cycle can be implemented in gluon.
type Fallback interface {
	Read(gcm cipher.AEAD, reader io.Reader) ([]byte, error)

	Write(gcm cipher.AEAD, filepath string, data []byte) error
}
