package integrations

import (
	"crypto/sha256"
	"encoding/hex"
)

func ChecksumHex(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
