package validation

import (
	"testing"
)

func TestValidator_ValidateRequired(t *testing.T) {
	v := NewValidator()
	
	tests := []struct {
		name     string
		field    string
		value    string
		hasError bool
	}{
		{"empty string", "test", "", true},
		{"whitespace only", "test", "   ", true},
		{"valid value", "test", "value", false},
		{"value with spaces", "test", " value ", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRequired(tt.field, tt.value)
			if (err != nil) != tt.hasError {
				t.Errorf("ValidateRequired() error = %v, hasError %v", err, tt.hasError)
			}
		})
	}
}

func TestValidator_ValidateEmail(t *testing.T) {
	v := NewValidator()
	
	tests := []struct {
		name     string
		email    string
		hasError bool
	}{
		{"valid email", "test@example.com", false},
		{"empty email", "", false}, // Empty is allowed
		{"invalid email", "invalid-email", true},
		{"email without domain", "test@", true},
		{"email without @", "testexample.com", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateEmail("email", tt.email)
			if (err != nil) != tt.hasError {
				t.Errorf("ValidateEmail() error = %v, hasError %v", err, tt.hasError)
			}
		})
	}
}

func TestValidator_ValidatePhoneNumber(t *testing.T) {
	v := NewValidator()
	
	tests := []struct {
		name     string
		phone    string
		hasError bool
	}{
		{"valid phone with hyphens", "03-1234-5678", false},
		{"valid phone without hyphens", "0312345678", false},
		{"valid mobile", "090-1234-5678", false},
		{"empty phone", "", false}, // Empty is allowed
		{"invalid phone", "123", true},
		{"phone with letters", "03-abcd-5678", true},
		{"phone with invalid format", "3-1234-5678", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidatePhoneNumber("phone", tt.phone)
			if (err != nil) != tt.hasError {
				t.Errorf("ValidatePhoneNumber() error = %v, hasError %v", err, tt.hasError)
			}
		})
	}
}

func TestValidator_ValidateDate(t *testing.T) {
	v := NewValidator()
	
	tests := []struct {
		name     string
		date     string
		hasError bool
	}{
		{"valid date", "2023-12-31", false},
		{"empty date", "", false}, // Empty is allowed
		{"invalid format", "31-12-2023", true},
		{"invalid date", "2023-13-01", true},
		{"invalid day", "2023-02-30", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateDate("date", tt.date)
			if (err != nil) != tt.hasError {
				t.Errorf("ValidateDate() error = %v, hasError %v", err, tt.hasError)
			}
		})
	}
}

func TestValidator_ValidateJapaneseName(t *testing.T) {
	v := NewValidator()
	
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"hiragana name", "たなか", false},
		{"katakana name", "タナカ", false},
		{"kanji name", "田中", false},
		{"mixed japanese", "田中 太郎", false},
		{"empty name", "", false}, // Empty is allowed
		{"english name", "John", true},
		{"mixed with english", "田中John", true},
		{"numbers", "田中123", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateJapaneseName("name", tt.value)
			if (err != nil) != tt.hasError {
				t.Errorf("ValidateJapaneseName() error = %v, hasError %v", err, tt.hasError)
			}
		})
	}
}

func TestValidator_ValidateKana(t *testing.T) {
	v := NewValidator()
	
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"katakana", "タナカ", false},
		{"katakana with space", "タナカ タロウ", false},
		{"empty", "", false}, // Empty is allowed
		{"hiragana", "たなか", true},
		{"kanji", "田中", true},
		{"mixed", "タナカ太郎", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateKana("kana", tt.value)
			if (err != nil) != tt.hasError {
				t.Errorf("ValidateKana() error = %v, hasError %v", err, tt.hasError)
			}
		})
	}
}

func TestValidator_ValidatePassword(t *testing.T) {
	v := NewValidator()
	
	tests := []struct {
		name     string
		password string
		hasError bool
	}{
		{"valid password", "Password123", false},
		{"empty password", "", true},
		{"too short", "Pass1", true},
		{"no uppercase", "password123", true},
		{"no lowercase", "PASSWORD123", true},
		{"no digits", "Password", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidatePassword("password", tt.password)
			if (err != nil) != tt.hasError {
				t.Errorf("ValidatePassword() error = %v, hasError %v", err, tt.hasError)
			}
		})
	}
}

func TestValidator_ValidateNotContainSQLKeywords(t *testing.T) {
	v := NewValidator()
	
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"safe text", "田中太郎", false},
		{"empty", "", false},
		{"sql injection attempt", "'; DROP TABLE users; --", true},
		{"select statement", "SELECT * FROM users", true},
		{"case insensitive", "select * from users", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateNotContainSQLKeywords("input", tt.value)
			if (err != nil) != tt.hasError {
				t.Errorf("ValidateNotContainSQLKeywords() error = %v, hasError %v", err, tt.hasError)
			}
		})
	}
}

func TestValidator_ValidateNotContainXSS(t *testing.T) {
	v := NewValidator()
	
	tests := []struct {
		name     string
		value    string
		hasError bool
	}{
		{"safe text", "田中太郎", false},
		{"empty", "", false},
		{"script tag", "<script>alert('xss')</script>", true},
		{"javascript url", "javascript:alert('xss')", true},
		{"onload event", "<img onload='alert(1)'>", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateNotContainXSS("input", tt.value)
			if (err != nil) != tt.hasError {
				t.Errorf("ValidateNotContainXSS() error = %v, hasError %v", err, tt.hasError)
			}
		})
	}
}

func TestFormValidator_ValidateLoginForm(t *testing.T) {
	fv := NewFormValidator()
	
	tests := []struct {
		name      string
		data      map[string]string
		errorCount int
	}{
		{
			"valid login",
			map[string]string{
				"username":   "admin",
				"password":   "MySecure123",
				"csrf_token": "",
			},
			0,
		},
		{
			"missing username",
			map[string]string{
				"username":   "",
				"password":   "MySecure123",
				"csrf_token": "valid-token",
			},
			1,
		},
		{
			"missing all fields",
			map[string]string{
				"username":   "",
				"password":   "",
				"csrf_token": "",
			},
			2, // Username required + Password required
		},
		{
			"sql injection attempt",
			map[string]string{
				"username":   "admin'; DROP TABLE users; --",
				"password":   "MySecure123",
				"csrf_token": "valid-token",
			},
			1, // SQL keywords in username
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := fv.ValidateLoginForm(tt.data)
			if len(errors) != tt.errorCount {
				t.Errorf("ValidateLoginForm() error count = %v, want %v", len(errors), tt.errorCount)
			}
		})
	}
}

func TestFormValidator_SanitizeInput(t *testing.T) {
	fv := NewFormValidator()
	
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal text", "田中太郎", "田中太郎"},
		{"text with spaces", "  田中太郎  ", "田中太郎"},
		{"text with null bytes", "田中\x00太郎", "田中太郎"},
		{"text with control chars", "田中\x01太郎", "田中太郎"},
		{"text with tabs and newlines", "田中\t太郎\n", "田中\t太郎"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fv.SanitizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeInput() = %q, want %q", result, tt.expected)
			}
		})
	}
}