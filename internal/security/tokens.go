package security

import (
	"crypto/rand"
	"encoding/base64"
)

// TokenLength is the byte length of raw tokens (before encoding)
const TokenLength = 32

// GenerateSecureToken creates a cryptographically random URL-safe token.
// Returns a base64 URL-encoded string with 256 bits (32 bytes) of entropy.
func GenerateSecureToken() (string, error) {
	b := make([]byte, TokenLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
