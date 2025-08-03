package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

type FieldCipher struct {
	gcm        cipher.AEAD
	keyManager KeyManager
}

// NewFieldCipher creates a new FieldCipher using secure key management
func NewFieldCipher() (*FieldCipher, error) {
	keyManager := NewOSKeyManager()
	return NewFieldCipherWithKeyManager(keyManager)
}

// NewFieldCipherWithKeyManager creates a FieldCipher with a custom KeyManager
func NewFieldCipherWithKeyManager(keyManager KeyManager) (*FieldCipher, error) {
	key, err := keyManager.GetOrCreateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	return &FieldCipher{
		gcm:        gcm,
		keyManager: keyManager,
	}, nil
}

// NewFieldCipherWithKey creates a FieldCipher with a provided key (for testing)
func NewFieldCipherWithKey(key []byte) (*FieldCipher, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	return &FieldCipher{gcm: gcm}, nil
}

func (c *FieldCipher) Encrypt(plaintext string) ([]byte, error) {
	if plaintext == "" {
		return nil, nil
	}

	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("generating nonce: %w", err)
	}

	ciphertext := c.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertext, nil
}

func (c *FieldCipher) Decrypt(ciphertext []byte) (string, error) {
	if len(ciphertext) == 0 {
		return "", nil
	}

	nonceSize := c.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := c.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	// Convert to string and clear the plaintext bytes
	result := string(plaintext)
	ClearBytes(plaintext)

	return result, nil
}

// DecryptSecure decrypts data and returns a SecureString that must be cleared after use
func (c *FieldCipher) DecryptSecure(ciphertext []byte) (*SecureString, error) {
	if len(ciphertext) == 0 {
		return NewSecureString(""), nil
	}

	plaintext, err := c.Decrypt(ciphertext)
	if err != nil {
		return nil, err
	}

	return NewSecureString(plaintext), nil
}
