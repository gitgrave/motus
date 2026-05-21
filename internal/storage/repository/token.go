package repository

import (
	"crypto/sha256"
	"encoding/hex"
)

// hashToken returns the lowercase hex SHA-256 of the raw token string.
// Use this whenever storing or looking up tokens to avoid plaintext storage.
func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
