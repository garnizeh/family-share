package security_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"familyshare/internal/security"
)

func TestGenerateSecureToken_Length(t *testing.T) {
	token, err := security.GenerateSecureToken()
	if err != nil {
		t.Fatalf("GenerateSecureToken failed: %v", err)
	}

	// Decode to verify raw token length
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("token is not valid base64: %v", err)
	}

	if len(decoded) != security.TokenLength {
		t.Fatalf("expected token length %d bytes, got %d", security.TokenLength, len(decoded))
	}

	// Encoded token should be longer due to base64 encoding
	// 32 bytes -> 44 chars in base64 (with padding) or 43 without
	if len(token) < 40 {
		t.Fatalf("encoded token seems too short: %d chars", len(token))
	}
}

func TestGenerateSecureToken_URLSafe(t *testing.T) {
	token, err := security.GenerateSecureToken()
	if err != nil {
		t.Fatalf("GenerateSecureToken failed: %v", err)
	}

	// URL-safe base64 should not contain + or /
	// It may contain - and _ instead
	if strings.Contains(token, "+") {
		t.Fatalf("token contains '+', not URL-safe: %s", token)
	}
	if strings.Contains(token, "/") {
		t.Fatalf("token contains '/', not URL-safe: %s", token)
	}

	// Should only contain URL-safe characters: A-Z, a-z, 0-9, -, _, =
	for _, char := range token {
		if !isURLSafeBase64Char(char) {
			t.Fatalf("token contains invalid character '%c': %s", char, token)
		}
	}
}

func TestGenerateSecureToken_Uniqueness(t *testing.T) {
	// Generate 1000 tokens and verify they're all unique
	tokens := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		token, err := security.GenerateSecureToken()
		if err != nil {
			t.Fatalf("GenerateSecureToken failed on iteration %d: %v", i, err)
		}

		if tokens[token] {
			t.Fatalf("duplicate token found: %s", token)
		}
		tokens[token] = true
	}

	if len(tokens) != iterations {
		t.Fatalf("expected %d unique tokens, got %d", iterations, len(tokens))
	}
}

func TestGenerateSecureToken_HighEntropy(t *testing.T) {
	// Basic entropy check: ensure tokens have varied characters
	token, err := security.GenerateSecureToken()
	if err != nil {
		t.Fatalf("GenerateSecureToken failed: %v", err)
	}

	// Count unique characters
	uniqueChars := make(map[rune]bool)
	for _, char := range token {
		uniqueChars[char] = true
	}

	// A truly random 32-byte token should have many unique characters
	// Even with base64 encoding (64 possible chars), we expect decent variety
	minUniqueChars := 10
	if len(uniqueChars) < minUniqueChars {
		t.Fatalf("token has only %d unique characters, expected at least %d (possible low entropy): %s",
			len(uniqueChars), minUniqueChars, token)
	}

	// Verify token is not all the same character (trivial check)
	if len(uniqueChars) == 1 {
		t.Fatalf("token contains only one character, no entropy: %s", token)
	}
}

func TestGenerateSecureToken_Randomness(t *testing.T) {
	// Generate two tokens and ensure they're different
	token1, err := security.GenerateSecureToken()
	if err != nil {
		t.Fatalf("GenerateSecureToken failed: %v", err)
	}

	token2, err := security.GenerateSecureToken()
	if err != nil {
		t.Fatalf("GenerateSecureToken failed: %v", err)
	}

	if token1 == token2 {
		t.Fatalf("consecutive tokens are identical (extremely unlikely with crypto/rand): %s", token1)
	}
}

func TestGenerateSecureToken_NoDeterministicPattern(t *testing.T) {
	// Verify tokens don't follow a predictable pattern
	tokens := make([]string, 10)

	for i := range tokens {
		token, err := security.GenerateSecureToken()
		if err != nil {
			t.Fatalf("GenerateSecureToken failed: %v", err)
		}
		tokens[i] = token
	}

	// Check that tokens don't have identical prefixes or suffixes
	// This is a basic sanity check
	for i := 0; i < len(tokens)-1; i++ {
		for j := i + 1; j < len(tokens); j++ {
			// Check first 10 chars aren't identical
			if len(tokens[i]) >= 10 && len(tokens[j]) >= 10 {
				if tokens[i][:10] == tokens[j][:10] {
					t.Fatalf("tokens %d and %d have identical 10-char prefix (suspicious): %s vs %s",
						i, j, tokens[i], tokens[j])
				}
			}

			// Check last 10 chars aren't identical
			if len(tokens[i]) >= 10 && len(tokens[j]) >= 10 {
				suffix1 := tokens[i][len(tokens[i])-10:]
				suffix2 := tokens[j][len(tokens[j])-10:]
				if suffix1 == suffix2 {
					t.Fatalf("tokens %d and %d have identical 10-char suffix (suspicious): %s vs %s",
						i, j, tokens[i], tokens[j])
				}
			}
		}
	}
}

func TestGenerateSecureToken_DecodesCorrectly(t *testing.T) {
	token, err := security.GenerateSecureToken()
	if err != nil {
		t.Fatalf("GenerateSecureToken failed: %v", err)
	}

	// Verify token can be decoded
	decoded, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("token cannot be decoded as base64 URL encoding: %v", err)
	}

	// Verify decoded length
	if len(decoded) != security.TokenLength {
		t.Fatalf("decoded token has incorrect length: expected %d, got %d", security.TokenLength, len(decoded))
	}

	// Re-encode and verify it matches (base64 is deterministic)
	reencoded := base64.URLEncoding.EncodeToString(decoded)
	if token != reencoded {
		t.Fatalf("re-encoded token doesn't match original: %s vs %s", token, reencoded)
	}
}

func TestGenerateSecureToken_ConsistentLength(t *testing.T) {
	// Verify all generated tokens have the same encoded length
	var expectedLength int
	iterations := 100

	for i := 0; i < iterations; i++ {
		token, err := security.GenerateSecureToken()
		if err != nil {
			t.Fatalf("GenerateSecureToken failed: %v", err)
		}

		if i == 0 {
			expectedLength = len(token)
		} else if len(token) != expectedLength {
			t.Fatalf("token length inconsistent: expected %d, got %d for token: %s",
				expectedLength, len(token), token)
		}
	}
}

// isURLSafeBase64Char checks if a character is valid in URL-safe base64
func isURLSafeBase64Char(char rune) bool {
	return (char >= 'A' && char <= 'Z') ||
		(char >= 'a' && char <= 'z') ||
		(char >= '0' && char <= '9') ||
		char == '-' ||
		char == '_' ||
		char == '='
}
