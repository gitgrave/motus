package config

import (
	"encoding/hex"
	"fmt"
)

// ParseCSRFSecret decodes a hex-encoded 32-byte CSRF secret.
// Returns an error on invalid hex or wrong length.
func ParseCSRFSecret(hexSecret string) ([]byte, error) {
	secret, err := hex.DecodeString(hexSecret)
	if err != nil {
		return nil, fmt.Errorf("MOTUS_CSRF_SECRET is not valid hex: %w", err)
	}
	if len(secret) != 32 {
		return nil, fmt.Errorf("MOTUS_CSRF_SECRET must be exactly 32 bytes (64 hex chars), got %d bytes", len(secret))
	}
	return secret, nil
}
