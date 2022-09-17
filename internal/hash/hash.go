package hash

import "crypto/sha256"

func SHA256(key []byte) []byte {
	hash := sha256.Sum256(key)

	return hash[:]
}
