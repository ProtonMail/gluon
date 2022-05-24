package parser

import (
	"testing"
)

func TestParserDoesntCrashWithInvalidByteSequence(t *testing.T) {
	// This sequence of bytes taken from an SSL handshake would cause the C++ parser code to crash.
	// While the crash has been addressed, we leave the test both on the parser side and go for future checks.
	bytes := []uint8{
		22, 3, 1, 1, 55, 1, 0, 1, 51, 3, 3, 197, 18, 92, 146, 206, 72, 40, 181, 29, 204, 229, 121, 102,
		109, 81, 40, 172, 107, 48, 135, 230, 173, 107, 115, 13, 165, 209, 62, 110, 57, 91, 172, 32, 177, 238, 14,
		114, 255, 237, 154, 71, 205, 130, 245, 131, 54, 61, 214, 67, 38, 91, 118, 207, 164, 187, 77, 58, 55, 68,
		245, 59, 166, 194, 7, 30, 0, 62, 19, 2, 19, 3, 19, 1, 192, 44, 192, 48, 0, 159, 204, 169, 204, 168, 204, 170,
		192, 43, 192, 47, 0, 158, 192, 36, 192, 40, 0, 107, 192, 35, 192, 39, 0, 103, 192, 10,
	}
	text := string(bytes)
	Parse(text, NewStringMap())
}
