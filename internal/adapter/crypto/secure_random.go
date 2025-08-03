package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
)

// SecureRandomGenerator provides cryptographically secure random number generation
type SecureRandomGenerator struct{}

// NewSecureRandomGenerator creates a new secure random generator
func NewSecureRandomGenerator() *SecureRandomGenerator {
	return &SecureRandomGenerator{}
}

// GenerateSessionID generates a cryptographically secure session ID
func (g *SecureRandomGenerator) GenerateSessionID(length int) (string, error) {
	if length < 16 {
		return "", fmt.Errorf("session ID length must be at least 16 bytes")
	}
	if length > 128 {
		return "", fmt.Errorf("session ID length must not exceed 128 bytes")
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Base64 URLエンコーディングでセッションIDを生成
	sessionID := base64.URLEncoding.EncodeToString(bytes)
	return sessionID, nil
}

// GenerateCSRFToken generates a cryptographically secure CSRF token
func (g *SecureRandomGenerator) GenerateCSRFToken(length int) (string, error) {
	if length < 16 {
		return "", fmt.Errorf("CSRF token length must be at least 16 bytes")
	}
	if length > 128 {
		return "", fmt.Errorf("CSRF token length must not exceed 128 bytes")
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Base64 URLエンコーディングでCSRFトークンを生成
	csrfToken := base64.URLEncoding.EncodeToString(bytes)
	return csrfToken, nil
}

// GenerateRandomBytes generates cryptographically secure random bytes
func (g *SecureRandomGenerator) GenerateRandomBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, fmt.Errorf("length must be positive")
	}
	if length > 1024 {
		return nil, fmt.Errorf("length must not exceed 1024 bytes")
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return bytes, nil
}

// ConstantTimeCompare performs constant-time comparison of two strings
// to prevent timing attacks
func (g *SecureRandomGenerator) ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// ValidateSessionID validates the format and security of a session ID
func (g *SecureRandomGenerator) ValidateSessionID(sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	// Base64 URLデコードして検証
	decoded, err := base64.URLEncoding.DecodeString(sessionID)
	if err != nil {
		return fmt.Errorf("invalid session ID format: %w", err)
	}

	if len(decoded) < 16 {
		return fmt.Errorf("session ID is too short")
	}

	if len(decoded) > 128 {
		return fmt.Errorf("session ID is too long")
	}

	return nil
}

// ValidateCSRFToken validates the format and security of a CSRF token
func (g *SecureRandomGenerator) ValidateCSRFToken(csrfToken string) error {
	if csrfToken == "" {
		return fmt.Errorf("CSRF token cannot be empty")
	}

	// Base64 URLデコードして検証
	decoded, err := base64.URLEncoding.DecodeString(csrfToken)
	if err != nil {
		return fmt.Errorf("invalid CSRF token format: %w", err)
	}

	if len(decoded) < 16 {
		return fmt.Errorf("CSRF token is too short")
	}

	if len(decoded) > 128 {
		return fmt.Errorf("CSRF token is too long")
	}

	return nil
}

// SecureWipe overwrites sensitive data in memory
func (g *SecureRandomGenerator) SecureWipe(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

// SecureWipeString overwrites sensitive string data in memory
// Note: This is best effort as Go strings are immutable
func (g *SecureRandomGenerator) SecureWipeString(data *string) {
	if data != nil {
		*data = ""
	}
}
