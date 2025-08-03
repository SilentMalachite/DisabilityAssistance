package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/zalando/go-keyring"
)

const (
	ServiceName   = "shien-system"
	KeyIdentifier = "encryption-key"
	KeySize       = 32 // AES-256 requires 32 bytes
)

// KeyManager handles secure storage and retrieval of encryption keys
type KeyManager interface {
	GetOrCreateKey() ([]byte, error)
	DeleteKey() error
}

// OSKeyManager implements KeyManager using OS-specific secure storage
type OSKeyManager struct {
	serviceName   string
	keyIdentifier string
}

// NewOSKeyManager creates a new OSKeyManager instance
func NewOSKeyManager() *OSKeyManager {
	return &OSKeyManager{
		serviceName:   ServiceName,
		keyIdentifier: KeyIdentifier,
	}
}

// GetOrCreateKey retrieves the encryption key from OS secure storage
// If no key exists, it generates a new one and stores it securely
func (km *OSKeyManager) GetOrCreateKey() ([]byte, error) {
	// Try to retrieve existing key from keyring
	keyBase64, err := keyring.Get(km.serviceName, km.keyIdentifier)
	if err == nil {
		// Key exists, decode and return
		key, err := decodeKey(keyBase64)
		if err != nil {
			return nil, fmt.Errorf("failed to decode existing key: %w", err)
		}
		if len(key) != KeySize {
			return nil, fmt.Errorf("invalid key size: expected %d, got %d", KeySize, len(key))
		}
		return key, nil
	}

	// Key doesn't exist or error occurred, generate new key
	if err != keyring.ErrNotFound {
		return nil, fmt.Errorf("failed to retrieve key from keyring: %w", err)
	}

	// Generate new 256-bit key
	key := make([]byte, KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	// Store key in keyring
	keyBase64 = encodeKey(key)
	if err := keyring.Set(km.serviceName, km.keyIdentifier, keyBase64); err != nil {
		return nil, fmt.Errorf("failed to store key in keyring: %w", err)
	}

	return key, nil
}

// DeleteKey removes the encryption key from OS secure storage
func (km *OSKeyManager) DeleteKey() error {
	err := keyring.Delete(km.serviceName, km.keyIdentifier)
	if err != nil && err != keyring.ErrNotFound {
		return fmt.Errorf("failed to delete key from keyring: %w", err)
	}
	return nil
}

// encodeKey converts byte array to base64 string for storage
func encodeKey(key []byte) string {
	// Using standard base64 URL encoding
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(key)
}

// decodeKey converts base64 string back to byte array
func decodeKey(keyBase64 string) ([]byte, error) {
	decoded, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 key: %w", err)
	}

	if len(decoded) != KeySize {
		return nil, fmt.Errorf("invalid decoded key size: expected %d, got %d", KeySize, len(decoded))
	}

	return decoded, nil
}
