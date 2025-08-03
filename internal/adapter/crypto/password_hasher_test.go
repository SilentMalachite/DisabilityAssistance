package crypto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBcryptPasswordHasher_HashPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	password := "testPassword123!"
	hash, err := hasher.HashPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
	assert.True(t, strings.HasPrefix(hash, "$2a$"), "Hash should start with bcrypt prefix")
	assert.True(t, len(hash) > 50, "Hash should be reasonably long")
}

func TestBcryptPasswordHasher_HashPassword_EmptyPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	hash, err := hasher.HashPassword("")

	assert.Error(t, err)
	assert.Empty(t, hash)
	assert.Equal(t, ErrPasswordRequired, err)
}

func TestBcryptPasswordHasher_HashPassword_WeakPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	weakPasswords := []string{
		"123",      // Too short
		"password", // Common password
		"abc",      // Too short and simple
		"1234567",  // Only numbers, too short
	}

	for _, weakPassword := range weakPasswords {
		t.Run("weak_password_"+weakPassword, func(t *testing.T) {
			hash, err := hasher.HashPassword(weakPassword)

			assert.Error(t, err)
			assert.Empty(t, hash)
			assert.Equal(t, ErrWeakPassword, err)
		})
	}
}

func TestBcryptPasswordHasher_HashPassword_DifferentHashes(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	password := "testPassword123!"

	// Hash the same password multiple times
	hash1, err1 := hasher.HashPassword(password)
	require.NoError(t, err1)

	hash2, err2 := hasher.HashPassword(password)
	require.NoError(t, err2)

	// Even though the password is the same, hashes should be different due to salt
	assert.NotEqual(t, hash1, hash2)
}

func TestBcryptPasswordHasher_CheckPassword_ValidPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	password := "testPassword123!"
	hash, err := hasher.HashPassword(password)
	require.NoError(t, err)

	// Check with correct password
	err = hasher.CheckPassword(hash, password)
	assert.NoError(t, err)
}

func TestBcryptPasswordHasher_CheckPassword_InvalidPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	password := "testPassword123!"
	wrongPassword := "wrongPassword456!"

	hash, err := hasher.HashPassword(password)
	require.NoError(t, err)

	// Check with wrong password
	err = hasher.CheckPassword(hash, wrongPassword)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidPassword, err)
}

func TestBcryptPasswordHasher_CheckPassword_EmptyInputs(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	tests := []struct {
		name     string
		hash     string
		password string
		wantErr  error
	}{
		{
			name:     "empty_hash",
			hash:     "",
			password: "password",
			wantErr:  ErrInvalidPassword,
		},
		{
			name:     "empty_password",
			hash:     "$2a$12$hash",
			password: "",
			wantErr:  ErrInvalidPassword,
		},
		{
			name:     "both_empty",
			hash:     "",
			password: "",
			wantErr:  ErrInvalidPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hasher.CheckPassword(tt.hash, tt.password)
			assert.Error(t, err)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestBcryptPasswordHasher_CheckPassword_MalformedHash(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	malformedHashes := []string{
		"not-a-hash",
		"$2a$invalid",
		"plaintext",
		"$1$oldhash$format",
	}

	for _, malformedHash := range malformedHashes {
		t.Run("malformed_"+malformedHash, func(t *testing.T) {
			err := hasher.CheckPassword(malformedHash, "password")
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidPassword, err)
		})
	}
}

func TestBcryptPasswordHasher_CheckPassword_RoundTrip(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	testPasswords := []string{
		"simplePassword123",
		"ComplexP@ssw0rd!",
		"password with spaces",
		"日本語パスワード123", // Japanese characters
		"emoji🔐password",
		"mediumLengthPassword123!", // Within bcrypt's 72-byte limit
	}

	for _, password := range testPasswords {
		t.Run("password_"+password[:min(10, len(password))], func(t *testing.T) {
			// Hash the password
			hash, err := hasher.HashPassword(password)
			require.NoError(t, err)
			require.NotEmpty(t, hash)

			// Verify the password
			err = hasher.CheckPassword(hash, password)
			assert.NoError(t, err)

			// Verify wrong password fails
			err = hasher.CheckPassword(hash, password+"wrong")
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidPassword, err)
		})
	}
}

func TestBcryptPasswordHasher_PasswordStrengthValidation(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	strongPasswords := []string{
		"StrongPassword123!",
		"MySecureP@ss2023",
		"C0mpl3x!P@ssw0rd",
		"VeryLongAndComplexPassword123!@#",
	}

	for _, password := range strongPasswords {
		t.Run("strong_password", func(t *testing.T) {
			hash, err := hasher.HashPassword(password)
			assert.NoError(t, err)
			assert.NotEmpty(t, hash)
		})
	}
}

func TestBcryptPasswordHasher_TooLongPassword(t *testing.T) {
	hasher := NewBcryptPasswordHasher()

	// Create a password that exceeds bcrypt's 72-byte limit
	tooLongPassword := strings.Repeat("a", 73)

	hash, err := hasher.HashPassword(tooLongPassword)

	assert.Error(t, err)
	assert.Empty(t, hash)
	assert.Equal(t, ErrWeakPassword, err)
}

func TestBcryptPasswordHasher_Interface(t *testing.T) {
	// Ensure BcryptPasswordHasher implements PasswordHasher interface
	var _ PasswordHasher = (*BcryptPasswordHasher)(nil)
}

// Helper function for minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
