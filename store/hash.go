package store

import (
	"crypto/sha256"
)

func hash(b []byte) []byte {
	hash := sha256.New()

	if _, err := hash.Write(b); err != nil {
		panic(err)
	}

	return hash.Sum(nil)
}
