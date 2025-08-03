package crypto

import (
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// BcryptPasswordHasher implements PasswordHasher interface using bcrypt
type BcryptPasswordHasher struct {
	cost int
}

// PasswordHasher defines interface for password hashing operations
type PasswordHasher interface {
	// HashPassword creates a secure hash from a password
	HashPassword(password string) (string, error)

	// CheckPassword verifies a password against its hash
	CheckPassword(hashedPassword, password string) error
}

// NewBcryptPasswordHasher creates a new bcrypt password hasher
func NewBcryptPasswordHasher() PasswordHasher {
	return &BcryptPasswordHasher{
		cost: bcrypt.DefaultCost, // Cost of 10 is recommended for production
	}
}

// NewBcryptPasswordHasherWithCost creates a new bcrypt password hasher with custom cost
func NewBcryptPasswordHasherWithCost(cost int) PasswordHasher {
	return &BcryptPasswordHasher{
		cost: cost,
	}
}

// HashPassword creates a secure hash from a password
func (h *BcryptPasswordHasher) HashPassword(password string) (string, error) {
	// Validate password
	if err := h.validatePassword(password); err != nil {
		return "", err
	}

	// Check bcrypt length limit (72 bytes)
	if len([]byte(password)) > 72 {
		return "", ErrWeakPassword
	}

	// Hash password using bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// CheckPassword verifies a password against its hash
func (h *BcryptPasswordHasher) CheckPassword(hashedPassword, password string) error {
	// Validate inputs
	if hashedPassword == "" || password == "" {
		return ErrInvalidPassword
	}

	// Compare password with hash
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		// bcrypt.ErrMismatchedHashAndPassword or other bcrypt errors
		return ErrInvalidPassword
	}

	return nil
}

// validatePassword checks if password meets security requirements
func (h *BcryptPasswordHasher) validatePassword(password string) error {
	if password == "" {
		return ErrPasswordRequired
	}

	// Check for weak passwords
	if h.isWeakPassword(password) {
		return ErrWeakPassword
	}

	return nil
}

// isWeakPassword checks if password is considered weak
func (h *BcryptPasswordHasher) isWeakPassword(password string) bool {
	// Minimum length requirement
	if len(password) < 8 {
		return true
	}

	// Check for common weak passwords
	commonPasswords := []string{
		"password", "123456", "123456789", "qwerty", "abc123",
		"password123", "admin", "letmein", "welcome", "monkey",
		"dragon", "pass", "master", "hello", "freedom",
	}

	lowerPassword := strings.ToLower(password)
	for _, common := range commonPasswords {
		if lowerPassword == common {
			return true
		}
	}

	// More flexible character type checking
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `\s]`).MatchString(password) // Include spaces
	hasUnicode := regexp.MustCompile(`[^\x00-\x7F]`).MatchString(password)                                        // Non-ASCII characters (Unicode)

	// Password should have at least 2 character types or contain Unicode characters
	characterTypes := 0
	if hasDigit {
		characterTypes++
	}
	if hasLower {
		characterTypes++
	}
	if hasUpper {
		characterTypes++
	}
	if hasSpecial {
		characterTypes++
	}
	if hasUnicode {
		characterTypes += 2 // Unicode characters count as 2 types for flexibility
	}

	// Be more lenient with longer passwords
	minTypes := 2
	if len(password) >= 12 {
		minTypes = 1 // Very long passwords can be more flexible
	}

	if characterTypes < minTypes {
		return true
	}

	// Check for simple patterns
	if isSimplePattern(password) {
		return true
	}

	return false
}

// isSimplePattern checks for simple patterns like "12345", "abcde", etc.
func isSimplePattern(password string) bool {
	// Check for sequential numbers
	if matched, _ := regexp.MatchString(`^[0-9]+$`, password); matched {
		if len(password) <= 8 {
			return true
		}
	}

	// Check for sequential letters
	if matched, _ := regexp.MatchString(`^[a-zA-Z]+$`, password); matched {
		if len(password) <= 6 {
			return true
		}
	}

	// Check for repeated characters (like "aaaa" or "1111")
	if len(password) <= 8 {
		firstChar := password[0]
		isRepeated := true
		for _, char := range password {
			if char != rune(firstChar) {
				isRepeated = false
				break
			}
		}
		if isRepeated {
			return true
		}
	}

	return false
}
