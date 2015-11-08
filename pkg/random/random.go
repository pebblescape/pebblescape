package random

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
)

// String returns a random hex-formatted string of specified length.
func String(n int) string {
	return Hex(n/2 + 1)[:n]
}

// Hex returns a random hex-formatted string of specified length.
func Hex(bytes int) string {
	return hex.EncodeToString(Bytes(bytes))
}

// Base64 returns a random base64-formatted string of specified length.
func Base64(bytes int) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(Bytes(bytes)), "=")
}

// Bytes returns a specified length of random bytes.
func Bytes(n int) []byte {
	data := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, data)
	if err != nil {
		panic(err)
	}
	return data
}

// UUID returns a new v4 UUID string.
func UUID() string {
	id := Bytes(16)
	id[6] &= 0x0F // clear version
	id[6] |= 0x40 // set version to 4 (random uuid)
	id[8] &= 0x3F // clear variant
	id[8] |= 0x80 // set to IETF variant
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:])
}
