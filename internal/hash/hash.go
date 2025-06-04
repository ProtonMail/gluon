package hash

import (
	"crypto/sha256"
	"encoding/hex"
)

func SHA256(key []byte) []byte {
	hash := sha256.Sum256(key)

	return hash[:]
}

func HashToString(entry string) string {
	return hex.EncodeToString(SHA256([]byte(entry)))
}
