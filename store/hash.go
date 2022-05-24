package store

import (
	"crypto/sha256"
	"encoding/hex"
)

func hash(b []byte) []byte {
	hash := sha256.New()

	if _, err := hash.Write(b); err != nil {
		panic(err)
	}

	return hash.Sum(nil)
}

func hashString(s string) string {
	return hex.EncodeToString(hash([]byte(s)))
}
