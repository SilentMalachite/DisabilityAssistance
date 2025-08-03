package crypto

import (
	"crypto/rand"
	"testing"
)

func TestNewFieldCipherWithKey(t *testing.T) {
	tests := []struct {
		name    string
		keySize int
		wantErr bool
	}{
		{
			name:    "valid AES-128 key",
			keySize: 16,
			wantErr: false,
		},
		{
			name:    "valid AES-192 key",
			keySize: 24,
			wantErr: false,
		},
		{
			name:    "valid AES-256 key",
			keySize: 32,
			wantErr: false,
		},
		{
			name:    "invalid key size",
			keySize: 15,
			wantErr: true,
		},
		{
			name:    "empty key",
			keySize: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keySize)
			if tt.keySize > 0 {
				_, err := rand.Read(key)
				if err != nil {
					t.Fatalf("failed to generate test key: %v", err)
				}
			}

			cipher, err := NewFieldCipherWithKey(key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFieldCipherWithKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && cipher == nil {
				t.Error("NewFieldCipherWithKey() returned nil cipher without error")
			}
		})
	}
}

func TestFieldCipher_EncryptDecrypt_RoundTrip(t *testing.T) {
	key := make([]byte, 32) // AES-256
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	cipher, err := NewFieldCipherWithKey(key)
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "empty string",
			plaintext: "",
		},
		{
			name:      "short text",
			plaintext: "test",
		},
		{
			name:      "japanese text",
			plaintext: "山田太郎",
		},
		{
			name:      "long text",
			plaintext: "This is a very long text that should be encrypted and decrypted properly without any issues or data corruption.",
		},
		{
			name:      "special characters",
			plaintext: "!@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
		{
			name:      "unicode characters",
			plaintext: "こんにちは世界🌍🎌",
		},
		{
			name:      "newlines and tabs",
			plaintext: "line1\nline2\tcolumn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, err := cipher.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			if tt.plaintext == "" && ciphertext != nil {
				t.Error("Encrypt() should return nil for empty string")
			}

			decrypted, err := cipher.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("Round trip failed: got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestFieldCipher_Encrypt_Randomness(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	cipher, err := NewFieldCipherWithKey(key)
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}

	plaintext := "same plaintext"

	ciphertext1, err := cipher.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("first Encrypt() error = %v", err)
	}

	ciphertext2, err := cipher.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("second Encrypt() error = %v", err)
	}

	if string(ciphertext1) == string(ciphertext2) {
		t.Error("Encrypt() should produce different ciphertexts for same plaintext (nonce should be random)")
	}

	// Both should decrypt to the same plaintext
	decrypted1, err := cipher.Decrypt(ciphertext1)
	if err != nil {
		t.Fatalf("first Decrypt() error = %v", err)
	}

	decrypted2, err := cipher.Decrypt(ciphertext2)
	if err != nil {
		t.Fatalf("second Decrypt() error = %v", err)
	}

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("Both ciphertexts should decrypt to original plaintext")
	}
}

func TestFieldCipher_Decrypt_ErrorCases(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	cipher, err := NewFieldCipherWithKey(key)
	if err != nil {
		t.Fatalf("failed to create cipher: %v", err)
	}

	tests := []struct {
		name       string
		ciphertext []byte
		wantErr    bool
		wantResult string
	}{
		{
			name:       "nil ciphertext",
			ciphertext: nil,
			wantErr:    false,
			wantResult: "",
		},
		{
			name:       "empty ciphertext",
			ciphertext: []byte{},
			wantErr:    false,
			wantResult: "",
		},
		{
			name:       "too short ciphertext",
			ciphertext: []byte{1, 2, 3},
			wantErr:    true,
		},
		{
			name:       "invalid ciphertext",
			ciphertext: make([]byte, 20), // Should be longer than nonce size but invalid
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cipher.Decrypt(tt.ciphertext)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.wantResult {
				t.Errorf("Decrypt() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestFieldCipher_CrossKeyDecryption(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)

	_, err := rand.Read(key1)
	if err != nil {
		t.Fatalf("failed to generate first test key: %v", err)
	}

	_, err = rand.Read(key2)
	if err != nil {
		t.Fatalf("failed to generate second test key: %v", err)
	}

	cipher1, err := NewFieldCipherWithKey(key1)
	if err != nil {
		t.Fatalf("failed to create first cipher: %v", err)
	}

	cipher2, err := NewFieldCipherWithKey(key2)
	if err != nil {
		t.Fatalf("failed to create second cipher: %v", err)
	}

	plaintext := "sensitive data"

	ciphertext, err := cipher1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Try to decrypt with wrong key - should fail
	_, err = cipher2.Decrypt(ciphertext)
	if err == nil {
		t.Error("Decrypt() with wrong key should fail")
	}
}

func BenchmarkFieldCipher_Encrypt(b *testing.B) {
	key := make([]byte, 32)
	rand.Read(key)

	cipher, err := NewFieldCipherWithKey(key)
	if err != nil {
		b.Fatalf("failed to create cipher: %v", err)
	}

	plaintext := "山田太郎"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cipher.Encrypt(plaintext)
		if err != nil {
			b.Fatalf("Encrypt() error = %v", err)
		}
	}
}

func BenchmarkFieldCipher_Decrypt(b *testing.B) {
	key := make([]byte, 32)
	rand.Read(key)

	cipher, err := NewFieldCipherWithKey(key)
	if err != nil {
		b.Fatalf("failed to create cipher: %v", err)
	}

	plaintext := "山田太郎"
	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		b.Fatalf("failed to encrypt test data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cipher.Decrypt(ciphertext)
		if err != nil {
			b.Fatalf("Decrypt() error = %v", err)
		}
	}
}
