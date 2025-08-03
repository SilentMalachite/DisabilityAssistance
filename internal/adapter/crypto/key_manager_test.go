package crypto

import (
	"fmt"
	"testing"
)

// MockKeyManager is a test implementation of KeyManager
type MockKeyManager struct {
	key         []byte
	shouldError bool
	errorMsg    string
}

func NewMockKeyManager(key []byte) *MockKeyManager {
	return &MockKeyManager{key: key}
}

func (m *MockKeyManager) SetError(shouldError bool, errorMsg string) {
	m.shouldError = shouldError
	m.errorMsg = errorMsg
}

func (m *MockKeyManager) GetOrCreateKey() ([]byte, error) {
	if m.shouldError {
		return nil, fmt.Errorf("%s", m.errorMsg)
	}
	return m.key, nil
}

func (m *MockKeyManager) DeleteKey() error {
	if m.shouldError {
		return fmt.Errorf("%s", m.errorMsg)
	}
	return nil
}

func TestOSKeyManager_GetOrCreateKey(t *testing.T) {
	km := NewOSKeyManager()

	// Clean up any existing key before test
	_ = km.DeleteKey()

	// First call should create a new key
	key1, err := km.GetOrCreateKey()
	if err != nil {
		t.Fatalf("GetOrCreateKey() first call error = %v", err)
	}

	if len(key1) != KeySize {
		t.Errorf("Key size = %d, want %d", len(key1), KeySize)
	}

	// Second call should return the same key
	key2, err := km.GetOrCreateKey()
	if err != nil {
		t.Fatalf("GetOrCreateKey() second call error = %v", err)
	}

	if string(key1) != string(key2) {
		t.Error("GetOrCreateKey() should return the same key on subsequent calls")
	}

	// Clean up
	_ = km.DeleteKey()
}

func TestOSKeyManager_DeleteKey(t *testing.T) {
	km := NewOSKeyManager()

	// Create a key first
	_, err := km.GetOrCreateKey()
	if err != nil {
		t.Fatalf("GetOrCreateKey() error = %v", err)
	}

	// Delete the key
	err = km.DeleteKey()
	if err != nil {
		t.Fatalf("DeleteKey() error = %v", err)
	}

	// Deleting again should not error
	err = km.DeleteKey()
	if err != nil {
		t.Errorf("DeleteKey() on non-existent key should not error, got: %v", err)
	}
}

func TestEncodeDecodeKey(t *testing.T) {
	// Test with known key
	key := make([]byte, KeySize)
	for i := range key {
		key[i] = byte(i)
	}

	encoded := encodeKey(key)
	expectedLen := 43 // Expected length for 32 bytes without padding
	if len(encoded) != expectedLen {
		t.Errorf("Encoded key length = %d, want %d", len(encoded), expectedLen)
	}

	decoded, err := decodeKey(encoded)
	if err != nil {
		t.Fatalf("decodeKey() error = %v", err)
	}

	if string(key) != string(decoded) {
		t.Error("Encode/decode round trip failed")
	}
}

func TestDecodeKey_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "too short",
			input:   "short",
			wantErr: true,
		},
		{
			name:    "too long",
			input:   "this-is-way-too-long-for-a-valid-key-encoding",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			input:   "this-contains-invalid-chars-like-@#$%^&*()",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeKey(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewFieldCipherWithKeyManager(t *testing.T) {
	// Test with valid key
	validKey := make([]byte, KeySize)
	for i := range validKey {
		validKey[i] = byte(i)
	}

	mockKM := NewMockKeyManager(validKey)
	cipher, err := NewFieldCipherWithKeyManager(mockKM)
	if err != nil {
		t.Fatalf("NewFieldCipherWithKeyManager() error = %v", err)
	}

	if cipher == nil {
		t.Error("NewFieldCipherWithKeyManager() returned nil cipher")
	}

	// Test encryption/decryption works
	plaintext := "テストデータ"
	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := cipher.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Round trip failed: got %q, want %q", decrypted, plaintext)
	}
}

func TestNewFieldCipherWithKeyManager_KeyManagerError(t *testing.T) {
	mockKM := NewMockKeyManager(nil)
	mockKM.SetError(true, "key manager error")

	_, err := NewFieldCipherWithKeyManager(mockKM)
	if err == nil {
		t.Error("NewFieldCipherWithKeyManager() should fail when KeyManager returns error")
	}
}

func TestNewFieldCipher_Integration(t *testing.T) {
	// Clean up any existing key
	km := NewOSKeyManager()
	_ = km.DeleteKey()

	// Create cipher using OS key management
	cipher, err := NewFieldCipher()
	if err != nil {
		t.Fatalf("NewFieldCipher() error = %v", err)
	}

	// Test basic functionality
	plaintext := "統合テスト用データ"
	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	decrypted, err := cipher.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Integration test round trip failed: got %q, want %q", decrypted, plaintext)
	}

	// Create another cipher instance - should use same key
	cipher2, err := NewFieldCipher()
	if err != nil {
		t.Fatalf("NewFieldCipher() second instance error = %v", err)
	}

	// Should be able to decrypt data encrypted by first instance
	decrypted2, err := cipher2.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() with second cipher error = %v", err)
	}

	if decrypted2 != plaintext {
		t.Errorf("Cross-instance decryption failed: got %q, want %q", decrypted2, plaintext)
	}

	// Clean up
	_ = km.DeleteKey()
}

func BenchmarkOSKeyManager_GetOrCreateKey(b *testing.B) {
	km := NewOSKeyManager()

	// Create key once
	_, err := km.GetOrCreateKey()
	if err != nil {
		b.Fatalf("Setup error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := km.GetOrCreateKey()
		if err != nil {
			b.Fatalf("GetOrCreateKey() error = %v", err)
		}
	}

	// Clean up
	_ = km.DeleteKey()
}

func BenchmarkEncodeDecodeKey(b *testing.B) {
	key := make([]byte, KeySize)
	for i := range key {
		key[i] = byte(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		encoded := encodeKey(key)
		_, err := decodeKey(encoded)
		if err != nil {
			b.Fatalf("Decode error: %v", err)
		}
	}
}
